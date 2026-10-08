import themeCss from './theme.css?raw';
import markdownCss from './markdown.css?raw';

// Impaginazione del file esportato: il contenuto usa gli stessi stili dell'anteprima.
const pageCss = `
body {
    margin: 0;
    padding: 32px 24px;
    background: var(--bg);
    color: var(--fg);
    font-family: var(--font-ui);
}

@media print {
    body {
        padding: 0;
    }
}
`;

/**
 * Crea un file HTML autonomo dall'anteprima: stessi stili (tema chiaro) e
 * immagini locali incorporate, così il file si può spostare o inviare.
 */
export async function buildHtml(preview: HTMLElement, title: string): Promise<string> {
    const article = document.createElement('article');
    article.className = 'markdown-body';
    article.append(...Array.from(preview.childNodes, n => n.cloneNode(true)));

    for (const el of article.querySelectorAll('[data-line]')) el.removeAttribute('data-line');
    await Promise.all(Array.from(article.querySelectorAll('img'), embedLocalImage));

    return `<!DOCTYPE html>
<html lang="it">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="generator" content="MarkMirror">
<title>${escapeHtml(title)}</title>
<style>
${themeCss}
${markdownCss}
${pageCss}
</style>
</head>
<body>
${article.outerHTML}
</body>
</html>
`;
}

/** Sostituisce un'immagine servita da /localfile con il suo contenuto in base64. */
async function embedLocalImage(img: HTMLImageElement): Promise<void> {
    const src = img.getAttribute('src');
    if (!src?.startsWith('/localfile')) return;
    try {
        const response = await fetch(src);
        if (!response.ok) return;
        img.setAttribute('src', await toDataUrl(await response.blob()));
    } catch {
        // Immagine non leggibile: resta il riferimento, come nell'anteprima.
    }
}

function toDataUrl(blob: Blob): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result as string);
        reader.onerror = () => reject(reader.error);
        reader.readAsDataURL(blob);
    });
}

function escapeHtml(text: string): string {
    return text.replace(/[&<>"]/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'})[c]!);
}
