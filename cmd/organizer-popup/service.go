package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	dragout "github.com/eaxum/wails-dragout"
	wailsdrag "github.com/eaxum/wails-dragout/wails"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type PopupChoice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type PopupDetails struct {
	Filename     string        `json:"filename"`
	Category     string        `json:"category"`
	Folder       string        `json:"folder"`
	Destination  string        `json:"destination"`
	Duration     int           `json:"duration"`
	Choices      []PopupChoice `json:"choices"`
	OpenFile     bool          `json:"openFile"`
	OpenLocation bool          `json:"openLocation"`
	Confirm      bool          `json:"confirm"`
	DragEnabled  bool          `json:"dragEnabled"`
}
type PopupService struct {
	path     string
	category string
	duration int
	app      *application.App
	window   *application.WebviewWindow
	mu       sync.Mutex
	options  PopupDetails
	callback string
	authKey  string
}

func NewPopupService(path, category string, duration int) *PopupService {
	return &PopupService{path: path, category: category, duration: duration}
}
func (s *PopupService) setApplication(app *application.App) { s.app = app }
func (s *PopupService) Filename() string                    { return filepath.Base(s.path) }
func (s *PopupService) Category() string                    { return s.category }
func (s *PopupService) Duration() int                       { return s.duration }
func (s *PopupService) Close() {
	if s.app != nil {
		go s.app.Quit()
	}
}
func (s *PopupService) StartDrag(ctx context.Context) (string, error) {
	if !s.options.DragEnabled {
		return "", errors.New("drag-out is disabled")
	}
	if !s.mu.TryLock() {
		return "", errors.New("drag already running")
	}
	defer s.mu.Unlock()
	if !filepath.IsAbs(s.path) {
		return "", errors.New("path must be absolute")
	}
	info, err := os.Stat(s.path)
	if err != nil {
		return "", fmt.Errorf("file unavailable: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("only regular files supported")
	}
	if !dragout.Available {
		return "", dragout.ErrUnsupported
	}
	w, ok := ctx.Value(application.WindowKey).(*application.WebviewWindow)
	if !ok || w == nil {
		return "", errors.New("Wails window unavailable")
	}
	effect, err := wailsdrag.StartWithOptions(ctx, w, []string{s.path}, dragOptionsFor(dragout.SupportedEffects()))
	if err != nil {
		return "", err
	}
	return dragEffectResult(effect), nil
}

func dragOptionsFor(supported dragout.Effects) dragout.Options {
	preferred := dragout.EffectCopy
	if supported&dragout.EffectSetMove != 0 {
		preferred = dragout.EffectMove
	}
	return dragout.Options{AllowedEffects: supported, PreferredEffect: preferred}
}

func dragEffectResult(effect dragout.Effect) string {
	switch effect {
	case dragout.EffectCopy:
		return "copy"
	case dragout.EffectMove:
		return "move"
	case dragout.EffectLink:
		return "link"
	case dragout.EffectNone:
		return "none"
	default:
		return "unknown"
	}
}

func (s *PopupService) configure(address, key string, details PopupDetails) {
	s.callback = address
	s.authKey = key
	s.options = details
}
func (s *PopupService) Details() PopupDetails {
	log.Print("[popup] Wails binding: Details called")
	return s.options
}

func (s *PopupService) setWindow(window *application.WebviewWindow) { s.window = window }

const (
	popupWidth  = 380
	popupHeight = 270
	popupMargin = 16
)

// Coordinates and dimensions are in Wails logical pixels (not scaled physical
// pixels). WorkArea excludes the taskbar and can have a non-zero origin.
func popupBottomRight(area application.Rect, width, height, margin int) (int, int) {
	x := area.X + area.Width - width - margin
	y := area.Y + area.Height - height - margin
	if x < area.X {
		x = area.X
	}
	if y < area.Y {
		y = area.Y
	}
	return x, y
}

func (s *PopupService) place() {
	if s.app == nil || s.window == nil {
		return
	}
	screen := s.app.Screen.GetPrimary()
	if screen == nil {
		return
	}
	x, y := popupBottomRight(screen.WorkArea, popupWidth, popupHeight, popupMargin)
	s.window.SetPosition(x, y)
}
