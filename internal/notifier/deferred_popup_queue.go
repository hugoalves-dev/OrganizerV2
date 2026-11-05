package notifier

import (
    "log"
    "os"
    "sync"
    "time"
)

const (
    maxDeferredPopups = 8
    deferredPopupTTL = 10 * time.Minute
    deferredPopupPoll = time.Second
)

type deferredPopup struct {
    event FileEvent
    queuedAt time.Time
}

// deferredPopupQueue holds notifications only while fullscreen, then replays
// them one at a time. It never blocks the filesystem watcher and never falls
// back to native toasts merely because the user is playing a game.
type deferredPopupQueue struct {
    mu sync.Mutex
    events []deferredPopup
    presenting bool
    closed bool
    wake chan struct{}
    stop chan struct{}
    done chan struct{}
    once sync.Once
    now func() time.Time
    fullscreen func() bool
    show func(FileEvent)
}

func newDeferredPopupQueue(fullscreen func() bool, show func(FileEvent)) *deferredPopupQueue {
    q := &deferredPopupQueue{
        wake: make(chan struct{}, 1),
        stop: make(chan struct{}),
        done: make(chan struct{}),
        now: time.Now,
        fullscreen: fullscreen,
        show: show,
    }
    go q.run()
    return q
}

func (q *deferredPopupQueue) enqueue(event FileEvent) {
    q.mu.Lock()
    if q.closed {
        q.mu.Unlock()
        return
    }
    if len(q.events) >= maxDeferredPopups {
        q.events = q.events[1:]
        log.Print("[notifier] fullscreen deferred queue full; oldest notification dropped")
    }
    q.events = append(q.events, deferredPopup{event: event, queuedAt: q.now()})
    q.mu.Unlock()
    select {
    case q.wake <- struct{}{}:
    default:
    }
}

func (q *deferredPopupQueue) pending() bool {
    q.mu.Lock()
    defer q.mu.Unlock()
    return q.presenting || len(q.events) != 0
}

func (q *deferredPopupQueue) take() (deferredPopup, bool) {
    q.mu.Lock()
    defer q.mu.Unlock()
    if len(q.events) == 0 {
        return deferredPopup{}, false
    }
    item := q.events[0]
    q.events = q.events[1:]
    q.presenting = true
    return item, true
}

func (q *deferredPopupQueue) finish() {
    q.mu.Lock()
    q.presenting = false
    q.mu.Unlock()
}

func (q *deferredPopupQueue) putBack(item deferredPopup) {
    q.mu.Lock()
    q.events = append([]deferredPopup{item}, q.events...)
    if len(q.events) > maxDeferredPopups {
        q.events = q.events[:maxDeferredPopups]
    }
    q.presenting = false
    q.mu.Unlock()
}

func (q *deferredPopupQueue) run() {
    defer close(q.done)
    ticker := time.NewTicker(deferredPopupPoll)
    defer ticker.Stop()
    for {
        select {
        case <-q.stop:
            return
        case <-q.wake:
        case <-ticker.C:
        }
        for {
            select {
            case <-q.stop:
                return
            default:
            }
            if q.fullscreen() {
                break
            }
            item, ok := q.take()
            if !ok {
                break
            }
            if q.now().Sub(item.queuedAt) > deferredPopupTTL {
                log.Print("[notifier] expired fullscreen notification discarded")
                q.finish()
                continue
            }
            file, err := os.Stat(item.event.Destination)
            if err != nil || !file.Mode().IsRegular() {
                log.Print("[notifier] deferred notification skipped: organized file is no longer available")
                q.finish()
                continue
            }
            // Fullscreen may resume while checking the file. Keep the event
            // rather than stealing focus when the next match/game starts.
            if q.fullscreen() {
                q.putBack(item)
                break
            }
            select {
            case <-q.stop:
                q.finish()
                return
            default:
            }
            q.show(item.event) // Blocks until dismissed; no popup avalanche.
            q.finish()
        }
    }
}

func (q *deferredPopupQueue) stopQueue() {
    q.once.Do(func() {
        q.mu.Lock()
        q.closed = true
        q.events = nil
        q.mu.Unlock()
        close(q.stop)
    })
}

func (q *deferredPopupQueue) wait() {
    <-q.done
}
