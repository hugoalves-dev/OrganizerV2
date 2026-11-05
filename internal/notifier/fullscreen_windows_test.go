//go:build windows

package notifier

import (
    "os"
    "path/filepath"
    "sync/atomic"
    "testing"
    "time"

    "github.com/vitorhugo-java/organizerv2/internal/config"
)

func TestFullscreenWindowBounds(t *testing.T) {
    primary := fullscreenRect{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
    secondary := fullscreenRect{Left: -2560, Top: 0, Right: 0, Bottom: 1440}
    cases := []struct {
        name string
        window fullscreenRect
        monitor fullscreenRect
        want bool
    }{
        {"exclusive", primary, primary, true},
        {"borderless secondary", secondary, secondary, true},
        {"fractional dpi rounding", fullscreenRect{1, 0, 1919, 1079}, primary, true},
        {"maximized with taskbar", fullscreenRect{0, 0, 1920, 1040}, primary, false},
        {"ordinary window", fullscreenRect{100, 100, 1100, 900}, primary, false},
        {"invalid rect", fullscreenRect{}, primary, false},
    }
    for _, test := range cases {
        t.Run(test.name, func(t *testing.T) {
            if got := fullscreenCoversMonitor(test.window, test.monitor); got != test.want {
                t.Fatalf("fullscreenCoversMonitor=%t want=%t", got, test.want)
            }
        })
    }
}

func TestFullscreenPolicySilentAndShow(t *testing.T) {
    n, _ := newWindowsNotifierFixture(t, func(*windowsToast) error {
        t.Error("unexpected native toast")
        return nil
    })
    n.cfg = config.Default().Notifications
    n.fullscreenActive = func() bool { return true }
    popup := &fakePopupPresenter{}
    n.popup = popup
    event := FileEvent{Destination: filepath.Join(t.TempDir(), "file.txt"), Category: "Documents"}

    n.cfg.Custom.FullscreenBehavior = "silent"
    n.deliver(event)
    if popup.calls != 0 {
        t.Fatalf("silent mode showed %d popups", popup.calls)
    }

    n.cfg.Custom.FullscreenBehavior = "show"
    n.deliver(event)
    if popup.calls != 1 {
        t.Fatalf("show mode showed %d popups, want 1", popup.calls)
    }
}

func TestFullscreenPolicyDefersWithoutNativeFallback(t *testing.T) {
    n, _ := newWindowsNotifierFixture(t, func(*windowsToast) error {
        t.Error("fullscreen defer fell back to native toast")
        return nil
    })
    n.cfg = config.Default().Notifications
    var fullscreen atomic.Bool
    fullscreen.Store(true)
    n.fullscreenActive = fullscreen.Load
    popup := &fakePopupPresenter{}
    n.popup = popup
    delivered := make(chan FileEvent, 1)
    n.deferred = newDeferredPopupQueue(fullscreen.Load, func(event FileEvent) { delivered <- event })
    defer func() { n.deferred.stopQueue(); n.deferred.wait() }()

    path := filepath.Join(t.TempDir(), "file.txt")
    if err := os.WriteFile(path, []byte("file"), 0600); err != nil {
        t.Fatal(err)
    }
    n.deliver(FileEvent{Destination: path, Category: "Documents"})
    if !n.deferred.pending() || popup.calls != 0 {
        t.Fatal("fullscreen notification was not deferred")
    }
    fullscreen.Store(false)
    select { case n.deferred.wake <- struct{}{}: default: }
    select {
    case event := <-delivered:
        if event.Destination != path {
            t.Fatalf("wrong deferred file: %+v", event)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("deferred notification not replayed")
    }
}
