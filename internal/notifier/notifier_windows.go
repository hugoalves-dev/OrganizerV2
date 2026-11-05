//go:build windows

package notifier

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/vitorhugo-java/organizerv2/internal/config"
)

type toastPusher func(*windowsToast) error

// Allows provider-routing tests without launching Wails/WebView2.
type popupPresenter interface {
    Show(notificationEvent, config.NotificationConfig, []resolvedShortcut) error
    Close()
}

type windowsNotifier struct {
	cfg        config.NotificationConfig
	executable string
	registry   *notificationEventRegistry
	shortcuts  *notificationShortcutResolver
	handler    *windowsNotificationActionHandler
	files      fileActionService
	push       toastPusher
	popup popupPresenter
	fullscreenActive func() bool
	deferred *deferredPopupQueue
	closeOnce  sync.Once
	closeErr   error
}

func newPlatform(cfg config.NotificationConfig) Notifier {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("[notifier] executable lookup failed: %v", err)
	}
	if executable != "" {
		absolute, absErr := filepath.Abs(executable)
		if absErr != nil {
			log.Printf("[notifier] executable normalization failed: %v", absErr)
		} else {
			executable = absolute
		}
	}

	registry := newNotificationEventRegistry(defaultEventRegistryOptions())
	resolver := newNotificationShortcutResolver(cfg.Shortcuts)
	files := newWindowsFileActionService(cfg.Actions.CopyPath)
	handler := &windowsNotificationActionHandler{
		registry:  registry,
		shortcuts: &resolver,
		files:     files,
		logf:      log.Printf,
	}
	if err := initializeWindowsActivation(executable); err != nil {
		log.Printf("[notifier] activation initialization failed: %v", err)
	}
	installWindowsActivationHandler(handler)

	notifier := &windowsNotifier{
        popup: newCustomPopupLauncher(executable,handler),
        fullscreenActive: windowsFullscreenActive,
		cfg:        cfg,
		executable: executable,
		registry:   registry,
		shortcuts:  &resolver,
		handler:    handler,
		files:      files,
		push: func(notification *windowsToast) error {
			return notification.Push()
		},
	}
    if cfg.Provider == "custom" && cfg.Custom.RespectFullscreen && cfg.Custom.FullscreenBehavior == "defer" {
        notifier.deferred = newDeferredPopupQueue(notifier.fullscreenActive, notifier.deliverNow)
    }
    return notifier
}

func (n *windowsNotifier) Notify(event FileEvent) error {
	go n.deliver(event)
	return nil
}

func (n *windowsNotifier) deliver(event FileEvent) {
    if n.cfg.Provider == "custom" && n.cfg.Custom.RespectFullscreen && n.fullscreenActive != nil {
        fullscreen := n.fullscreenActive()
        switch n.cfg.Custom.FullscreenBehavior {
        case "silent":
            if fullscreen {
                log.Print("[notifier] popup suppressed while fullscreen")
                return
            }
        case "defer":
            // Preserve the pending order when returning to the desktop.
            if fullscreen || (n.deferred != nil && n.deferred.pending()) {
                if n.deferred != nil {
                    n.deferred.enqueue(event)
                }
                return
            }
        }
    }
    n.deliverNow(event)
}

func (n *windowsNotifier) deliverNow(event FileEvent) {
	destination, err := filepath.Abs(event.Destination)
	if err != nil {
		log.Printf("[notifier] destination normalization failed: %v", err)
		return
	}
	registered, err := n.registry.Register(destination, event.Category)
	if err != nil {
		log.Printf("[notifier] event registration failed: %v", err)
		return
	}

	var shortcuts []resolvedShortcut
	if n.shortcuts != nil {
		shortcuts = n.shortcuts.All()
	}
    if n.cfg.Provider == "custom" && n.popup != nil {
        if err := n.popup.Show(registered, n.cfg, shortcuts); err == nil {
            // The popup is finished (dismissed, action, or timeout); unlike
            // a native toast, it can no longer be activated.
            n.registry.Remove(registered.ID)
            return
        } else {
            log.Printf("[notifier] custom popup failed; falling back to native: %v", err)
        }
    }
	notification := buildWindowsToast(registered, n.cfg, shortcuts, n.executable)
	if n.push == nil {
		n.registry.Remove(registered.ID)
		log.Printf("[notifier] toast pusher is unavailable")
		return
	}
	if err := n.push(&notification); err != nil {
		n.registry.Remove(registered.ID)
		log.Printf("[notifier] toast push error: %v", err)
	}
}

func (n *windowsNotifier) Close() error {
	n.closeOnce.Do(func() {
        if n.deferred != nil { n.deferred.stopQueue() }
        if n.popup != nil { n.popup.Close() }
        if n.deferred != nil { n.deferred.wait() }
		clearWindowsActivationHandler(n.handler)
		if n.registry != nil {
			n.closeErr = n.registry.Close()
		}
	})
	return n.closeErr
}
