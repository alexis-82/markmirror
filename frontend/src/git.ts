import {GitCommit, GitHasChanges, GitPush, GitStatus} from '../wailsjs/go/main/App';
import {git} from '../wailsjs/go/models';
import {EventsOn} from '../wailsjs/runtime/runtime';

/** Funzioni di main.ts usate da commit e push. */
export interface GitDeps {
    /** Esegue un comando file evitando sovrapposizioni. */
    exclusive(task: () => Promise<unknown>): Promise<void>;
    /** Percorso del documento corrente, "" se non salvato. */
    currentPath(): string;
    /** true se il documento ha modifiche non salvate. */
    isModified(): boolean;
    /** Salva sul file corrente; false se non salvato. */
    save(): Promise<boolean>;
    /** Mostra un messaggio con il solo pulsante OK. */
    showMessage(title: string, message: string): Promise<unknown>;
    /** Chiede conferma; true se l'utente sceglie il pulsante principale. */
    confirm(title: string, message: string, okLabel: string): Promise<boolean>;
}

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const TITLE = 'Commit e push';
const BUTTON_LABEL = 'Commit e push';

let deps: GitDeps;
let info = new git.Info({repo: '', branch: ''});
let request = 0;

export function initGit(d: GitDeps): void {
    deps = d;
    $('git-commit').addEventListener('click', () => deps.exclusive(commitAndPush));
    EventsOn('menu:git-commit', () => deps.exclusive(commitAndPush));
    // Ctrl+Invio conferma dal campo del messaggio.
    $('git-message').addEventListener('keydown', e => {
        if (e.key === 'Enter' && e.ctrlKey) {
            e.preventDefault();
            $<HTMLButtonElement>('git-confirm').click();
        }
    });
}

/** Aggiorna il repository del documento corrente e abilita o disabilita il pulsante. */
export async function refreshGit(): Promise<void> {
    const id = ++request;
    let next: git.Info;
    try {
        next = await GitStatus();
    } catch {
        next = new git.Info({repo: '', branch: ''});
    }
    if (id !== request) return; // nel frattempo è cambiato documento
    info = next;
    updateButton();
}

/** Il pulsante resta sempre visibile, abilitato solo dentro un repository. */
function updateButton(): void {
    const button = $<HTMLButtonElement>('git-commit');
    button.disabled = !info.repo;
    button.title = info.repo
        ? 'Commit del solo file aperto e push (Ctrl+Shift+G)'
        : 'Disponibile solo per file in un repository Git';
}

async function commitAndPush(): Promise<void> {
    if (!deps.currentPath()) return;
    await refreshGit(); // il repository può essere nato o sparito nel frattempo
    if (!info.repo) {
        await deps.showMessage(TITLE, 'Il documento non si trova in un repository Git.');
        return;
    }
    if (deps.isModified() && !(await deps.save())) return;

    if (!(await GitHasChanges())) {
        if (await deps.confirm(TITLE, 'Il file non ha modifiche da committare. Eseguire solo il push?', 'Push')) {
            await running('Push in corso…', GitPush);
            await deps.showMessage(TITLE, 'Push completato.');
        }
        return;
    }

    const message = await askMessage();
    if (message === null) return;
    await running('Commit in corso…', () => GitCommit(message));
    try {
        await running('Push in corso…', GitPush);
    } catch (err) {
        throw `Il commit è stato creato, ma il push non è riuscito:\n\n${err}`;
    }
    await deps.showMessage(TITLE, 'Commit e push completati.');
}

/** Chiede il messaggio di commit; null se annullato. */
function askMessage(): Promise<string | null> {
    const dialog = $<HTMLDialogElement>('git');
    const name = deps.currentPath().split(/[\\/]/).pop() ?? '';
    const branch = info.branch ? `sul branch ${info.branch}` : 'con HEAD staccato';
    $('git-target').textContent = `Verrà committato solo ${name} ${branch}, poi inviato al remote.`;
    const input = $<HTMLTextAreaElement>('git-message');
    input.value = `Aggiorna ${name}`;
    dialog.returnValue = '';
    dialog.showModal();
    input.select();
    return new Promise(resolve => {
        dialog.addEventListener('close', () => {
            const message = input.value.trim();
            resolve(dialog.returnValue === 'commit' && message ? message : null);
        }, {once: true});
    });
}

/** Esegue un comando git mostrando lo stato sul pulsante della toolbar. */
async function running(label: string, task: () => Promise<void>): Promise<void> {
    const button = $<HTMLButtonElement>('git-commit');
    button.textContent = label;
    button.disabled = true;
    try {
        await task();
    } finally {
        button.textContent = BUTTON_LABEL;
        updateButton();
    }
}
