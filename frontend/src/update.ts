import MarkdownIt from 'markdown-it';
import DOMPurify from 'dompurify';
import {CheckUpdate, GetVersion, InstallUpdate} from '../wailsjs/go/main/App';
import {main} from '../wailsjs/go/models';
import {BrowserOpenURL, EventsOn} from '../wailsjs/runtime/runtime';
import {browserSafe} from './preview';

/** Funzioni di main.ts usate dagli aggiornamenti. */
export interface UpdateDeps {
    /** Esegue un comando file evitando sovrapposizioni. */
    exclusive(task: () => Promise<unknown>): Promise<void>;
    /** Chiede se salvare le modifiche; false se l'utente annulla. */
    confirmDiscard(): Promise<boolean>;
    /** Mostra un messaggio con il solo pulsante OK. */
    showMessage(title: string, message: string): Promise<unknown>;
}

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;

// Note di rilascio: niente HTML, solo Markdown.
const md = new MarkdownIt({html: false, linkify: true});

let deps: UpdateDeps;
let offer: main.UpdateInfo | null = null;
let updating = false;

/** true dal clic su "Scarica e installa" finché l'installazione non fallisce. */
export function isUpdating(): boolean {
    return updating;
}

export function initUpdates(d: UpdateDeps): void {
    deps = d;
    const dialog = $<HTMLDialogElement>('update');
    // Durante l'installazione la finestra non si chiude (Esc).
    dialog.addEventListener('cancel', e => {
        if (updating) e.preventDefault();
    });
    $('update-later').addEventListener('click', () => dialog.close());
    $('update-install').addEventListener('click', () => deps.exclusive(install));
    // I link delle note si aprono nel browser di sistema, mai nell'app.
    $('update-notes').addEventListener('click', e => {
        const link = (e.target as Element).closest('a');
        if (!link) return;
        e.preventDefault();
        const href = link.getAttribute('href') ?? '';
        if (/^https?:/i.test(href)) BrowserOpenURL(browserSafe(href));
    });
    EventsOn('update:progress', (p: {done: number; total: number}) => {
        const percent = p.total > 0 ? Math.floor((p.done / p.total) * 100) : 0;
        $('update-bar').style.width = `${percent}%`;
        setStatus(percent < 100 ? `Download in corso… ${percent}%` : 'Avvio dell’installazione…');
    });
    EventsOn('menu:check-updates', () => deps.exclusive(checkFromMenu));
}

/** Controllo all'avvio: silenzioso se offline, senza release o già aggiornato. */
export async function checkUpdatesAtStartup(enabled: boolean): Promise<void> {
    if (!enabled || (await GetVersion()) === 'dev') return;
    let info: main.UpdateInfo;
    try {
        info = await CheckUpdate();
    } catch {
        return;
    }
    if (info.available) showWhenFree(info);
}

/** Mostra la finestra appena non c'è un altro dialogo aperto. */
function showWhenFree(info: main.UpdateInfo): void {
    const open = document.querySelector<HTMLDialogElement>('dialog[open]');
    if (open) open.addEventListener('close', () => showWhenFree(info), {once: true});
    else showUpdate(info);
}

async function checkFromMenu(): Promise<void> {
    const info = await CheckUpdate();
    if (info.noRelease) {
        await deps.showMessage('Verifica aggiornamenti', 'Nessuna release pubblicata.');
    } else if (info.dev) {
        await deps.showMessage('Verifica aggiornamenti',
            `Build di sviluppo: aggiornamenti non disponibili (ultima release: ${info.latest}).`);
    } else if (!info.available) {
        await deps.showMessage('Verifica aggiornamenti', `MarkMirror è aggiornato (${info.current}).`);
    } else {
        // Non si attende la chiusura: "Scarica e installa" usa a sua volta exclusive.
        showUpdate(info);
    }
}

function showUpdate(info: main.UpdateInfo): void {
    const dialog = $<HTMLDialogElement>('update');
    if (dialog.open) return;
    offer = info;
    $('update-versions').textContent = `È disponibile MarkMirror ${info.latest} (installata: ${info.current}).`;
    const notes = info.notes.trim()
        ? DOMPurify.sanitize(md.render(info.notes), {RETURN_DOM_FRAGMENT: true})
        : document.createDocumentFragment();
    $('update-notes').replaceChildren(notes);
    $('update-no-asset').hidden = info.canInstall;
    $('update-install').hidden = !info.canInstall;
    setInstalling(false);
    setError('');
    dialog.showModal();
    (info.canInstall ? $('update-install') : $('update-later')).focus();
}

async function install(): Promise<void> {
    if (!offer?.canInstall || updating) return;
    if (!(await deps.confirmDiscard())) return;
    updating = true;
    setError('');
    setInstalling(true);
    try {
        await InstallUpdate(); // in caso di successo l'app si chiude
    } catch (err) {
        updating = false;
        setInstalling(false);
        setError(String(err));
    }
}

function setInstalling(on: boolean): void {
    $<HTMLButtonElement>('update-install').disabled = on;
    $<HTMLButtonElement>('update-later').disabled = on;
    $('update-progress').hidden = !on;
    $('update-bar').style.width = '0';
    setStatus(on ? 'Download in corso… 0%' : '');
}

function setStatus(text: string): void {
    $('update-status').textContent = text;
    $('update-status').hidden = !text;
}

function setError(text: string): void {
    $('update-error').textContent = text;
    $('update-error').hidden = !text;
}
