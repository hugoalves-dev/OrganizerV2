package main

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestPopupDiagnosticEndpoint(t *testing.T){
    handler:=popupDiagnosticHandler(http.NotFoundHandler())
    for _,tc:=range []struct{
        method,body string
        want int
    }{
        {http.MethodPost,"module-import-error: missing bindings",http.StatusNoContent},
        {http.MethodGet,"",http.StatusMethodNotAllowed},
        {http.MethodPost,strings.Repeat("x",maxFrontendDiagnosticBytes+1),http.StatusRequestEntityTooLarge},
    }{
        response:=httptest.NewRecorder()
        handler.ServeHTTP(response,httptest.NewRequest(tc.method,"/popup-diagnostic",strings.NewReader(tc.body)))
        if response.Code != tc.want {
            t.Fatalf("method %s len %d: got %d, want %d",tc.method,len(tc.body),response.Code,tc.want)
        }
    }
    response:=httptest.NewRecorder()
    handler.ServeHTTP(response,httptest.NewRequest(http.MethodGet,"/unrecognized",nil))
    if response.Code != http.StatusNotFound{t.Fatalf("unexpected route result %d",response.Code)}
}

func TestPopupBootstrapCapturesImportFailures(t *testing.T){
    response:=httptest.NewRecorder()
    handler:=popupDiagnosticHandler(http.FileServer(http.Dir("assets")))
    handler.ServeHTTP(response,httptest.NewRequest(http.MethodGet,"/",nil))
    if response.Code!=http.StatusOK {t.Fatal(response.Code)}
    for _,required:=range []string{
        "window.popupDiagnostic = function",
        "import('./main.js').catch",
        "module-import-error",
        "unhandledrejection",
    }{
        if !strings.Contains(response.Body.String(),required){
            t.Fatalf("missing frontend startup diagnostic %q",required)
        }
    }
}
