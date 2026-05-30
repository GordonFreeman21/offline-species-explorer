package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	if err := InitDB(); err != nil {
		log.Fatalf("Database init failed: %s", err.Error())
	}
	defer CloseDB()

	if err := wails.Run(&options.App{
		Title:  "Offline Species Explorer",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			&App{},
		},
	}); err != nil {
		log.Fatal(err.Error())
	}
}
