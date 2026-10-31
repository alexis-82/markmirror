# MarkMirror

Editor Markdown per Windows e Linux con anteprima in tempo reale: scrivi a sinistra, vedi il documento formattato a destra.

## Funzionalità

- Editor con evidenziazione Markdown, numeri di riga e cerca/sostituisci (Ctrl+F)
- Anteprima in tempo reale, senza salti di scroll e senza ricaricare le immagini a ogni tasto
- Colorazione dei blocchi di codice in base al linguaggio indicato
- Scroll sincronizzato in entrambe le direzioni tra editor e anteprima
- Immagini locali con percorsi relativi alla cartella del documento
- Link aperti nel browser di sistema
- Tema chiaro, scuro o di sistema, ricordato tra una sessione e l'altra (in tema scuro anche menu e barra dei menu)
- Posizione, dimensione e stato ingrandito della finestra ricordati alla riapertura
- Esportazione in HTML (un unico file con stili e immagini incorporati) e in PDF / stampa (Ctrl+P)
- Barra degli strumenti: grassetto, corsivo, titolo, elenchi, citazione, link, codice, immagini
- Trascinamento nella finestra di file `.md` (per aprirli) e di immagini (per inserirle)
- Conferma delle modifiche non salvate su Nuovo / Apri / Ricarica / chiusura
- File → Ricarica (F5): rilegge il file modificato da un altro programma
- Ricarica automatica quando il file cambia sul disco; con modifiche non salvate viene chiesta conferma
- Conserva le terminazioni di riga del file (Windows o Linux)
- Apertura diretta di un file `.md` con doppio clic o con "Apri con → MarkMirror"
- Percorso completo del file nel titolo della finestra, per distinguere file con lo stesso nome in cartelle diverse
- Aggiornamenti dall'app: Aiuto → Verifica aggiornamenti, più un controllo automatico all'avvio
- Commit e push del file aperto (File → Commit e push, Ctrl+Shift+G, o pulsante nella barra degli strumenti, sempre visibile e abilitato solo) quando il file si trova in un repository Git; richiede Git installato

## Download

Scarica l'ultima versione dalla pagina [Releases](https://github.com/alexis-82/markmirror/releases). Nei nomi dei file, `<ver>` è il numero di versione, es. `2.1.0`.

| Sistema | File |
|---|---|
| Windows (con installazione) | `MarkMirror-<ver>-amd64-installer.exe` |
| Windows (portable, senza installazione) | `MarkMirror-<ver>-amd64-portable.exe` |
| Debian / Ubuntu e derivate | `markmirror_<ver>_amd64.deb` |
| Fedora | `markmirror-<ver>-1.x86_64.rpm` |
| Arch Linux e derivate (Manjaro, EndeavourOS, CachyOS) | `markmirror-<ver>-1-x86_64.pkg.tar.zst` |
| Altre distribuzioni Linux | `MarkMirror-<ver>-linux-amd64.tar.gz` |

Il file `SHA256SUMS.txt` contiene i checksum per verificare i download.

## Installazione

### Windows

Avvia `MarkMirror-<ver>-amd64-installer.exe`:

- scegli la lingua (italiano o inglese);
- scegli se creare i collegamenti sul desktop e nel menu Start.

MarkMirror viene installato solo per il tuo utente e non servono i diritti di amministratore. Se è presente la vecchia versione 2.0.0, installata per tutti gli utenti, il setup propone di rimuoverla (in quel caso Windows chiede l'autorizzazione).

La versione portable non va installata: basta avviare il file.

### Debian / Ubuntu

```sh
sudo apt install ./markmirror_<ver>_amd64.deb
```

### Fedora

```sh
sudo dnf install ./markmirror-<ver>-1.x86_64.rpm
```

### Arch Linux

```sh
sudo pacman -U ./markmirror-<ver>-1-x86_64.pkg.tar.zst
```

### Altre distribuzioni Linux

Estrai l'archivio ed entra nella cartella estratta, poi:

```sh
sudo sh build/linux/install.sh
```

Serve la libreria WebKitGTK 4.1 (`libwebkit2gtk-4.1-0` o equivalente).

Dopo l'installazione trovi MarkMirror nel menu delle applicazioni, e puoi aprire un file `.md` con "Apri con → MarkMirror".

## Aggiornamenti

- **Aiuto → Verifica aggiornamenti** controlla se c'è una versione più recente. Se c'è, mostra le novità e la installa con "Scarica e installa". Prima di installare, il file scaricato viene verificato. Se il documento ha modifiche non salvate, MarkMirror chiede prima se salvarle.
- All'avvio il controllo avviene da solo e non mostra nulla se non c'è niente di nuovo. Si disattiva da **Opzioni → Controlla aggiornamenti all'avvio**.
- **Windows**: parte il setup della nuova versione, anche se usi la versione portable.
- **Linux**: viene chiesta la password di amministratore, poi la nuova versione viene installata nello stesso modo dell'attuale (deb, rpm, pacchetto Arch o archivio) e MarkMirror si riavvia. Da terminale si può aggiornare anche con:
  ```sh
  sudo markmirror-update.sh
  ```

La versione 2.0.0 non ha questa funzione: per passare a una versione successiva installala a mano una volta. Da lì in poi gli aggiornamenti arrivano dall'app.

## Disinstallazione

- **Windows**: Impostazioni → App installate → MarkMirror → Disinstalla. Puoi scegliere se eliminare anche le impostazioni; i tuoi documenti non vengono mai toccati.
- **Debian / Ubuntu**: `sudo apt remove markmirror`
- **Fedora**: `sudo dnf remove markmirror`
- **Arch Linux**: `sudo pacman -R markmirror`
- **Altre distribuzioni Linux**: dalla cartella dell'archivio estratto, `sudo sh build/linux/install.sh --uninstall`

Licenza MIT — © Alessio Abrugiati
