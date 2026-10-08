package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	donateURL  = "https://pal.me/serse82"
	websiteURL = "https://www.alexis82.it"
	authorMail = "alessioabrugiati@gmail.com"
)

var themeChoices = []struct{ id, label string }{
	{"system", "Sistema"},
	{"light", "Chiaro"},
	{"dark", "Scuro"},
}

// buildMenu crea il menu dell'applicazione. Le azioni sul documento sono
// inoltrate al frontend come eventi "menu:*".
func (a *App) buildMenu() *menu.Menu {
	m := menu.NewMenu()

	file := m.AddSubmenu("File")
	file.AddText("Nuovo", keys.CmdOrCtrl("n"), a.emitter("menu:new"))
	file.AddText("Apri...", keys.CmdOrCtrl("o"), a.emitter("menu:open"))
	file.AddText("Ricarica", keys.Key("f5"), a.emitter("menu:reload"))
	file.AddSeparator()
	file.AddText("Salva", keys.CmdOrCtrl("s"), a.emitter("menu:save"))
	file.AddText("Salva come...", keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey), a.emitter("menu:save-as"))
	file.AddSeparator()
	file.AddText("Esci", keys.CmdOrCtrl("q"), func(*menu.CallbackData) { runtime.Quit(a.ctx) })

	// Senza scorciatoie: i tasti (Ctrl+Z, Ctrl+C, ...) li gestisce già l'editor.
	edit := m.AddSubmenu("Modifica")
	edit.AddText("Annulla", nil, a.emitter("menu:undo"))
	edit.AddText("Ripeti", nil, a.emitter("menu:redo"))
	edit.AddSeparator()
	edit.AddText("Taglia", nil, a.emitter("menu:cut"))
	edit.AddText("Copia", nil, a.emitter("menu:copy"))
	edit.AddText("Incolla", nil, a.emitter("menu:paste"))
	edit.AddSeparator()
	edit.AddText("Seleziona tutto", nil, a.emitter("menu:select-all"))

	options := m.AddSubmenu("Opzioni")
	theme := options.AddSubmenu("Tema")
	current := a.GetSettings()
	var themeItems []*menu.MenuItem
	for _, choice := range themeChoices {
		id := choice.id
		item := theme.AddRadio(choice.label, current.Theme == id, nil, func(cd *menu.CallbackData) {
			// Wails alterna lo stato della voce cliccata: lo riallineiamo noi.
			for _, it := range themeItems {
				it.Checked = it == cd.MenuItem
			}
			runtime.MenuUpdateApplicationMenu(a.ctx)
			a.updateSettings(func(s *Settings) { s.Theme = id })
		})
		themeItems = append(themeItems, item)
	}
	options.AddCheckbox("Scroll sincronizzato", current.SyncScroll, nil, func(cd *menu.CallbackData) {
		checked := cd.MenuItem.Checked
		a.updateSettings(func(s *Settings) { s.SyncScroll = checked })
	})

	help := m.AddSubmenu("Aiuto")
	help.AddText("Guida Markdown", nil, a.emitter("menu:guide"))
	help.AddSeparator()
	help.AddText("Donazione", nil, func(*menu.CallbackData) { a.showDonate() })
	help.AddText("Informazioni", nil, func(*menu.CallbackData) { a.showAbout() })

	return m
}

func (a *App) emitter(event string) menu.Callback {
	return func(*menu.CallbackData) {
		runtime.EventsEmit(a.ctx, event)
	}
}

func (a *App) showDonate() {
	answer, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:  runtime.QuestionDialog,
		Title: "Donazione",
		Message: "Grazie per supportare lo sviluppo di " + appName + "!\n\n" +
			"Puoi fare una donazione tramite PayPal (" + donateURL + ")\n" +
			"oppure contattarmi via email: " + authorMail + "\n\n" +
			"Vuoi aprire la pagina PayPal?",
	})
	if err == nil && answer == "Yes" {
		runtime.BrowserOpenURL(a.ctx, donateURL)
	}
}

func (a *App) showAbout() {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:  runtime.InfoDialog,
		Title: "Informazioni",
		Message: fmt.Sprintf("%s v%s\n\n"+
			"Editor Markdown con anteprima in tempo reale.\n\n"+
			"Sviluppato da Alessio Abrugiati\n%s", appName, appVersion, websiteURL),
	})
}
