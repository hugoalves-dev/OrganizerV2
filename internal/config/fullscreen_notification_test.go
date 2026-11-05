package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestFullscreenDefaultsAndOverrides(t *testing.T) {
    tests := []struct {
        name string
        yaml string
        respect bool
        behavior string
    }{
        {"defaults", "notifications:\n  enabled: true\n", true, "defer"},
        {"legacy custom block", "notifications:\n  custom:\n    duration_seconds: 5\n", true, "defer"},
        {"explicit show", "notifications:\n  custom:\n    fullscreen_behavior: show\n", true, "show"},
        {"explicit silent", "notifications:\n  custom:\n    fullscreen_behavior: silent\n", true, "silent"},
        {"disable detection", "notifications:\n  custom:\n    respect_fullscreen: false\n", false, "defer"},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            path := filepath.Join(t.TempDir(), "config.yaml")
            if err := os.WriteFile(path, []byte(test.yaml), 0600); err != nil {
                t.Fatal(err)
            }
            cfg, err := Load(path)
            if err != nil {
                t.Fatal(err)
            }
            if cfg.Notifications.Custom.RespectFullscreen != test.respect ||
                cfg.Notifications.Custom.FullscreenBehavior != test.behavior {
                t.Fatalf("got fullscreen config %+v; want respect=%t behavior=%q",
                    cfg.Notifications.Custom, test.respect, test.behavior)
            }
        })
    }
    cfg := Default()
    if !cfg.Notifications.Custom.RespectFullscreen || cfg.Notifications.Custom.FullscreenBehavior != "defer" {
        t.Fatalf("invalid defaults: %+v", cfg.Notifications.Custom)
    }
}

func TestFullscreenBehaviorRejectsInvalidConfig(t *testing.T) {
    path := filepath.Join(t.TempDir(), "config.yaml")
    if err := os.WriteFile(path, []byte("notifications:\n  custom:\n    fullscreen_behavior: surprise\n"), 0600); err != nil {
        t.Fatal(err)
    }
    if _, err := Load(path); err == nil {
        t.Fatal("invalid fullscreen behavior accepted")
    }
}
