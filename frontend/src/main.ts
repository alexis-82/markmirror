import './style.css';

import {createEditor, edit, EditCommand, format, FormatCommand, insertImage, replaceDocument, resetDocument, scrollEditorToLine, setEditorDark, topVisibleLine} from './editor';
import {Preview, renderStatic} from './preview';
import {GUIDE, WELCOME} from './content';
import {
    GetSettings, ImageLink, InitialFile, NewFile, OpenFile, OpenPath, PickImage, Quit, ReloadFile, SaveFile, SaveFileAs, SetModified,
} from '../wailsjs/go/main/App';
import {main} from '../wailsjs/go/models';
import {EventsOn, OnFileDrop} from '../wailsjs/runtime/runtime';

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;

// ---- Stato del documento ----------------------------------------------------

let currentPath = '';
let modified = false;
let settings: main.Settings = {theme: 'system', syncScroll: true} as main.Settings;

function setModified(value: boolean): void {
    if (modified === value) return;
    modified = value;
    SetModified(value);
}

// ---- Editor e anteprima -----------------------------------------------------

const preview = new Preview($('preview'));
let renderTimer = 0;

const view = createEditor($('editor'), {
    onChange() {
        setModified(true);
        scheduleRender();
    },
    onScroll: syncScroll,
    onCursor: updateStats,
});

function scheduleRender(): void {
    clearTimeout(renderTimer);
    renderTimer = window.setTimeout(renderNow, 150);
}

function renderNow(): void {
    clearTimeout(renderTimer);
    preview.update(view.state.doc.toString(), currentPath);
    syncScroll();
    updateStats();
}

// Pannello che sta guidando lo scroll. Gli eventi di scroll generati sull'altro
// pannello dalla sincronizzazione vengono ignorati, così i due non si rincorrono.
type ScrollSource = 'editor' | 'preview';
let scrollLeader: ScrollSource | null = null;
let leaderTimer = 0;

function takeScrollLead(source: ScrollSource): boolean {
    if (scrollLeader && scrollLeader !== source) return false;
    scrollLeader = source;
    clearTimeout(leaderTimer);
    leaderTimer = window.setTimeout(() => scrollLeader = null, 120);
    return true;
}

/** Allinea l'anteprima all'editor. */
function syncScroll(): void {
    if (!settings.syncScroll || !takeScrollLead('editor')) return;
    const {line, fraction, atEnd} = topVisibleLine(view);
    preview.scrollToLine(line, fraction, atEnd);
}

/** Allinea l'editor all'anteprima. */
function syncScrollFromPreview(): void {
    if (!settings.syncScroll || !takeScrollLead('preview')) return;
    const {line, atEnd} = preview.lineAtTop();
    scrollEditorToLine(view, line, atEnd);
}

preview.onScroll(syncScrollFromPreview);

function updateStats(): void {
    const {state} = view;
    const head = state.selection.main.head;
    const line = state.doc.lineAt(head);
    const words = state.doc.toString().match(/\S+/g)?.length ?? 0;
    $('status-stats').textContent = `Riga ${line.number}, Col ${head - line.from + 1}  ·  ${words} parole`;
    $('status-file').textContent = currentPath || 'Documento non salvato';
}

function loadDocument(content: string, path: string): void {
    currentPath = path;
    resetDocument(view, content);
    modified = false;
    preview.update('', ''); // svuota: i percorsi relativi cambiano base
    renderNow();
    $('preview').scrollTop = 0;
}

// ---- Dialoghi ---------------------------------------------------------------

interface DialogButton {
    id: string;
    label: string;
    primary?: boolean;
}

function showDialog(title: string, message: string, buttons: DialogButton[]): Promise<string> {
    const dialog = $<HTMLDialogElement>('dialog');
    $('dialog-title').textContent = title;
    $('dialog-message').textContent = message;
    $('dialog-buttons').replaceChildren(...buttons.map(b => {
        const btn = document.createElement('button');
        btn.value = b.id;
        btn.textContent = b.label;
        if (b.primary) btn.className = 'primary';
        return btn;
    }));
    dialog.returnValue = '';
    dialog.showModal();
    (dialog.querySelector('button.primary') as HTMLButtonElement | null)?.focus();
    return new Promise(resolve => {
        dialog.addEventListener('close', () => resolve(dialog.returnValue || 'cancel'), {once: true});
    });
}

function showError(err: unknown): void {
    showDialog('Errore', String(err), [{id: 'ok', label: 'OK', primary: true}]);
}

// ---- Comandi file -----------------------------------------------------------

let busy = false;

/** Esegue un comando file evitando sovrapposizioni (es. doppio Ctrl+S). */
async function exclusive(task: () => Promise<unknown>): Promise<void> {
    if (busy) return;
    busy = true;
    try {
        await task();
    } catch (err) {
        showError(err);
    } finally {
        busy = false;
        view.focus();
        handleExternalChange();
    }
}

/** Chiede se salvare le modifiche. Restituisce false se l'utente annulla. */
async function confirmDiscard(): Promise<boolean> {
    if (!modified) return true;
    const choice = await showDialog(
        'Modifiche non salvate',
        'Il documento è stato modificato. Vuoi salvare le modifiche?',
        [
            {id: 'save', label: 'Salva', primary: true},
            {id: 'discard', label: 'Non salvare'},
            {id: 'cancel', label: 'Annulla'},
        ],
    );
    if (choice === 'save') return save(false);
    return choice === 'discard';
}

async function save(as: boolean): Promise<boolean> {
    const content = view.state.doc.toString();
    const result = await (as ? SaveFileAs(content) : SaveFile(content));
    if (!result) return false;
    const pathChanged = result.path !== currentPath;
    currentPath = result.path;
    modified = false;
    // Se nel frattempo il testo è cambiato, il documento resta modificato.
    setModified(view.state.doc.toString() !== content);
    if (pathChanged) renderNow();
    else updateStats();
    return true;
}

async function openPath(path: string): Promise<void> {
    if (!(await confirmDiscard())) return;
    const result = await OpenPath(path);
    loadDocument(result.content, result.path);
}

/**
 * Rilegge il file dal disco (es. modificato da un altro programma). Se ci sono
 * modifiche non salvate chiede conferma mostrando il messaggio indicato.
 */
async function reload(message: string, cancelLabel: string): Promise<void> {
    if (!currentPath) return;
    if (modified) {
        const choice = await showDialog('Ricarica', message, [
            {id: 'reload', label: 'Ricarica', primary: true},
            {id: 'cancel', label: cancelLabel},
        ]);
        if (choice !== 'reload') return;
    }
    const result = await ReloadFile();
    if (!result) return;
    replaceDocument(view, result.content);
    setModified(false);
    renderNow();
}

// Modifica esterna segnalata dal backend e non ancora gestita: se arriva
// mentre è in corso un altro comando, viene gestita appena questo termina.
let externalChange = '';

function handleExternalChange(): void {
    if (!externalChange || busy) return;
    const path = externalChange;
    externalChange = '';
    if (path !== currentPath) return;
    exclusive(async () => {
        try {
            await reload(
                'Il file è stato modificato da un altro programma. Ricaricarlo? Le modifiche non salvate andranno perse.',
                'Mantieni le mie modifiche',
            );
        } catch {
            // File momentaneamente illeggibile (es. ancora in scrittura): il
            // backend lo segnalerà di nuovo alla prossima modifica.
        }
    });
}

EventsOn('file:changed', (path: string) => {
    externalChange = path;
    handleExternalChange();
});

const fileCommands: Record<string, () => Promise<unknown>> = {
    'menu:new': async () => {
        if (!(await confirmDiscard())) return;
        await NewFile();
        loadDocument('', '');
    },
    'menu:open': async () => {
        if (!(await confirmDiscard())) return;
        const result = await OpenFile();
        if (result) loadDocument(result.content, result.path);
    },
    'menu:reload': () => reload(
        'Il documento ha modifiche non salvate che andranno perse. Ricaricare il file dal disco?',
        'Annulla',
    ),
    'menu:save': () => save(false),
    'menu:save-as': () => save(true),
    'app:close-requested': async () => {
        if (await confirmDiscard()) await Quit();
    },
};

for (const [event, task] of Object.entries(fileCommands)) {
    EventsOn(event, () => exclusive(task));
}

for (const cmd of ['undo', 'redo', 'cut', 'copy', 'paste', 'select-all'] as EditCommand[]) {
    EventsOn(`menu:${cmd}`, () => edit(view, cmd));
}

EventsOn('menu:guide', () => $<HTMLDialogElement>('guide').showModal());

// ---- Barra degli strumenti --------------------------------------------------

async function insertImageFromFile(): Promise<void> {
    const ref = await PickImage();
    if (ref) insertImage(view, ref, altFromPath(ref));
}

function altFromPath(path: string): string {
    const name = path.split(/[\\/]/).pop() ?? 'image';
    return name.replace(/\.[^.]+$/, '');
}

document.querySelector('.toolbar')!.addEventListener('click', e => {
    const btn = (e.target as Element).closest<HTMLButtonElement>('button[data-cmd]');
    if (!btn) return;
    const cmd = btn.dataset.cmd!;
    if (cmd === 'image-file') exclusive(insertImageFromFile);
    else format(view, cmd as FormatCommand);
});

// ---- Trascinamento file -----------------------------------------------------

const imagePattern = /\.(png|jpe?g|gif|bmp|webp|svg|avif|ico)$/i;
const documentPattern = /\.(md|markdown|txt)$/i;

OnFileDrop((_x, _y, paths) => {
    const doc = paths.find(p => documentPattern.test(p));
    if (doc) {
        exclusive(() => openPath(doc));
        return;
    }
    exclusive(async () => {
        for (const p of paths.filter(p => imagePattern.test(p))) {
            insertImage(view, await ImageLink(p), altFromPath(p));
        }
    });
}, false);

// ---- Divisore ridimensionabile ----------------------------------------------

const SPLIT_KEY = 'markmirror.split';

function setSplit(percent: number): void {
    const clamped = Math.min(80, Math.max(20, percent));
    $('panes').style.setProperty('--split', `${clamped}%`);
    try {
        localStorage.setItem(SPLIT_KEY, String(clamped));
    } catch { /* non essenziale */ }
}

function initDivider(): void {
    const divider = $('divider');
    const panes = $('panes');
    try {
        const saved = Number(localStorage.getItem(SPLIT_KEY));
        if (saved) setSplit(saved);
    } catch { /* non essenziale */ }

    divider.addEventListener('pointerdown', e => {
        divider.setPointerCapture(e.pointerId);
        panes.classList.add('resizing');
    });
    divider.addEventListener('pointermove', e => {
        if (!divider.hasPointerCapture(e.pointerId)) return;
        const rect = panes.getBoundingClientRect();
        setSplit(((e.clientX - rect.left) / rect.width) * 100);
    });
    divider.addEventListener('pointerup', e => {
        divider.releasePointerCapture(e.pointerId);
        panes.classList.remove('resizing');
        syncScroll();
    });
    divider.addEventListener('dblclick', () => setSplit(50));
}

// ---- Tema e impostazioni ----------------------------------------------------

const systemDark = window.matchMedia('(prefers-color-scheme: dark)');

function applySettings(next: main.Settings): void {
    settings = next;
    const dark = next.theme === 'dark' || (next.theme === 'system' && systemDark.matches);
    document.documentElement.dataset.theme = dark ? 'dark' : 'light';
    setEditorDark(view, dark);
    syncScroll();
}

systemDark.addEventListener('change', () => applySettings(settings));
EventsOn('settings:changed', (s: main.Settings) => applySettings(s));

// ---- Avvio ------------------------------------------------------------------

async function start(): Promise<void> {
    initDivider();
    renderStatic($('guide-body'), GUIDE);
    applySettings(await GetSettings());
    try {
        const initial = await InitialFile();
        if (initial) {
            loadDocument(initial.content, initial.path);
            return;
        }
    } catch (err) {
        showError(err);
    }
    loadDocument(WELCOME, '');
}

start();
