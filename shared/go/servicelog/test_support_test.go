package servicelog

import (
	"io"
	"io/fs"
	"os"
	"time"
)

type fakeClock struct {
	current time.Time
}

func (c *fakeClock) now() time.Time {
	return c.current
}

type fakeFilesystem struct {
	mkdir    func(string, fs.FileMode) error
	openFile func(string, int, fs.FileMode) (*os.File, error)
	readFile func(string) ([]byte, error)
	readDir  func(string) ([]os.DirEntry, error)
	rename   func(string, string) error
	remove   func(string) error
	stat     func(string) (os.FileInfo, error)
}

func (f fakeFilesystem) dependencies(writer io.Writer, clock *fakeClock) runtimeDependencies {
	return runtimeDependencies{
		consoleWriter: writer,
		now:           clock.now,
		mkdir:         f.mkdir,
		openFile:      f.openFile,
		readFile:      f.readFile,
		readDir:       f.readDir,
		rename:        f.rename,
		remove:        f.remove,
		stat:          f.stat,
	}
}
