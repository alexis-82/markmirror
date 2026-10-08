# Implementazioni pianificate

Elenco delle funzionalità da aggiungere a MarkMirror, in ordine di implementazione consigliato:
prima le più semplici e indipendenti, per ultima l'esportazione, che deve già includere
formule e diagrammi.

| # | Funzionalità | Complessità | Tocca |
|---|---|---|---|
| 1 | Ricordare la finestra | Bassa | Go |
| 2 | Inserisci tabella | Bassa | Frontend |
| 3 | Incolla immagini dagli appunti | Media | Go + frontend |
| 4 | Formule KaTeX e diagrammi Mermaid | Media | Frontend |
| 5 | Esporta in HTML / PDF | Media | Go + frontend |

---

## 1. Ricordare la finestra ✅

**Obiettivo:** alla riapertura, la finestra ha la stessa dimensione, posizione e stato
(ingrandita o no) dell'ultima chiusura.

### Task
- [x] Aggiungere a `Settings` (`settings.go`) il campo `Window` (`X`, `Y`, `Width`, `Height`, `Maximised`).
- [x] Alla chiusura salvare lo stato in `beforeClose` (Wails la chiama anche da `Quit`).
      Si usa `GetWindowPlacement` di Win32 invece delle funzioni `runtime.Window*`: restituisce
      le dimensioni "normali" anche se la finestra è ingrandita o ridotta a icona, e da ridotta
      a icona ricorda se prima era ingrandita.
- [x] In `main.go` `WindowStartState: options.Maximised` se l'ultima volta era ingrandita.
- [x] In `startup` ripristinare posizione e dimensioni con `SetWindowPlacement`, prima che Wails mostri la finestra.
- [x] **Controllo monitor:** se il centro della barra del titolo non cade su nessuno schermo
      (monitor scollegato), non si ripristina nulla e la finestra resta centrata.
- [x] Rispettare i minimi esistenti (`MinWidth` 700, `MinHeight` 400), ora costanti in `settings.go`.
- [x] Test: lettura di un `settings.json` vecchio, valori non validi, andata e ritorno JSON,
      dimensione della struttura `WINDOWPLACEMENT`.

File: `window.go`, `window_windows.go`, `settings.go`, `app.go`, `main.go`.

---

## 2. Inserisci tabella dalla barra strumenti

**Obiettivo:** un pulsante "Tabella" apre una piccola griglia (stile Word): si scelgono
righe × colonne passando il mouse, al clic viene inserita la tabella markdown.

### Task
- [ ] Aggiungere il pulsante in `frontend/index.html` (`data-cmd="table"`).
- [ ] Creare il popup a griglia (es. massimo 8 × 8) con l'etichetta "3 × 4" in tempo reale;
      chiusura con Esc o clic fuori. Stili in `style.css` per tema chiaro e scuro.
- [ ] In `editor.ts` aggiungere `insertTable(view, rows, cols)`: genera intestazione,
      riga separatrice `| --- |` e righe vuote; lascia una riga vuota prima/dopo se il cursore
      è in mezzo al testo; seleziona la prima cella d'intestazione.
- [ ] Gestire il clic in `main.ts` (come `image-file`, il comando non passa da `format`).

### Facoltativo
- Tab / Shift+Tab per spostarsi tra le celle quando il cursore è dentro una tabella.

---

## 3. Incolla immagini dagli appunti

**Obiettivo:** Ctrl+V con un'immagine negli appunti (es. uno screenshot) salva il file
accanto al documento e inserisce il link markdown.

### Task
- [ ] In `editor.ts` intercettare l'evento `paste` (`EditorView.domEventHandlers`): se
      `clipboardData.items` contiene un `image/*`, bloccare l'incolla standard e passare l'immagine a `main.ts`.
      Il testo continua a incollarsi normalmente.
- [ ] Anche il comando di menu **Modifica → Incolla** (`edit(view, 'paste')`, che usa `ClipboardGetText`)
      deve gestire le immagini: valutare la lettura dagli appunti di Windows lato Go (formato `CF_DIB` / PNG).
- [ ] Nuovo metodo Go `SaveClipboardImage(dataBase64 string, ext string) (string, error)`:
  - salva in `img/` accanto al documento (crea la cartella se manca);
  - nome univoco, es. `img/incollata-2026-10-08-153012.png`, senza sovrascrivere file esistenti;
  - restituisce il percorso relativo tramite `ImageLink`.
- [ ] **Documento non ancora salvato:** chiedere di salvarlo prima (senza una cartella non si sa
      dove mettere l'immagine) e annullare se l'utente rifiuta.
- [ ] Inserire con `insertImage(view, ref, alt)`.
- [ ] Test Go: salvataggio, creazione della cartella, nomi univoci, documento non salvato → errore.

---

## 4. Formule matematiche (KaTeX) e diagrammi Mermaid

**Obiettivo:** l'anteprima mostra `$...$` / `$$...$$` come formule e i blocchi
```` ```mermaid ```` come diagrammi.

### KaTeX
- [ ] Dipendenze: `katex` + un plugin markdown-it (es. `@vscode/markdown-it-katex` o `markdown-it-texmath`).
- [ ] Registrare il plugin in `preview.ts` (`md.use(...)`) e importare `katex/dist/katex.min.css`
      (i font KaTeX vengono inclusi da Vite).
- [ ] **DOMPurify:** l'output KaTeX contiene MathML e molti `span` con stile: verificare che il
      sanitize non lo rovini (eventualmente `USE_PROFILES: {html: true, mathMl: true}`).
- [ ] Le formule con errori devono mostrare il messaggio in rosso invece di rompere l'anteprima (`throwOnError: false`).
- [ ] Il `$` usato normalmente (es. "costa 5$") non deve diventare una formula: controllare il comportamento del plugin scelto.

### Mermaid
- [ ] Dipendenza `mermaid`, caricata con `import()` dinamico solo se il documento contiene
      un blocco mermaid (è pesante, non deve rallentare l'avvio).
- [ ] In `preview.ts` i blocchi ```` ```mermaid ```` vengono resi come contenitori con il sorgente,
      poi `mermaid.render()` produce l'SVG dopo l'aggiornamento.
- [ ] **Aggiornamento incrementale:** ridisegnare solo i diagrammi il cui sorgente è cambiato,
      per non perdere il vantaggio dell'aggiornamento a blocchi (niente sfarfallio a ogni tasto).
- [ ] Tema Mermaid `dark` / `default` in base al tema dell'app; ridisegnare al cambio tema.
- [ ] Errori di sintassi: mostrare un messaggio nel blocco, non un'eccezione.
- [ ] Sicurezza: `securityLevel: 'strict'`.
- [ ] Aggiungere esempi alla guida Markdown (`content.ts`).

### Attenzione
- Lo scroll sincronizzato usa `data-line`: verificare che formule e diagrammi mantengano
  l'attributo sul blocco contenitore.

---

## 5. Esporta in HTML / PDF ✅

**Obiettivo:** File → Esporta → HTML… / PDF… produce un documento uguale all'anteprima.

### HTML
- [x] Voce di menu `menu.go`: sottomenu **File → Esporta** con "HTML..." e "PDF / Stampa..." (Ctrl+P).
- [x] Il frontend (`export.ts`) genera l'HTML completo: titolo = nome del file, stessi CSS dell'anteprima.
      Il CSS è stato diviso in `theme.css` (colori), `markdown.css` (contenuto + highlight.js) e
      `style.css` (interfaccia); l'export incorpora i primi due con l'import `?raw` di Vite.
- [x] **Immagini locali:** incorporate in base64 (file unico, si può spostare o inviare).
- [x] Tema dell'export: sempre chiaro. Il tema scuro in `theme.css` vale solo `@media screen`.
- [x] Metodo Go `ExportHTML(html string) (string, error)` (`export.go`) con finestra "Salva"
      (filtro `*.html`, nome predefinito = nome del documento) e scrittura atomica.
- [ ] Da fare con il punto 4: aggiungere il CSS di KaTeX all'export. Mermaid: l'export copia il DOM
      dell'anteprima, quindi gli SVG già disegnati sono inclusi automaticamente.

### PDF
- [x] CSS `@media print` (nasconde barra strumenti, editor, divisore e barra di stato, mostra solo
      l'anteprima a tutta pagina, sempre in tema chiaro) + `window.print()`: dal dialogo di stampa si sceglie
      "Salva come PDF". Il nome proposto per il PDF è il nome del documento.
- [ ] Non fatto: PDF diretto senza dialogo (`PrintToPdf` di WebView2 non è esposto da Wails v2).
- [x] Interruzioni di pagina: niente tagli dentro codice, citazioni, immagini e righe di tabella;
      il codice lungo va a capo invece di essere tagliato; i titoli non restano soli in fondo alla pagina.

---

## Note generali

- Ogni funzionalità va chiusa con `go test ./...`, `npx tsc --noEmit` e una prova con `wails dev`
  in tema chiaro e scuro.
- Aggiornare `README.md` (elenco funzionalità e tabella della struttura) a ogni funzionalità completata.
- Un commit per funzionalità, così ognuna si può annullare singolarmente.
