//go:build windows
package notifier
import (
  "path/filepath"
  "strings"
  "testing"
  "github.com/vitorhugo-java/organizerv2/internal/config"
)
func TestCustomPopupMissingHelper(t *testing.T) {
 l:=newCustomPopupLauncher(filepath.Join(t.TempDir(),"organizer.exe"),&windowsNotificationActionHandler{})
 err:=l.Show(notificationEvent{CurrentPath:filepath.Join(t.TempDir(),"file.txt")},config.Default().Notifications,nil)
 if err==nil||!strings.Contains(err.Error(),"not installed"){t.Fatalf("expected fallback error: %v",err)}
}
