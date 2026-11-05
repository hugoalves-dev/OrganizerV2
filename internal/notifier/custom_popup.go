package notifier

import (
    "crypto/rand"
    "crypto/subtle"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net"
    "net/http"
    "os"
    "runtime"
    "os/exec"
    "path/filepath"
    "sync"
    "time"

    "github.com/vitorhugo-java/organizerv2/internal/config"
)

// customPopupLauncher starts an optional, platform-specific Wails helper.
// The headless watcher never links CGO/WebView or AGPL code. Actions use
// the same event registry and file-action handler as native Windows toasts.
type customPopupLauncher struct {
    executable string
    handler *windowsNotificationActionHandler
    slots chan struct{}
    closed chan struct{}
    once sync.Once
}
type popupActionRequest struct {
    Action string `json:"action"`
    Selection string `json:"selection"`
}
func newCustomPopupLauncher(executable string, handler *windowsNotificationActionHandler) *customPopupLauncher {
    return &customPopupLauncher{executable: executable, handler: handler, slots: make(chan struct{}, 4), closed: make(chan struct{})}
}
func (l *customPopupLauncher) Close() { l.once.Do(func(){ close(l.closed) }) }
func (l *customPopupLauncher) Show(event notificationEvent, cfg config.NotificationConfig, shortcuts []resolvedShortcut) error {
    select {
    case <-l.closed: return fmt.Errorf("popup launcher is closing")
    case l.slots <- struct{}{}:
    default: return fmt.Errorf("popup capacity reached; using native notification")
    }
    defer func(){<-l.slots}()

    // Helper is installed next to the watcher executable. It has its own
    // Go module, build and license policy. Never search PATH for arbitrary
    // executables bearing this name.
    helperName := "organizer-popup"
    if runtime.GOOS == "windows" { helperName += ".exe" }
    helper := filepath.Join(filepath.Dir(l.executable), helperName)
    if fi,err:=os.Stat(helper);err!=nil || fi.IsDir() { return fmt.Errorf("custom popup helper not installed next to organizer executable") }

    listener,err:=net.Listen("tcp","127.0.0.1:0")
    if err!=nil{return fmt.Errorf("cannot bind local popup callback: %w",err)}
    defer listener.Close()
    secret:=make([]byte,32)
    if _,err=rand.Read(secret);err!=nil{return err}
    token:=hex.EncodeToString(secret)
    ready:=make(chan struct{},1)

    mux:=http.NewServeMux()
    auth:=func(r *http.Request) bool {
        supplied:=r.Header.Get("X-Organizer-Key")
        return subtle.ConstantTimeCompare([]byte(supplied),[]byte(token))==1 &&
            r.Header.Get("Origin")=="" // helper uses Go http.Client, not browser fetch
    }
    mux.HandleFunc("/ready",func(w http.ResponseWriter,r *http.Request){
        if r.Method!="POST"||!auth(r){http.Error(w,"unauthorized",http.StatusUnauthorized);return}
        select{case ready<-struct{}{}:default:}
        w.WriteHeader(http.StatusNoContent)
    })
    mux.HandleFunc("/action",func(w http.ResponseWriter,r *http.Request){
        if r.Method!="POST"||!auth(r){http.Error(w,"unauthorized",http.StatusUnauthorized);return}
        r.Body=http.MaxBytesReader(w,r.Body,1024)
        defer r.Body.Close()
        var input popupActionRequest
        if err:=json.NewDecoder(r.Body).Decode(&input);err!=nil{http.Error(w,"invalid payload",400);return}
        // Only UI-visible actions are allowed, and each is guarded by the
        // existing user's notification flags, not by input from the helper.
        allowed:=map[string]bool{
            "open_file":cfg.Actions.OpenFile,
            "open_location":cfg.Actions.OpenLocation,
            "confirm":cfg.Actions.Confirm,
        }
        if !allowed[input.Action] {http.Error(w,"disabled action",403);return}
        if input.Selection=="" {input.Selection=currentDestinationSelectionID}
        valid:=input.Selection==currentDestinationSelectionID
        if cfg.Actions.MoveTo && !valid {
            for _,sc:=range shortcuts {if sc.ID==input.Selection{valid=true;break}}
        }
        if !valid {http.Error(w,"invalid shortcut selection",403);return}
        select{case <-l.closed:http.Error(w,"closing",503);return;default:}
        // This applies the same one-shot Claim and collision avoidance as a
        // toast. The helper cannot supply an arbitrary destination path.
        l.handler.Handle(encodeNotificationAction(notificationAction(input.Action),event.ID),
            []inputValue{{Key:destinationInputID, Value:input.Selection}})
        w.WriteHeader(http.StatusNoContent)
    })
    server:=&http.Server{Handler:mux,ReadHeaderTimeout:2*time.Second}
    serverStopped:=make(chan struct{})
    go func(){ _=server.Serve(listener);close(serverStopped) }()
    defer func(){_ =server.Close();<-serverStopped}()

    args:=[]string{
        "--file",event.CurrentPath,
        "--category",event.Category,
        "--event",event.ID,
        "--callback",listener.Addr().String(),
        "--token",token,
        "--duration",fmt.Sprint(cfg.Custom.DurationSeconds),
        "--drag-enabled",fmt.Sprint(cfg.Custom.DragAndDrop.Enabled),
    }
    // Pass a bounded set of opaque shortcut IDs and display names only.
    type choice struct{ ID string `json:"id"`; Name string `json:"name"` }
    choices:=[]choice{{currentDestinationSelectionID,currentDestinationLabel(filepath.Dir(event.CurrentPath))}}
    if cfg.Actions.MoveTo {
        for _,sc:=range shortcuts {
            if !sameDestinationDirectory(event.CurrentPath,sc.Path) { choices=append(choices,choice{sc.ID,sc.Name}) }
        }
    }
    choicesJSON,err:=json.Marshal(choices);if err!=nil{return err}
    args=append(args,"--choices",string(choicesJSON),
        "--open-file",fmt.Sprint(cfg.Actions.OpenFile),
        "--open-location",fmt.Sprint(cfg.Actions.OpenLocation),
        "--confirm",fmt.Sprint(cfg.Actions.Confirm))
    cmd:=exec.Command(helper,args...)
    if err:=cmd.Start();err!=nil{return fmt.Errorf("cannot start custom popup: %w",err)}
    exited:=make(chan error,1)
    go func(){exited<-cmd.Wait()}()
    // Returning success before WebView is ready would silently lose the
    // notification. The 10s deadline triggers reliable native fallback.
    timer:=time.NewTimer(10*time.Second)
    defer timer.Stop()
    select{
    case <-ready:
        // Let hover/drag postpone dismissal; a hard upper bound handles frozen helpers.
        select{
        case err:=<-exited:
            if err!=nil {return fmt.Errorf("custom popup crashed after initialization: %w",err)}
        case <-l.closed:
            _=cmd.Process.Kill();<-exited
        case <-time.After(2*time.Hour):
            _=cmd.Process.Kill();<-exited
        }
        return nil
    case err:=<-exited:return fmt.Errorf("popup helper exited before ready: %v",err)
    case <-timer.C:
        _=cmd.Process.Kill();<-exited
        return fmt.Errorf("popup failed to initialize in 10 seconds")
    case <-l.closed:
        _=cmd.Process.Kill();<-exited
        return fmt.Errorf("popup launcher was closed")
    }
}

