export const WELCOME = `# Benvenuto in MarkMirror

## Funzionalità

- Modifica del testo markdown con evidenziazione della sintassi
- Anteprima in tempo reale
- Interfaccia divisa e ridimensionabile
- Sincronizzazione dello scroll

## Come utilizzarlo

1. Scrivi il markdown nel pannello di sinistra
2. Visualizza l'anteprima nel pannello di destra
3. Utilizza la barra degli strumenti per le operazioni comuni
4. Usa il menu File per salvare o caricare documenti
5. Trascina un file \`.md\` o un'immagine nella finestra

\`\`\`go
func main() {
    fmt.Println("Buona scrittura!")
}
\`\`\`

**Buona scrittura!**
`;

export const GUIDE = `## Testo

| Scrivi | Ottieni |
|---|---|
| \`**grassetto**\` | **grassetto** |
| \`*corsivo*\` | *corsivo* |
| \`~~barrato~~\` | ~~barrato~~ |
| \`\`\` \`codice\` \`\`\` | \`codice\` |

## Titoli

\`\`\`markdown
# Titolo 1
## Titolo 2
### Titolo 3
\`\`\`

## Elenchi

\`\`\`markdown
- elemento
- elemento
  - sotto-elemento

1. primo
2. secondo
\`\`\`

## Link e immagini

\`\`\`markdown
[testo del link](https://esempio.it)
![descrizione](immagini/foto.png)
![con spazi](<cartella con spazi/foto.png>)
\`\`\`

I percorsi relativi delle immagini partono dalla cartella del documento.

## Citazioni e separatori

\`\`\`markdown
> citazione

---
\`\`\`

## Blocchi di codice

Indica il linguaggio dopo i tre backtick per la colorazione:

\`\`\`\`markdown
\`\`\`python
print("ciao")
\`\`\`
\`\`\`\`

## Tabelle

\`\`\`markdown
| Colonna A | Colonna B |
|-----------|----------:|
| sinistra  |    destra |
\`\`\`

## Scorciatoie

| Tasti | Azione |
|---|---|
| Ctrl+N / Ctrl+O | Nuovo / Apri |
| Ctrl+S / Ctrl+Maiusc+S | Salva / Salva come |
| Ctrl+B / Ctrl+I / Ctrl+K | Grassetto / Corsivo / Link |
| Ctrl+F | Cerca e sostituisci |
| Ctrl+Z / Ctrl+Y | Annulla / Ripeti |
`;
