package notifier

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestPopupDiagnosticLogKeepsTail(t *testing.T) {
	var log popupDiagnosticLog
	log.Write([]byte(strings.Repeat("a", maxPopupDiagnosticBytes)))
	log.Write([]byte("LAST-ERROR"))
	got := log.String()
	if len(got) != maxPopupDiagnosticBytes || !strings.HasSuffix(got, "LAST-ERROR") {
		t.Fatalf("tail log lost the last error, len=%d", len(got))
	}
	log.Write([]byte(strings.Repeat("b", maxPopupDiagnosticBytes+23)))
	if got := log.String(); len(got) != maxPopupDiagnosticBytes || strings.Contains(got, "a") {
		t.Fatalf("oversized write was not bounded: %d", len(got))
	}
}

func TestPopupDiagnosticLogConcurrentStreams(t *testing.T) {
	var log popupDiagnosticLog
	var wg sync.WaitGroup
	for i:=0; i<8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j:=0;j<100;j++ { _, _ = log.Write([]byte("some stderr message\n")) }
		}()
	}
	wg.Wait()
	if len(log.String()) > maxPopupDiagnosticBytes || len(log.String()) == 0 {
		t.Fatal("concurrent output log was not bounded")
	}
}

func TestPopupStartupFailureContainsDiagnostics(t *testing.T) {
	err := popupStartupFailure(errors.New("exit status 1"), "[popup] Wails error: failed to initialize WebView2")
	if !strings.Contains(err.Error(), "WebView2") || !strings.Contains(err.Error(), "exit status 1") {
		t.Fatalf("lost helper diagnostics: %v", err)
	}
	err = popupStartupFailure(nil, "")
	if !strings.Contains(err.Error(), "without sending READY") {
		t.Fatalf("lost unexpected clean exit: %v", err)
	}
}
