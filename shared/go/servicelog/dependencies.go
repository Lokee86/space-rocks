package servicelog

import (
	"io"
	"io/fs"
	"os"
	"time"
)

// maintenanceTicker abstracts the periodic flush loop timer for deterministic tests.
type maintenanceTicker interface {
	C() <-chan time.Time
	Stop()
}

type realMaintenanceTicker struct {
	ticker *time.Ticker
}

func (t *realMaintenanceTicker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *realMaintenanceTicker) Stop() {
	t.ticker.Stop()
}

// runtimeDependencies keeps runtime-owned I/O and time operations concrete so
// the rolling sink can be added without coupling policy to process globals.
type runtimeDependencies struct {
	consoleWriter io.Writer
	now           func() time.Time
	newTicker     func(time.Duration) maintenanceTicker
	mkdir         func(string, fs.FileMode) error
	openFile      func(string, int, fs.FileMode) (*os.File, error)
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
		newTicker: func(d time.Duration) maintenanceTicker {
			return &realMaintenanceTicker{ticker: time.NewTicker(d)}
		},
		mkdir:    os.MkdirAll,
		openFile: os.OpenFile,
		readFile: os.ReadFile,
		readDir:  os.ReadDir,
		rename:   os.Rename,
		remove:   os.Remove,
		stat:     os.Stat,
	}
}
