import {EMOJI_CATEGORIES} from './emoji-data';

export interface EmojiPickerHooks {
    /** L'utente ha scelto un'emoji: il pannello è già chiuso. */
    onPick(emoji: string): void;
    /** Pannello chiuso con Esc, con il pulsante o ridimensionando la finestra, senza scegliere. */
    onDismiss(): void;
}

const MARGIN = 8;

function panelButton(className: string, text: string): HTMLButtonElement {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = className;
    btn.textContent = text;
    return btn;
}

/** Pannello a comparsa ancorato al pulsante della barra degli strumenti. */
export function initEmojiPicker(button: HTMLElement, panel: HTMLElement, hooks: EmojiPickerHooks): void {
    const tabs = document.createElement('div');
    tabs.className = 'emoji-tabs';
    tabs.setAttribute('role', 'tablist');
    const tabButtons = EMOJI_CATEGORIES.map(category => {
        const tab = panelButton('emoji-tab', category.icon);
        tab.title = category.name;
        tab.setAttribute('aria-label', category.name);
        tab.setAttribute('role', 'tab');
        return tab;
    });
    tabs.replaceChildren(...tabButtons);

    const grid = document.createElement('div');
    grid.className = 'emoji-grid';
    grid.setAttribute('role', 'tabpanel');
    panel.replaceChildren(tabs, grid);

    function selectCategory(index: number): void {
        tabButtons.forEach((tab, i) => tab.setAttribute('aria-selected', String(i === index)));
        grid.replaceChildren(...EMOJI_CATEGORIES[index].emojis.map(emoji => panelButton('emoji-cell', emoji)));
        grid.scrollTop = 0;
    }

    tabs.addEventListener('click', e => {
        const tab = (e.target as Element).closest<HTMLButtonElement>('.emoji-tab');
        if (tab) selectCategory(tabButtons.indexOf(tab));
    });

    grid.addEventListener('click', e => {
        const cell = (e.target as Element).closest<HTMLButtonElement>('.emoji-cell');
        if (!cell) return;
        close();
        hooks.onPick(cell.textContent!);
    });

    function open(): void {
        selectCategory(0);
        panel.hidden = false;
        button.setAttribute('aria-expanded', 'true');
        // Dimensioni note solo a pannello visibile: resta dentro la finestra.
        const rect = button.getBoundingClientRect();
        const top = rect.bottom + 4;
        panel.style.top = `${top}px`;
        panel.style.left = `${Math.max(MARGIN, Math.min(rect.left, window.innerWidth - panel.offsetWidth - MARGIN))}px`;
        panel.style.maxHeight = `${window.innerHeight - top - MARGIN}px`;
        tabButtons[0].focus();
    }

    function close(): void {
        panel.hidden = true;
        button.setAttribute('aria-expanded', 'false');
    }

    button.addEventListener('click', () => {
        if (panel.hidden) {
            open();
        } else {
            close();
            hooks.onDismiss();
        }
    });

    // In fase di cattura: nessun altro elemento può nascondere il clic al pannello.
    document.addEventListener('pointerdown', e => {
        const target = e.target as Node;
        if (!panel.hidden && !panel.contains(target) && !button.contains(target)) close();
    }, true);

    document.addEventListener('keydown', e => {
        if (e.key === 'Escape' && !panel.hidden) {
            close();
            hooks.onDismiss();
        }
    });

    window.addEventListener('resize', () => {
        if (panel.hidden) return;
        close();
        hooks.onDismiss();
    });
}
