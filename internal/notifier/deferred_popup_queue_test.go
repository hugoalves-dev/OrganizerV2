package notifier

import (
    "os"
    "path/filepath"
    "sync/atomic"
    "testing"
    "time"
)

func TestDeferredPopupWaitsForFullscreenExit(t *testing.T) {
    path := filepath.Join(t.TempDir(), "download.txt")
    if err := os.WriteFile(path, []byte("file"), 0600); err != nil {
        t.Fatal(err)
    }
    var fullscreen atomic.Bool
    fullscreen.Store(true)
    delivered := make(chan FileEvent, 1)
    q := newDeferredPopupQueue(fullscreen.Load, func(event FileEvent) {
        delivered <- event
    })
    defer func() { q.stopQueue(); q.wait() }()

    q.enqueue(FileEvent{Destination: path, Category: "Documents"})
    if !q.pending() {
        t.Fatal("notification not queued")
    }
    select {
    case <-delivered:
        t.Fatal("notification shown while fullscreen")
    case <-time.After(50 * time.Millisecond):
    }

    fullscreen.Store(false)
    select { case q.wake <- struct{}{}: default: }
    select {
    case event := <-delivered:
        if event.Destination != path {
            t.Fatalf("wrong notification: %+v", event)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("notification was not delivered after exiting fullscreen")
    }
}

func TestDeferredPopupQueueBoundedAndOldestDropped(t *testing.T) {
    q := newDeferredPopupQueue(func() bool { return true }, func(FileEvent) {
        t.Error("unexpected fullscreen delivery")
    })
    defer func() { q.stopQueue(); q.wait() }()

    for i := 0; i < maxDeferredPopups+1; i++ {
        q.enqueue(FileEvent{Destination: string(rune('a' + i))})
    }
    q.mu.Lock()
    count := len(q.events)
    first := q.events[0].event.Destination
    q.mu.Unlock()
    if count != maxDeferredPopups || first != "b" {
        t.Fatalf("count=%d first=%q; want %d and b", count, first, maxDeferredPopups)
    }
}

func TestDeferredPopupDropsExpiredAndMissingFiles(t *testing.T) {
    var fullscreen atomic.Bool
    fullscreen.Store(true)
    delivered := make(chan FileEvent, 2)
    q := newDeferredPopupQueue(fullscreen.Load, func(event FileEvent) {
        delivered <- event
    })
    defer func() { q.stopQueue(); q.wait() }()
    path := filepath.Join(t.TempDir(), "old.txt")
    if err := os.WriteFile(path, []byte("file"), 0600); err != nil {
        t.Fatal(err)
    }
    q.enqueue(FileEvent{Destination: path})
    q.enqueue(FileEvent{Destination: filepath.Join(t.TempDir(), "missing.txt")})
    q.mu.Lock()
    q.events[0].queuedAt = time.Now().Add(-deferredPopupTTL - time.Second)
    q.mu.Unlock()
    fullscreen.Store(false)
    select { case q.wake <- struct{}{}: default: }
    deadline := time.After(2 * time.Second)
    for {
        if !q.pending() {
            break
        }
        select {
        case <-deadline:
            t.Fatal("queue failed to drain expired and missing files")
        case <-time.After(10 * time.Millisecond):
        }
    }
    select {
    case event := <-delivered:
        t.Fatalf("stale or missing file delivered: %+v", event)
    default:
    }
}

func TestDeferredPopupDeliveryIsSequential(t *testing.T) {
    var fullscreen atomic.Bool
    fullscreen.Store(true)
    entered := make(chan struct{}, 2)
    release := make(chan struct{})
    q := newDeferredPopupQueue(fullscreen.Load, func(FileEvent) {
        entered <- struct{}{}
        <-release
    })
    path := filepath.Join(t.TempDir(), "file.txt")
    if err := os.WriteFile(path, []byte("file"), 0600); err != nil {
        t.Fatal(err)
    }
    q.enqueue(FileEvent{Destination: path})
    q.enqueue(FileEvent{Destination: path})
    fullscreen.Store(false)
    select { case q.wake <- struct{}{}: default: }
    select {
    case <-entered:
    case <-time.After(2 * time.Second):
        t.Fatal("first popup not delivered")
    }
    select {
    case <-entered:
        t.Fatal("second popup shown concurrently")
    case <-time.After(50 * time.Millisecond):
    }
    close(release)
    select {
    case <-entered:
    case <-time.After(2 * time.Second):
        t.Fatal("second popup did not follow the first")
    }
    q.stopQueue()
    q.wait()
}
