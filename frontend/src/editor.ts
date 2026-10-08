import {Compartment, EditorSelection, EditorState, Prec} from '@codemirror/state';
import {EditorView, keymap, placeholder} from '@codemirror/view';
import {basicSetup} from 'codemirror';
import {indentWithTab, redo, selectAll, undo} from '@codemirror/commands';
import {markdown, markdownLanguage} from '@codemirror/lang-markdown';
import {languages} from '@codemirror/language-data';
import {oneDark} from '@codemirror/theme-one-dark';
import {ClipboardGetText, ClipboardSetText} from '../wailsjs/runtime/runtime';

export interface EditorHooks {
    onChange(): void;
    onScroll(): void;
    onCursor(): void;
}

const themeCompartment = new Compartment();
let darkMode = false;
let hooks: EditorHooks;

const baseTheme = EditorView.theme({
    '&': {height: '100%', fontSize: '15px'},
    '.cm-scroller': {fontFamily: 'var(--font-mono)', lineHeight: '1.6'},
    '.cm-content': {padding: '12px 0'},
    '&.cm-focused': {outline: 'none'},
});

function createState(doc: string): EditorState {
    return EditorState.create({
        doc,
        extensions: [
            basicSetup,
            markdown({base: markdownLanguage, codeLanguages: languages}),
            EditorView.lineWrapping,
            placeholder('Scrivi il tuo markdown qui...'),
            Prec.high(keymap.of([
                {key: 'Mod-b', run: v => wrap(v, '**', '**')},
                {key: 'Mod-i', run: v => wrap(v, '*', '*')},
                {key: 'Mod-k', run: v => insertLink(v)},
                indentWithTab,
            ])),
            baseTheme,
            themeCompartment.of(darkMode ? oneDark : []),
            EditorView.updateListener.of(update => {
                if (update.docChanged) hooks.onChange();
                if (update.docChanged || update.selectionSet) hooks.onCursor();
            }),
        ],
    });
}

export function createEditor(parent: HTMLElement, editorHooks: EditorHooks): EditorView {
    hooks = editorHooks;
    const view = new EditorView({parent, state: createState('')});
    view.scrollDOM.addEventListener('scroll', () => hooks.onScroll(), {passive: true});
    return view;
}

/** Sostituisce il documento azzerando la cronologia di annulla. */
export function resetDocument(view: EditorView, text: string): void {
    view.setState(createState(text));
    view.focus();
}

/** Sostituisce tutto il testo mantenendo cursore, scroll e cronologia di annulla. */
export function replaceDocument(view: EditorView, text: string): void {
    const {doc, selection} = view.state;
    if (doc.toString() === text) return;
    const head = Math.min(selection.main.head, text.length);
    view.dispatch({changes: {from: 0, to: doc.length, insert: text}, selection: {anchor: head}});
}

export function setEditorDark(view: EditorView, dark: boolean): void {
    darkMode = dark;
    view.dispatch({effects: themeCompartment.reconfigure(dark ? oneDark : [])});
}

/** Racchiude ogni selezione tra before/after, o li rimuove se già presenti. */
export function wrap(view: EditorView, before: string, after: string, fallback = 'testo'): boolean {
    const {state} = view;
    view.dispatch(state.changeByRange(range => {
        const pre = state.sliceDoc(range.from - before.length, range.from);
        const post = state.sliceDoc(range.to, range.to + after.length);
        if (!range.empty && pre === before && post === after) {
            return {
                changes: [
                    {from: range.from - before.length, to: range.from},
                    {from: range.to, to: range.to + after.length},
                ],
                range: EditorSelection.range(range.from - before.length, range.to - before.length),
            };
        }
        const text = state.sliceDoc(range.from, range.to) || fallback;
        const start = range.from + before.length;
        return {
            changes: {from: range.from, to: range.to, insert: before + text + after},
            range: EditorSelection.range(start, start + text.length),
        };
    }), {scrollIntoView: true});
    view.focus();
    return true;
}

/**
 * Applica un prefisso a tutte le righe selezionate. Se tutte lo hanno già,
 * lo rimuove. `prefix` riceve l'indice della riga (per gli elenchi numerati).
 */
function toggleLinePrefix(view: EditorView, prefix: (i: number) => string, matcher: RegExp): boolean {
    const {state} = view;
    const lines = new Set<number>();
    for (const r of state.selection.ranges) {
        const first = state.doc.lineAt(r.from).number;
        const last = state.doc.lineAt(r.to).number;
        for (let n = first; n <= last; n++) lines.add(n);
    }
    const sorted = [...lines].sort((a, b) => a - b).map(n => state.doc.line(n));
    const allPrefixed = sorted.every(l => matcher.test(l.text));
    const changes = sorted.map((line, i) => {
        if (allPrefixed) {
            const m = line.text.match(matcher)!;
            return {from: line.from, to: line.from + m[0].length};
        }
        return {from: line.from, insert: prefix(i)};
    });
    view.dispatch({changes, scrollIntoView: true});
    view.focus();
    return true;
}

/** Titolo ciclico sulla riga corrente: nessuno → # → ## → ### → nessuno. */
function cycleHeading(view: EditorView): boolean {
    const {state} = view;
    const line = state.doc.lineAt(state.selection.main.head);
    const m = line.text.match(/^(#{1,6})\s+/);
    const level = m ? m[1].length : 0;
    const next = level >= 3 ? '' : '#'.repeat(level + 1) + ' ';
    view.dispatch({changes: {from: line.from, to: line.from + (m ? m[0].length : 0), insert: next}});
    view.focus();
    return true;
}

function insertAndSelect(view: EditorView, text: string, selStart: number, selEnd: number): void {
    const {from, to} = view.state.selection.main;
    view.dispatch({
        changes: {from, to, insert: text},
        selection: EditorSelection.range(from + selStart, from + selEnd),
        scrollIntoView: true,
    });
    view.focus();
}

function insertLink(view: EditorView): boolean {
    const {from, to} = view.state.selection.main;
    const label = view.state.sliceDoc(from, to) || 'testo';
    const text = `[${label}](url)`;
    insertAndSelect(view, text, label.length + 3, label.length + 6);
    return true;
}

function insertCodeBlock(view: EditorView): void {
    const {from, to} = view.state.selection.main;
    const body = view.state.sliceDoc(from, to) || 'codice';
    const lineStart = view.state.doc.lineAt(from).from === from ? '' : '\n';
    const text = `${lineStart}\`\`\`\n${body}\n\`\`\`\n`;
    const start = lineStart.length + 4;
    insertAndSelect(view, text, start, start + body.length);
}

/** Inserisce un'immagine; i percorsi con spazi vanno racchiusi tra < >. */
export function insertImage(view: EditorView, ref: string, alt = 'image'): void {
    const target = /\s/.test(ref) ? `<${ref}>` : ref;
    const text = `![${alt}](${target})`;
    insertAndSelect(view, text, text.length, text.length);
}

function insertImageTemplate(view: EditorView): void {
    insertAndSelect(view, '![image](url)', 9, 12);
}

export type FormatCommand =
    'bold' | 'italic' | 'heading' | 'link' | 'code' | 'codeblock' |
    'list' | 'ordered' | 'quote' | 'image';

export function format(view: EditorView, cmd: FormatCommand): void {
    switch (cmd) {
        case 'bold': wrap(view, '**', '**'); break;
        case 'italic': wrap(view, '*', '*'); break;
        case 'code': wrap(view, '`', '`', 'codice'); break;
        case 'heading': cycleHeading(view); break;
        case 'link': insertLink(view); break;
        case 'codeblock': insertCodeBlock(view); break;
        case 'list': toggleLinePrefix(view, () => '- ', /^\s*[-*+]\s+/); break;
        case 'ordered': toggleLinePrefix(view, i => `${i + 1}. `, /^\s*\d+[.)]\s+/); break;
        case 'quote': toggleLinePrefix(view, () => '> ', /^\s*>\s?/); break;
        case 'image': insertImageTemplate(view); break;
    }
}

export type EditCommand = 'undo' | 'redo' | 'cut' | 'copy' | 'paste' | 'select-all';

/** Comandi del menu Modifica (gli appunti passano dal runtime di Wails). */
export async function edit(view: EditorView, cmd: EditCommand): Promise<void> {
    const {state} = view;
    const selected = () => state.selection.ranges.map(r => state.sliceDoc(r.from, r.to)).join('\n');
    switch (cmd) {
        case 'undo': undo(view); break;
        case 'redo': redo(view); break;
        case 'select-all': selectAll(view); break;
        case 'copy': await ClipboardSetText(selected()); break;
        case 'cut':
            await ClipboardSetText(selected());
            view.dispatch(view.state.replaceSelection(''));
            break;
        case 'paste': {
            const text = await ClipboardGetText();
            if (text) view.dispatch(view.state.replaceSelection(text.replace(/\r\n/g, '\n')));
            break;
        }
    }
    view.focus();
}

/** Scorre l'editor in modo che la riga (0-based, frazionaria) sia in cima. */
export function scrollEditorToLine(view: EditorView, line: number, atEnd: boolean): void {
    const scroller = view.scrollDOM;
    if (atEnd) {
        scroller.scrollTop = scroller.scrollHeight;
        return;
    }
    const {doc} = view.state;
    const n = Math.min(doc.lines, Math.max(1, Math.floor(line) + 1));
    const block = view.lineBlockAt(doc.line(n).from);
    const fraction = Math.min(1, Math.max(0, line - Math.floor(line)));
    // Distanza tra l'inizio dell'area scorrevole e l'inizio del documento (padding).
    const docOffset = view.documentTop - scroller.getBoundingClientRect().top + scroller.scrollTop;
    scroller.scrollTop = docOffset + block.top + block.height * fraction;
}

/** Riga (0-based) e frazione di riga in cima all'area visibile dell'editor. */
export function topVisibleLine(view: EditorView): {line: number, fraction: number, atEnd: boolean} {
    const scroller = view.scrollDOM;
    const height = scroller.getBoundingClientRect().top - view.documentTop;
    const block = view.lineBlockAtHeight(Math.max(0, height));
    const line = view.state.doc.lineAt(block.from).number - 1;
    const fraction = block.height > 0 ? Math.min(1, Math.max(0, (height - block.top) / block.height)) : 0;
    const atEnd = scroller.scrollTop + scroller.clientHeight >= scroller.scrollHeight - 2;
    return {line, fraction, atEnd};
}
