package main

import (
	"embed"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"markmirror-go/internal/settings"
)

//go:embed all:frontend/dist
var assets embed.FS

// version è impostata dalla pipeline di release con
// -ldflags "-X main.version=vX.Y.Z". "dev" indica una build di sviluppo,
// alla quale non vengono proposti aggiornamenti.
var version = "dev"

func main() {
	app := NewApp(initialFileFromArgs())

	err := wails.Run(&options.App{
		Title:            appName,
		Width:            1200,
		Height:           800,
		MinWidth:         settings.MinWidth,
		MinHeight:        settings.MinHeight,
		WindowStartState: app.windowStartState(),
		Menu:             app.buildMenu(),
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: app.doc.LocalFiles,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		OnStartup:     app.startup,
		OnBeforeClose: app.beforeClose,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			Theme: windows.SystemDefault,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// initialFileFromArgs restituisce il file passato da riga di comando
// (es. doppio clic su un .md associato all'applicazione).
func initialFileFromArgs() string {
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}
