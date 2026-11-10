<div align="center">

# MarkMirror

**A Markdown editor for Windows and Linux with a live preview: write on the left, see the formatted document on the right.**<br>
*Editor Markdown per Windows e Linux con anteprima in tempo reale: scrivi a sinistra, vedi il documento formattato a destra.*

[![Version](https://img.shields.io/badge/version-2.6.1-blue?style=plastic)](CHANGELOG.md)&nbsp;&nbsp;&nbsp;[![Go](https://img.shields.io/badge/Go-00ADD8?style=plastic&logo=go&logoColor=white)](#)&nbsp;&nbsp;&nbsp;[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux-lightgrey?style=plastic)](#)

🇬🇧 [English](#en)&nbsp;&nbsp;·&nbsp;&nbsp;🇮🇹 [Italiano](#it)

</div>

---

<a id="en"></a>
<a id="contents"></a>
## 📑 Contents

- [✨ What it is](#what-it-is)
- [🚀 Installation](#installation)
- [🧭 How to use](#how-to-use)
- [🛟 Troubleshooting](#troubleshooting)
- [📄 License](#license)

---

<a id="what-it-is"></a>
## ✨ What it is

MarkMirror is an editor for Markdown documents (`.md` files): you write plain text on the left and see the formatted document on the right while you type.

- Text coloring for Markdown, line numbers, and find and replace (Ctrl+F)
- Live preview that does not jump around while you type
- Code blocks colored according to their language
- Editor and preview scroll together, in both directions
- Images from your computer, placed next to the document
- Links open in your web browser
- Light, dark or system theme, remembered between sessions
- Full screen from the View menu ("Visualizza") or with F11
- Window position, size and maximized state remembered when you reopen the app
- Export to HTML (a single file with styles and images included) and to PDF or print (Ctrl+P)
- Toolbar for bold, italic, headings, lists, quotes, links, code, images and emoji
- About 300 of the most used emoji, grouped into categories (😀 button in the toolbar)
- Drag a `.md` file into the window to open it, or an image to insert it
- Asks before discarding unsaved changes
- Reload (F5) and automatic reload when another program changes the file
- Keeps the line endings of the file (Windows or Linux)
- Open a `.md` file with a double click or with "Open with → MarkMirror"
- Full file path in the window title, to tell apart files with the same name
- Updates from inside the app
- Commit and push of the open file when it is inside a Git repository (requires Git)

<a id="installation"></a>
## 🚀 Installation

Download the latest version from the [Releases](https://github.com/alexis-82/markmirror/releases) page. In the file names, `<ver>` is the version number, for example `2.6.0`.

| System | File |
|---|---|
| Windows (with setup) | `MarkMirror-<ver>-amd64-installer.exe` |
| Windows (portable, no setup) | `MarkMirror-<ver>-amd64-portable.exe` |
| Debian / Ubuntu and derivatives | `markmirror_<ver>_amd64.deb` |
| Fedora | `markmirror-<ver>-1.x86_64.rpm` |
| Arch Linux and derivatives (Manjaro, EndeavourOS, CachyOS) | `markmirror-<ver>-1-x86_64.pkg.tar.zst` |

**Windows**: open the setup file, choose the language and whether you want shortcuts on the desktop and in the Start menu. MarkMirror is installed for your user only and does not need administrator rights. The portable version needs no setup: just open the file.

**Linux**: open the Terminal in the folder where you downloaded the file and paste the command for your system.

```sh
# Debian / Ubuntu
sudo apt install ./markmirror_<ver>_amd64.deb

# Fedora
sudo dnf install ./markmirror-<ver>-1.x86_64.rpm

# Arch Linux
sudo pacman -U ./markmirror-<ver>-1-x86_64.pkg.tar.zst
```

After installing, you find MarkMirror in the applications menu.

**Uninstall**: on Windows go to Settings → Installed apps → MarkMirror → Uninstall; your documents are never touched. On Linux paste the command for your system in the Terminal.

```sh
# Debian / Ubuntu
sudo apt remove markmirror

# Fedora
sudo dnf remove markmirror

# Arch Linux
sudo pacman -R markmirror
```

<a id="how-to-use"></a>
## 🧭 How to use

- **Write**: type on the left, the preview on the right updates by itself. Use the toolbar to format text without remembering the Markdown syntax.
- **Full screen**: View ("Visualizza") → "Schermo intero", or press F11. Press F11 again to go back to the window as it was.
- **Export**: File → Export ("Esporta"), as HTML or as PDF / print.
- **Updates**: Help ("Aiuto") → "Verifica aggiornamenti" checks for a new version and installs it. The app also checks by itself at startup; you can turn this off in Options ("Opzioni") → "Controlla aggiornamenti all'avvio". On Linux you are asked for the administrator password, then MarkMirror restarts with the new version.

<a id="troubleshooting"></a>
## 🛟 Troubleshooting

**Emoji appear as empty squares on Linux.** Install the Noto Color Emoji font: open the Terminal and paste the command for your system.

```sh
# Debian / Ubuntu
sudo apt install fonts-noto-color-emoji

# Fedora
sudo dnf install google-noto-color-emoji-fonts

# Arch Linux
sudo pacman -S noto-fonts-emoji
```

**Flags appear as pairs of letters on Windows** (for example IT): the Windows font does not draw them, it is not an error.

**F11 does nothing on Linux.** Some desktops use F11 for themselves. Use View ("Visualizza") → "Schermo intero" instead.

<a id="license"></a>
## 📄 License

MIT — © Alessio Abrugiati

---

<a id="it"></a>
<a id="indice"></a>
## 📑 Indice

- [✨ Cos'è](#cose)
- [🚀 Installazione](#installazione)
- [🧭 Come si usa](#come-si-usa)
- [🛟 Problemi comuni](#problemi-comuni)
- [📄 Licenza](#licenza)

---

<a id="cose"></a>
## ✨ Cos'è

MarkMirror è un editor per documenti Markdown (file `.md`): scrivi testo semplice a sinistra e vedi il documento formattato a destra mentre scrivi.

- Colorazione del testo Markdown, numeri di riga, cerca e sostituisci (Ctrl+F)
- Anteprima in tempo reale che non salta mentre scrivi
- Blocchi di codice colorati in base al linguaggio
- Editor e anteprima scorrono insieme, in entrambe le direzioni
- Immagini dal tuo computer, messe accanto al documento
- Link aperti nel browser
- Tema chiaro, scuro o di sistema, ricordato tra una sessione e l'altra
- Schermo intero dal menu Visualizza o con F11
- Posizione, dimensione e stato ingrandito della finestra ricordati alla riapertura
- Esportazione in HTML (un unico file con stili e immagini incorporati) e in PDF o stampa (Ctrl+P)
- Barra degli strumenti per grassetto, corsivo, titoli, elenchi, citazioni, link, codice, immagini ed emoji
- Circa 300 emoji tra le più usate, divise in categorie (pulsante 😀 nella barra degli strumenti)
- Trascina nella finestra un file `.md` per aprirlo, o un'immagine per inserirla
- Chiede conferma prima di perdere modifiche non salvate
- Ricarica (F5) e ricarica automatica quando un altro programma modifica il file
- Conserva le terminazioni di riga del file (Windows o Linux)
- Apri un file `.md` con doppio clic o con "Apri con → MarkMirror"
- Percorso completo del file nel titolo della finestra, per distinguere file con lo stesso nome
- Aggiornamenti dall'app
- Commit e push del file aperto quando si trova in un repository Git (serve Git installato)

<a id="installazione"></a>
## 🚀 Installazione

Scarica l'ultima versione dalla pagina [Releases](https://github.com/alexis-82/markmirror/releases). Nei nomi dei file, `<ver>` è il numero di versione, per esempio `2.6.0`.

| Sistema | File |
|---|---|
| Windows (con installazione) | `MarkMirror-<ver>-amd64-installer.exe` |
| Windows (portable, senza installazione) | `MarkMirror-<ver>-amd64-portable.exe` |
| Debian / Ubuntu e derivate | `markmirror_<ver>_amd64.deb` |
| Fedora | `markmirror-<ver>-1.x86_64.rpm` |
| Arch Linux e derivate (Manjaro, EndeavourOS, CachyOS) | `markmirror-<ver>-1-x86_64.pkg.tar.zst` |

**Windows**: apri il file di installazione, scegli la lingua e se vuoi i collegamenti sul desktop e nel menu Start. MarkMirror viene installato solo per il tuo utente e non servono i diritti di amministratore. La versione portable non va installata: basta aprire il file.

**Linux**: apri il Terminale nella cartella dove hai scaricato il file e incolla il comando per il tuo sistema.

```sh
# Debian / Ubuntu
sudo apt install ./markmirror_<ver>_amd64.deb

# Fedora
sudo dnf install ./markmirror-<ver>-1.x86_64.rpm

# Arch Linux
sudo pacman -U ./markmirror-<ver>-1-x86_64.pkg.tar.zst
```

Dopo l'installazione trovi MarkMirror nel menu delle applicazioni.

**Disinstallazione**: su Windows vai in Impostazioni → App installate → MarkMirror → Disinstalla; i tuoi documenti non vengono mai toccati. Su Linux incolla nel Terminale il comando per il tuo sistema.

```sh
# Debian / Ubuntu
sudo apt remove markmirror

# Fedora
sudo dnf remove markmirror

# Arch Linux
sudo pacman -R markmirror
```

<a id="come-si-usa"></a>
## 🧭 Come si usa

- **Scrivere**: scrivi a sinistra, l'anteprima a destra si aggiorna da sola. Usa la barra degli strumenti per formattare il testo senza ricordare la sintassi Markdown.
- **Schermo intero**: Visualizza → Schermo intero, oppure premi F11. Premi di nuovo F11 per tornare alla finestra com'era.
- **Esportare**: File → Esporta, in HTML o in PDF / stampa.
- **Aggiornamenti**: Aiuto → Verifica aggiornamenti controlla se c'è una versione nuova e la installa. L'app controlla anche da sola all'avvio; puoi disattivarlo da Opzioni → Controlla aggiornamenti all'avvio. Su Linux viene chiesta la password di amministratore, poi MarkMirror si riavvia con la nuova versione.

<a id="problemi-comuni"></a>
## 🛟 Problemi comuni

**Su Linux le emoji appaiono come quadratini.** Installa il font Noto Color Emoji: apri il Terminale e incolla il comando per il tuo sistema.

```sh
# Debian / Ubuntu
sudo apt install fonts-noto-color-emoji

# Fedora
sudo dnf install google-noto-color-emoji-fonts

# Arch Linux
sudo pacman -S noto-fonts-emoji
```

**Su Windows le bandiere appaiono come coppie di lettere** (per esempio IT): il font di Windows non le disegna, non è un errore.

**Su Linux F11 non fa nulla.** Alcuni desktop usano F11 per sé. Usa invece Visualizza → Schermo intero.

<a id="licenza"></a>
## 📄 Licenza

MIT — © Alessio Abrugiati
