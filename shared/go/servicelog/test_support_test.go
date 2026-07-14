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

func (c *fakeClock) advance(delta time.Duration) {
	c.current = c.current.Add(delta)
}

type trackedFile struct {
	*os.File
	closeCount int
	closeErr   error
}

func (f *trackedFile) Close() error {
	f.closeCount++
	if f.File == nil {
		return f.closeErr
	}
	err := f.File.Close()
	if f.closeErr != nil {
		return f.closeErr
	}
	return err
}

type fakeFilesystem struct {
	mkdir    func(string, fs.FileMode) error
	openFile func(string, int, fs.FileMode) (io.WriteCloser, error)
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
		readDir:       f.readDir,
		rename:        f.rename,
		remove:        f.remove,
		stat:          f.stat,
	}
}
