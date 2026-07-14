package servicelog

import (
	"io"
	"io/fs"
	"os"
	"time"
)

// runtimeDependencies keeps runtime-owned I/O and time operations concrete so
// the rolling sink can be added without coupling policy to process globals.
type runtimeDependencies struct {
	consoleWriter io.Writer
	now           func() time.Time
	mkdir         func(string, fs.FileMode) error
	openFile      func(string, int, fs.FileMode) (io.WriteCloser, error)
	readDir       func(string) ([]os.DirEntry, error)
	rename        func(string, string) error
	remove        func(string) error
	stat          func(string) (os.FileInfo, error)
}

func defaultRuntimeDependencies() runtimeDependencies {
	return runtimeDependencies{
		consoleWriter: os.Stderr,
		now:           time.Now,
		mkdir:         os.MkdirAll,
		openFile: func(name string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			return os.OpenFile(name, flag, perm)
		},
		readDir:  os.ReadDir,
		rename:   os.Rename,
		remove:   os.Remove,
		stat:     os.Stat,
	}
}
