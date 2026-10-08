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

## 1. Ricordare la finestra

**Obiettivo:** alla riapertura, la finestra ha la stessa dimensione, posizione e stato
(ingrandita o no) dell'ultima chiusura.

### Task
- [ ] Aggiungere a `Settings` (`settings.go`) i campi della finestra: `X`, `Y`, `Width`, `Height`, `Maximised`.
- [ ] Alla chiusura (`beforeClose` in `app.go`, e in `Quit`) leggere lo stato con
      `runtime.WindowGetSize`, `runtime.WindowGetPosition`, `runtime.WindowIsMaximised` e salvarlo.
      Se la finestra è ingrandita, salvare solo `Maximised = true` e mantenere le ultime dimensioni "normali".
      Se è ridotta a icona, non salvare nulla.
- [ ] In `main.go` usare le dimensioni salvate per `Width`/`Height` e `WindowStartState: options.Maximised` se serve.
- [ ] In `startup` ripristinare la posizione con `runtime.WindowSetPosition`.
- [ ] **Controllo monitor:** se la posizione salvata è fuori da tutti gli schermi (monitor scollegato),
      centrare la finestra (`runtime.ScreenGetAll` + `runtime.WindowCenter`).
- [ ] Rispettare i minimi esistenti (`MinWidth` 700, `MinHeight` 400).

### Attenzione
- `beforeClose` può bloccare la chiusura (modifiche non salvate): lo stato va salvato anche
  nel percorso `Quit()`, che chiude senza ripassare dal blocco.
- Test: serializzazione/lettura dei nuovi campi e valori di default per un `settings.json` vecchio.

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

## 5. Esporta in HTML / PDF

**Obiettivo:** File → Esporta → HTML… / PDF… produce un documento uguale all'anteprima.

### HTML
- [ ] Voce di menu `menu.go`: sottomenu **Esporta** con "HTML..." e "PDF...".
- [ ] Il frontend genera l'HTML completo: `<!doctype html>`, titolo = nome del file, CSS dell'anteprima
      (solo la parte tipografica, non quella dell'interfaccia), CSS di highlight.js e KaTeX.
- [ ] **Immagini locali:** nell'anteprima puntano a `/localfile?...`, che fuori dall'app non esiste.
      Opzioni: (a) incorporarle in base64 (file unico, consigliato), (b) riscrivere i percorsi relativi originali.
- [ ] **Mermaid:** esportare gli SVG già disegnati, non il sorgente.
- [ ] Tema dell'export: sempre chiaro (adatto a stampa e condivisione), o scelta in un'opzione.
- [ ] Nuovo metodo Go `ExportHTML(html string) error` con finestra "Salva" (filtro `*.html`,
      nome predefinito = nome del documento) e scrittura atomica (`writeFileAtomic`).

### PDF
- [ ] Strada più semplice: CSS `@media print` (nasconde barra strumenti, editor e barra di stato,
      mostra solo l'anteprima a tutta pagina) + `window.print()`, poi l'utente sceglie "Microsoft Print to PDF".
- [ ] Strada migliore (da verificare): WebView2 offre `PrintToPdf`, ma Wails v2 non lo espone;
      richiederebbe accesso diretto all'oggetto WebView2. Valutare costi/benefici prima di procedere.
- [ ] Interruzioni di pagina: evitare tagli dentro blocchi di codice, tabelle e immagini (`break-inside: avoid`).

---

## Note generali

- Ogni funzionalità va chiusa con `go test ./...`, `npx tsc --noEmit` e una prova con `wails dev`
  in tema chiaro e scuro.
- Aggiornare `README.md` (elenco funzionalità e tabella della struttura) a ogni funzionalità completata.
- Un commit per funzionalità, così ognuna si può annullare singolarmente.
