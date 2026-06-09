package main

import (
	"embed"

	"github.com/mimo-cli/mimo-cli/desktop"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:desktop/frontend/dist
var assets embed.FS

func main() {
	app, err := desktop.NewApp()
	if err != nil {
		panic("Failed to initialize app: " + err.Error())
	}

	err = wails.Run(&options.App{
		Title:    "MiMo Desktop",
		Width:    1200,
		Height:   800,
		MinWidth: 800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 18, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
		},
		Frameless:        true,
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.Dark,
		},
	})
	if err != nil {
		panic("Wails error: " + err.Error())
	}
}
