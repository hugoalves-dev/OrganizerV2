package main

import (
    "io"
    "log"
    "net/http"
    "strings"
)

const maxFrontendDiagnosticBytes = 1024

// popupDiagnosticHandler gives a page with broken JS imports a way to report
// errors to the parent watcher's captured stderr, without requiring Wails
// bindings to work. This endpoint accepts diagnostics only, never actions.
func popupDiagnosticHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
        if r.URL.Path != "/popup-diagnostic" {
            next.ServeHTTP(w,r)
            return
        }
        if r.Method != http.MethodPost {
            w.Header().Set("Allow","POST")
            http.Error(w,"method not allowed",http.StatusMethodNotAllowed)
            return
        }
        defer r.Body.Close()
        body,err:=io.ReadAll(io.LimitReader(r.Body,maxFrontendDiagnosticBytes+1))
        if err!=nil { http.Error(w,"invalid diagnostics",http.StatusBadRequest);return }
        if len(body)>maxFrontendDiagnosticBytes {
            http.Error(w,"diagnostic too large",http.StatusRequestEntityTooLarge)
            return
        }
        // Quote rather than interpolate raw content: one error = one log line.
        log.Printf("[popup] frontend diagnostic: %q", strings.TrimSpace(string(body)))
        w.WriteHeader(http.StatusNoContent)
    })
}

// Log only Wails bootstrap resources; do not expose filenames, tokens or
// callback addresses from the popup command line.
func popupAssetRequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        if r.Method == http.MethodGet {
            path:=r.URL.Path
            if path=="/" || path=="/index.html" || path=="/main.js" ||
                path=="/wails/runtime.js" || strings.HasPrefix(path,"/bindings/") {
                log.Printf("[popup] browser requested %s",path)
            }
        }
        next.ServeHTTP(w,r)
    })
}
