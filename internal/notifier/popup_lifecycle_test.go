package notifier

import (
    "path/filepath"
    "runtime"
    "strings"
    "testing"

    "github.com/vitorhugo-java/organizerv2/internal/config"
)

func TestPopupMissingHelperFallsBackOnEveryPlatform(t *testing.T) {
    helper := filepath.Join(t.TempDir(), "organizer")
    if runtime.GOOS == "windows" { helper += ".exe" }
    launcher := newCustomPopupLauncher(helper, &windowsNotificationActionHandler{})
    err := launcher.Show(notificationEvent{CurrentPath: filepath.Join(t.TempDir(),"test.txt")}, config.Default().Notifications, nil)
    if err == nil || !strings.Contains(err.Error(), "not installed") { t.Fatalf("expected native fallback: %v", err) }
    if len(launcher.slots) != 0 { t.Fatal("capacity token leaked after failed popup") }
}

func TestPopupClosedRejectsNewSessions(t *testing.T) {
    launcher := newCustomPopupLauncher(filepath.Join(t.TempDir(), "organizer"), &windowsNotificationActionHandler{})
    launcher.Close()
    if err := launcher.Show(notificationEvent{}, config.Default().Notifications, nil); err == nil || !strings.Contains(err.Error(), "closing") {
        t.Fatalf("closed launcher accepted event: %v", err)
    }
}

func TestPopupCapacityFallsBackWithoutBlocking(t *testing.T) {
    launcher := newCustomPopupLauncher(filepath.Join(t.TempDir(), "organizer"), &windowsNotificationActionHandler{})
    for i:=0;i<cap(launcher.slots);i++ { launcher.slots<-struct{}{} }
    if err := launcher.Show(notificationEvent{}, config.Default().Notifications, nil); err == nil || !strings.Contains(err.Error(), "capacity") {
        t.Fatalf("expected capacity fallback: %v", err)
    }
}
