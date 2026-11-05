package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// This must run AFTER wails3 generate bindings in CI. A missing ES module
// previously produced an empty popup followed by a misleading native toast.
func TestEmbeddedFrontendHasWorkingBindings(t *testing.T) {
	content, err := fs.Sub(assets, "assets")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"index.html",
		"main.js",
		"bindings/github.com/vitorhugo-java/organizerv2-popup/popupservice.js",
	} {
		data, err := fs.ReadFile(content, name)
		if err != nil || len(data) == 0 {
			t.Fatalf("required popup asset %s missing or empty: %v", name, err)
		}
	}
	bindings, err := fs.ReadFile(content, "bindings/github.com/vitorhugo-java/organizerv2-popup/popupservice.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"Details", "Ready", "Action", "StartDrag", "Close"} {
		if !strings.Contains(string(bindings), "function "+method+"(") {
			t.Errorf("missing Wails binding for %s", method)
		}
	}

	server := popupDiagnosticHandler(application.BundledAssetFileServer(assets))
	for _, path := range []string{"/", "/main.js", "/bindings/github.com/vitorhugo-java/organizerv2-popup/popupservice.js", "/wails/runtime.js"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("%s returned %d", path, response.Code)
		}
		if response.Body.Len() == 0 {
			t.Errorf("%s returned empty body", path)
		}
	}
}
