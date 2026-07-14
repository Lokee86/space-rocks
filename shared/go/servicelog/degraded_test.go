package servicelog

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"
)

func TestOpenFallsBackAndWarnsWithoutConsoleHandler(t *testing.T) {
	initErr := errors.New("open failed")
	var console bytes.Buffer

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig("logs"),
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, runtimeDependencies{
		consoleWriter: &console,
		now:           time.Now,
		mkdir:         func(string, fs.FileMode) error { return nil },
		openFile:      func(string, int, fs.FileMode) (io.WriteCloser, error) { return nil, initErr },
	})
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v, want nil", err)
	}
	if runtime == nil {
		t.Fatal("openWithDependencies() returned nil runtime")
	}

	status := runtime.Status()
	if !status.Degraded || status.FailureCount != 1 || status.LastError != initErr.Error() {
		t.Fatalf("status = %+v, want degraded snapshot", status)
	}

	runtime.Logger().Info("discarded ok")
	if got := strings.Count(console.String(), "servicelog warning:"); got != 1 {
		t.Fatalf("warning count = %d, want 1; output = %q", got, console.String())
	}
	if strings.Contains(console.String(), "discarded ok") {
		t.Fatalf("console-disabled runtime should not emit normal slog output: %q", console.String())
	}
}
