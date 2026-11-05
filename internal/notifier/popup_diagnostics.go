package notifier

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

const maxPopupDiagnosticBytes = 16 * 1024

// popupDiagnosticLog is a bounded, concurrency-safe tail buffer for helper
// stdout and stderr. Windows GUI executables do not have a visible console,
// and os/exec copies these two streams concurrently.
type popupDiagnosticLog struct {
	mu sync.Mutex
	data []byte
}

func (p *popupDiagnosticLog) Write(b []byte) (int, error) {
	n := len(b)
	p.mu.Lock()
	defer p.mu.Unlock()
	if n >= maxPopupDiagnosticBytes {
		p.data = append(p.data[:0], b[n-maxPopupDiagnosticBytes:]...)
		return n, nil
	}
	if len(p.data)+n > maxPopupDiagnosticBytes {
		keep := maxPopupDiagnosticBytes-n
		copy(p.data, p.data[len(p.data)-keep:])
		p.data = p.data[:keep]
	}
	p.data = append(p.data, b...)
	return n, nil
}

func (p *popupDiagnosticLog) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return string(p.data)
}

func popupStartupFailure(err error, output string) error {
	if err == nil {
		err = errors.New("helper exited without sending READY")
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return fmt.Errorf("popup helper exited before READY: %w (no helper diagnostics)", err)
	}
	return fmt.Errorf("popup helper exited before READY: %w; helper diagnostics: %s", err, output)
}
