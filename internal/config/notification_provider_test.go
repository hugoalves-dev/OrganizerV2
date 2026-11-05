package config

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)
func TestNotificationProviderDefaultsAndLegacyYAML(t *testing.T) {
    cfg:=Default()
    if cfg.Notifications.Provider!="custom" || cfg.Notifications.Custom.DurationSeconds!=15 || cfg.Notifications.Custom.Position!="bottom_right" || !cfg.Notifications.Custom.DragAndDrop.Enabled { t.Fatalf("unexpected defaults: %+v", cfg.Notifications) }
    for _, input := range []string{"notifications:\n  enabled: true\n","notifications:\n  enabled: true\n  provider: native\n"} {
        path:=filepath.Join(t.TempDir(),"config.yaml")
        if err:=os.WriteFile(path,[]byte(input),0600);err!=nil {t.Fatal(err)}
        loaded,err:=Load(path);if err!=nil {t.Fatal(err)}
        want:="custom";if strings.Contains(input,"provider: native"){want="native"}
        if loaded.Notifications.Provider!=want {t.Fatalf("provider: got %q want %q",loaded.Notifications.Provider,want)}
    }
}
func TestNotificationProviderValidation(t *testing.T) {
    cases:=[]string{
        "notifications:\n  provider: invalid\n",
        "notifications:\n  provider: custom\n  custom:\n    position: top_left\n",
        "notifications:\n  custom:\n    drag_and_drop:\n      operation: move\n",
    }
    for _, content:=range cases {
        path:=filepath.Join(t.TempDir(),"config.yaml")
        if err:=os.WriteFile(path,[]byte(content),0600);err!=nil{t.Fatal(err)}
        if _,err:=Load(path);err==nil{t.Fatalf("accepted invalid config: %s",content)}
    }
}
