import MarkdownIt from 'markdown-it';
import DOMPurify from 'dompurify';
import hljs from 'highlight.js/lib/common';
import {BrowserOpenURL} from '../wailsjs/runtime/runtime';

const md = new MarkdownIt({
    html: true,
    linkify: true,
    breaks: true, // a capo singolo = <br>, come nella versione Python
    highlight(code, lang) {
        if (lang && hljs.getLanguage(lang)) {
            try {
                return hljs.highlight(code, {language: lang, ignoreIllegals: true}).value;
            } catch { /* ricade sull'escape standard */ }
        }
        return '';
    },
});

// Annota ogni blocco con la riga di origine, usata dallo scroll sincronizzato.
md.core.ruler.push('source_lines', state => {
    for (const token of state.tokens) {
        if (token.map && token.nesting !== -1) {
            token.attrSet('data-line', String(token.map[0]));
        }
    }
});

/**
 * Identificativo di un titolo come su GitHub: minuscole, spazi → "-",
 * punteggiatura tolta. "## Perché usarlo?" → "perché-usarlo".
 */
export function slugify(text: string): string {
    return text.trim().toLowerCase().replace(/[^\p{L}\p{N}\s_-]/gu, '').replace(/\s/g, '-');
}

// Dà un id a ogni titolo, così i link dell'indice (#titolo) funzionano.
// I doppioni ricevono un suffisso: "note", "note-1", "note-2".
md.core.ruler.push('heading_ids', state => {
    const used = new Map<string, number>();
    state.tokens.forEach((token, i) => {
        if (token.type !== 'heading_open') return;
        const text = (state.tokens[i + 1].children ?? [])
            .filter(t => t.type === 'text' || t.type === 'code_inline')
            .map(t => t.content)
            .join('');
        const base = slugify(text);
        const count = used.get(base) ?? 0;
        used.set(base, count + 1);
        token.attrSet('id', count ? `${base}-${count}` : base);
    });
});

// DOMPurify antepone questo prefisso agli id (SANITIZE_NAMED_PROPS), così un
// titolo come "Title" non può sovrascrivere proprietà di document.
export const ID_PREFIX = 'user-content-';

const remoteUrl = /^(?:[a-z][a-z0-9+.-]*:|\/\/|#)/i;
const windowsPath = /^[a-z]:[\\/]/i;

function isLocal(src: string): boolean {
    return windowsPath.test(src) || !remoteUrl.test(src);
}

function safeDecode(src: string): string {
    try {
        return decodeURI(src);
    } catch {
        return src;
    }
}

/** Converte un riferimento locale nell'URL servito dal backend Go. */
function localUrl(src: string, base: string): string {
    return `/localfile?path=${encodeURIComponent(safeDecode(src))}&base=${encodeURIComponent(base)}`;
}

function render(source: string, base: string): DocumentFragment {
    const frag = DOMPurify.sanitize(md.render(source), {RETURN_DOM_FRAGMENT: true, SANITIZE_NAMED_PROPS: true});
    for (const img of frag.querySelectorAll('img')) {
        const src = img.getAttribute('src');
        if (src && isLocal(src)) img.setAttribute('src', localUrl(src, base));
    }
    return frag;
}

// Chiave di confronto dei blocchi: l'HTML senza i numeri di riga, così un
// blocco solo spostato non viene ricreato (e le immagini non ricaricate).
const keys = new WeakMap<Node, string>();
const lineAttr = / data-line="\d+"/g;

function keyOf(node: Node): string {
    let key = keys.get(node);
    if (key === undefined) {
        key = node instanceof Element ? node.outerHTML.replace(lineAttr, '') : `#${node.nodeType}:${node.textContent}`;
        keys.set(node, key);
    }
    return key;
}

/** Copia i numeri di riga da un blocco nuovo a quello esistente equivalente. */
function syncLines(target: Node, source: Node): void {
    if (!(target instanceof Element) || !(source instanceof Element)) return;
    const t = [target, ...target.querySelectorAll('[data-line]')];
    const s = [source, ...source.querySelectorAll('[data-line]')];
    s.forEach((el, i) => {
        const line = el.getAttribute('data-line');
        if (line !== null && t[i]) t[i].setAttribute('data-line', line);
    });
}

// Spazio lasciato sopra il blocco allineato in cima all'anteprima.
const SCROLL_MARGIN = 8;

export class Preview {
    private anchors: {line: number, el: HTMLElement}[] | null = null;

    constructor(private readonly root: HTMLElement) {
        root.addEventListener('click', e => this.onClick(e));
    }

    /** Aggiorna l'anteprima toccando solo i blocchi cambiati. */
    update(source: string, base: string): void {
        const next = Array.from(render(source, base).childNodes);
        const prev = Array.from(this.root.childNodes);

        let start = 0;
        while (start < prev.length && start < next.length && keyOf(prev[start]) === keyOf(next[start])) {
            syncLines(prev[start], next[start]);
            start++;
        }
        let endPrev = prev.length;
        let endNext = next.length;
        while (endPrev > start && endNext > start && keyOf(prev[endPrev - 1]) === keyOf(next[endNext - 1])) {
            syncLines(prev[endPrev - 1], next[endNext - 1]);
            endPrev--;
            endNext--;
        }

        const anchor = prev[endPrev] ?? null;
        for (let i = start; i < endPrev; i++) prev[i].remove();
        for (let i = start; i < endNext; i++) this.root.insertBefore(next[i], anchor);
        this.anchors = null;
    }

    /** Porta in cima all'anteprima il punto corrispondente alla riga indicata. */
    scrollToLine(line: number, fraction: number, atEnd: boolean): void {
        const root = this.root;
        if (atEnd) {
            root.scrollTop = root.scrollHeight;
            return;
        }
        const anchors = this.getAnchors();
        if (anchors.length === 0) return;

        let i = 0;
        while (i + 1 < anchors.length && anchors[i + 1].line <= line) i++;
        const cur = anchors[i];
        const next = anchors[i + 1];
        const curTop = this.offsetOf(cur.el);
        let target: number;
        if (line < cur.line) {
            target = 0;
        } else if (next) {
            const progress = (line + fraction - cur.line) / (next.line - cur.line);
            target = curTop + (this.offsetOf(next.el) - curTop) * progress;
        } else {
            target = curTop + cur.el.offsetHeight * fraction;
        }
        root.scrollTop = target - SCROLL_MARGIN;
    }

    /**
     * Inverso di scrollToLine: la riga di origine (con parte frazionaria)
     * corrispondente al punto in cima all'anteprima.
     */
    lineAtTop(): {line: number, atEnd: boolean} {
        const root = this.root;
        const atEnd = root.scrollTop + root.clientHeight >= root.scrollHeight - 2;
        const anchors = this.getAnchors();
        if (anchors.length === 0) return {line: 0, atEnd};

        const y = root.scrollTop + SCROLL_MARGIN;
        let i = -1;
        while (i + 1 < anchors.length && this.offsetOf(anchors[i + 1].el) <= y) i++;
        if (i < 0) return {line: 0, atEnd};

        const cur = anchors[i];
        const next = anchors[i + 1];
        const curTop = this.offsetOf(cur.el);
        if (next) {
            const nextTop = this.offsetOf(next.el);
            const progress = nextTop > curTop ? (y - curTop) / (nextTop - curTop) : 0;
            return {line: cur.line + (next.line - cur.line) * progress, atEnd};
        }
        const height = cur.el.offsetHeight;
        return {line: cur.line + (height > 0 ? Math.min(1, (y - curTop) / height) : 0), atEnd};
    }

    onScroll(handler: () => void): void {
        this.root.addEventListener('scroll', handler, {passive: true});
    }

    private getAnchors() {
        if (!this.anchors) {
            const seen = new Set<number>();
            this.anchors = [];
            for (const el of this.root.querySelectorAll<HTMLElement>('[data-line]')) {
                const line = Number(el.dataset.line);
                if (seen.has(line)) continue; // tiene l'elemento più esterno
                seen.add(line);
                this.anchors.push({line, el});
            }
            this.anchors.sort((a, b) => a.line - b.line);
        }
        return this.anchors;
    }

    private offsetOf(el: HTMLElement): number {
        return el.getBoundingClientRect().top - this.root.getBoundingClientRect().top + this.root.scrollTop;
    }

    private onClick(e: MouseEvent): void {
        const link = (e.target as Element).closest('a');
        if (!link) return;
        e.preventDefault();
        const href = link.getAttribute('href') ?? '';
        if (href.startsWith('#')) this.scrollToAnchor(href.slice(1));
        else if (/^(https?:|mailto:)/i.test(href)) BrowserOpenURL(browserSafe(href));
        else if (/^www\./i.test(href)) BrowserOpenURL(browserSafe(`https://${href}`));
    }

    /** Porta in cima all'anteprima il titolo indicato da un link "#titolo". */
    private scrollToAnchor(fragment: string): void {
        let name = fragment;
        try {
            name = decodeURIComponent(fragment);
        } catch { /* lascia il frammento com'è */ }
        const target = this.root.querySelector<HTMLElement>(`#${CSS.escape(ID_PREFIX + slugify(name))}`);
        if (target) this.root.scrollTop = this.offsetOf(target) - SCROLL_MARGIN;
    }
}

/**
 * Wails rifiuta in silenzio gli URL con caratteri come ( ) ~ ! (es. molte
 * pagine di Wikipedia): li codifichiamo, il browser li interpreta allo stesso modo.
 */
function browserSafe(url: string): string {
    return url.replace(/[;|`$\\<>*{}[\]()~! ]/g, c => '%' + c.charCodeAt(0).toString(16).toUpperCase().padStart(2, '0'));
}

export function renderStatic(target: HTMLElement, source: string): void {
    target.replaceChildren(render(source, ''));
}
