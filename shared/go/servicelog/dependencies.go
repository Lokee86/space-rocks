package servicelog

import (
	"io"
	"io/fs"
	"os"
	"time"
)

// runtimeTicker abstracts the periodic flush clock so tests can drive the loop
// deterministically.
type runtimeTicker interface {
	C() <-chan time.Time
	Stop()
}

type stdTicker struct {
	ticker *time.Ticker
}

func (t *stdTicker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *stdTicker) Stop() {
	t.ticker.Stop()
}

// runtimeDependencies keeps runtime-owned I/O and time operations concrete so
// the rolling sink can be added without coupling policy to process globals.
type runtimeDependencies struct {
	consoleWriter io.Writer
	now           func() time.Time
	newTicker     func(time.Duration) runtimeTicker
	reportFailure func(error)
	mkdir         func(string, fs.FileMode) error
	openFile      func(string, int, fs.FileMode) (io.WriteCloser, error)
	readFile      func(string) ([]byte, error)
	readDir       func(string) ([]os.DirEntry, error)
	rename        func(string, string) error
	remove        func(string) error
	stat          func(string) (os.FileInfo, error)
}

func defaultRuntimeDependencies() runtimeDependencies {
	return runtimeDependencies{
		consoleWriter: os.Stderr,
		now:           time.Now,
		newTicker: func(d time.Duration) runtimeTicker {
			return &stdTicker{ticker: time.NewTicker(d)}
		},
		mkdir: os.MkdirAll,
		openFile: func(name string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			return os.OpenFile(name, flag, perm)
		},
		readFile: os.ReadFile,
		readDir:  os.ReadDir,
		rename:   os.Rename,
		remove:   os.Remove,
		stat:     os.Stat,
	}
}
