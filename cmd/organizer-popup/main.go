package main

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed assets
var assets embed.FS

func main() {
	file := flag.String("file", "", "organized file path")
	category := flag.String("category", "", "category")
	duration := flag.Int("duration", 15, "timeout in seconds")
	event := flag.String("event", "", "event identifier")
	callback := flag.String("callback", "", "callback address")
	token := flag.String("token", "", "callback key")
	choicesRaw := flag.String("choices", "[]", "destination selections")
	position := flag.String("position", "bottom_right", "popup position")
	dragEnabled := flag.Bool("drag-enabled", true, "drag permission")
	openFile := flag.Bool("open-file", true, "open file action")
	openLocation := flag.Bool("open-location", true, "open folder action")
	confirm := flag.Bool("confirm", true, "confirm action")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("unexpected positional argument; boolean flags must use --flag=true or --flag=false")
	}

	if !filepath.IsAbs(*file) || *event == "" || *callback == "" || *token == "" || *duration <= 0 || *duration > 3600 {
		log.Print("popup requires a valid OrganizerV2 notification")
		os.Exit(2)
	}
	if *position != "bottom_right" {
		log.Print("unsupported popup position")
		os.Exit(2)
	}
	host, _, addressErr := net.SplitHostPort(*callback)
	if addressErr != nil || host != "127.0.0.1" {
		log.Fatal("popup callbacks must use an ephemeral IPv4 loopback endpoint")
	}
	var choices []PopupChoice
	if err := json.Unmarshal([]byte(*choicesRaw), &choices); err != nil {
		log.Fatal(err)
	}
	if len(choices) == 0 {
		log.Fatal("popup requires a current destination selection")
	}

	details := PopupDetails{
		Filename: filepath.Base(*file),
		Folder: filepath.Base(filepath.Dir(*file)),
		Destination: filepath.Dir(*file),
		Category: *category,
		Duration: *duration,
		Choices: choices,
		OpenFile: *openFile,
		OpenLocation: *openLocation,
		Confirm: *confirm,
		DragEnabled: *dragEnabled,
	}
	service := NewPopupService(*file, *category, *duration)
	service.configure(*callback, *token, details)
	// The Wails production default discards its own logs. In a Windows GUI
	// binary there is no console either, so keep errors on stderr for the
	// watcher's bounded subprocess diagnostic capture.
	log.Printf("[popup] initializing Wails (pid=%d)", os.Getpid())
	app := application.New(application.Options{
		Name: "OrganizerV2 Popup",
		Logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
		ErrorHandler: func(err error) {
			log.Printf("[popup] Wails error: %v", err)
		},
		WarningHandler: func(message string) {
			log.Printf("[popup] Wails warning: %s", message)
		},
		Services: []application.Service{application.NewService(service)},
		Assets: application.AssetOptions{Handler: application.BundledAssetFileServer(assets)},
	})
	service.setApplication(app)
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "OrganizerV2",
		Width: popupWidth,
		Height: popupHeight,
		URL: "/",
		Hidden: true, // A partially loaded UI must never appear or count as READY.
		Frameless: true,
		AlwaysOnTop: true,
		DisableResize: true,
		HideOnEscape: true,
		DefaultContextMenuDisabled: true,
		Windows: application.WindowsWindow{HiddenOnTaskbar: true, DisableMenu: true},
	})
	service.setWindow(window)
	if err := app.Run(); err != nil {
		log.Fatalf("[popup] Wails Run failed: %v", err)
	}
}
