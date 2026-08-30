package main

import (
	"embed"
	"log"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/application"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

var version = "dev"

func main() {
	app := application.New(version)
	if err := wails.Run(&options.App{
		Title:                    "EasyInput Flasher",
		Width:                    1320,
		Height:                   860,
		MinWidth:                 1024,
		MinHeight:                680,
		BackgroundColour:         &options.RGBA{R: 11, G: 13, B: 18, A: 1},
		AssetServer:              &assetserver.Options{Assets: assets},
		OnStartup:                app.Startup,
		OnShutdown:               app.Shutdown,
		EnableDefaultContextMenu: true,
		Bind:                     []interface{}{app},
		Mac: &mac.Options{About: &mac.AboutInfo{
			Title:   "EasyInput Flasher",
			Message: "GitHub Release firmware installer for EasyInput V2.0.",
			Icon:    appIcon,
		}},
	}); err != nil {
		log.Fatal(err)
	}
}
