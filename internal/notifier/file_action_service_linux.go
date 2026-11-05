//go:build linux

package notifier

import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "sync"

    "github.com/vitorhugo-java/organizerv2/internal/pathutil"
)

// linuxFileActionService shares the Windows toast's action handler and collision
// policy, but invokes Linux shell tools. No user-supplied paths cross the IPC API.
type linuxFileActionService struct {
    notifier *linuxNotifier
    mu sync.Mutex
}

func (s *linuxFileActionService) OpenFile(path string) error {
    info, err := os.Stat(path)
    if err != nil { return fmt.Errorf("stat file: %w", err) }
    if !filepath.IsAbs(path) || !info.Mode().IsRegular() { return errors.New("invalid file") }
    return exec.Command("xdg-open", path).Start()
}

func (s *linuxFileActionService) OpenLocation(path string) error {
    if !filepath.IsAbs(path) { return errors.New("path must be absolute") }
    s.notifier.openLocation(path)
    return nil
}

func (s *linuxFileActionService) CopyPath(path string) error {
    if !filepath.IsAbs(path) { return errors.New("path must be absolute") }
    s.notifier.copyToClipboard(path)
    return nil
}

func (s *linuxFileActionService) MoveTo(path, destinationDir string) (string, error) {
    return s.transfer(path, destinationDir, true)
}

func (s *linuxFileActionService) CopyTo(path, destinationDir string) (string, error) {
    return s.transfer(path, destinationDir, false)
}

func (s *linuxFileActionService) transfer(path, directory string, move bool) (string, error) {
    if !filepath.IsAbs(path) || !filepath.IsAbs(directory) {
        return "", errors.New("source and destination must be absolute")
    }
    info, err := os.Stat(path)
    if err != nil { return "", fmt.Errorf("stat source: %w", err) }
    if !info.Mode().IsRegular() { return "", errors.New("source is not a regular file") }
    if err := pathutil.EnsureDir(directory); err != nil { return "", fmt.Errorf("create destination: %w", err) }
    s.mu.Lock()
    defer s.mu.Unlock()
    base := filepath.Join(directory, filepath.Base(path))
    for attempt := 0; attempt < 1000; attempt++ {
        target, err := pathutil.ResolveDuplicate(base)
        if err != nil { return "", err }
        if move { err = pathutil.MoveFile(path, target) } else { err = pathutil.CopyFile(path, target) }
        if err == nil { return target, nil }
        if !errors.Is(err, os.ErrExist) { return "", err }
    }
    return "", fmt.Errorf("could not reserve destination for %s", filepath.Base(path))
}
