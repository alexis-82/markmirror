# MarkMirror

Editor Markdown desktop con anteprima in tempo reale, scritto in Go + [Wails v2](https://wails.io) (frontend TypeScript + Vite).

## Funzionalità

- Editor CodeMirror 6 con evidenziazione Markdown, numeri di riga, cerca/sostituisci (Ctrl+F)
- Anteprima in tempo reale (markdown-it) aggiornata in modo incrementale: niente salti di scroll, immagini non ricaricate a ogni tasto
- Colorazione dei blocchi di codice (highlight.js) in base al linguaggio indicato
- Scroll sincronizzato in entrambe le direzioni (editor ↔ anteprima), basato sulle righe di origine
- Immagini locali con percorsi relativi alla cartella del documento
- HTML dell'anteprima sanificato con DOMPurify; i link si aprono nel browser di sistema
- Tema chiaro / scuro / di sistema, salvato tra una sessione e l'altra; in tema scuro anche la barra dei menu e i menu a tendina sono scuri
- Posizione, dimensione e stato ingrandito della finestra ricordati alla riapertura
- Esportazione in HTML (file unico con stili e immagini incorporate) e in PDF / stampa (Ctrl+P)
- Barra strumenti: grassetto, corsivo, titolo, elenchi, citazione, link, codice, immagini
- Trascinamento di file `.md` (apertura) e immagini (inserimento) nella finestra
- Conferma delle modifiche non salvate su Nuovo / Apri / Ricarica / chiusura
- File → Ricarica (F5): rilegge dal disco il file modificato da un altro programma
- Ricarica automatica: se il file cambia su disco viene ricaricato da solo; con modifiche non salvate viene chiesta conferma
- Conserva le terminazioni di riga (CRLF/LF) del file; salvataggio atomico
- Apertura di un file passato da riga di comando (`MarkMirror.exe documento.md`)

## Installazione su Linux

1. Installa (sempre dalla cartella principale):
   ```sh
   sudo sh build/linux/install.sh
   ```
2. Avvia MarkMirror dal menu delle applicazioni, oppure apri un file `.md` con "Apri con → MarkMirror".

Per disinstallare:
  ```sh
  sudo sh build/linux/install.sh --uninstall
  ```

Licenza MIT — © Alessio Abrugiati
