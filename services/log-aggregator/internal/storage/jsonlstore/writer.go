package jsonlstore

import (
	"bufio"
	"errors"
	"os"
	"sync"
	"time"
)

type writer struct {
	file     *os.File
	buffer   *bufio.Writer
	mu       sync.Mutex
	stop     chan struct{}
	done     chan struct{}
	once     sync.Once
	closed   bool
	closeErr error
}

func newWriter(file *os.File, interval time.Duration) *writer {
	result := &writer{
		file: file, buffer: bufio.NewWriter(file),
		stop: make(chan struct{}), done: make(chan struct{}),
	}
	if interval > 0 {
		go result.flushLoop(interval)
	} else {
		close(result.done)
	}
	return result
}

func (writer *writer) write(data []byte) error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.closed {
		return errors.New("jsonlstore: writer is closed")
	}
	_, err := writer.buffer.Write(data)
	return err
}

func (writer *writer) flush() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.closed {
		return errors.New("jsonlstore: writer is closed")
	}
	return writer.buffer.Flush()
}

func (writer *writer) durableFlush() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.closed {
		return errors.New("jsonlstore: writer is closed")
	}
	if err := writer.buffer.Flush(); err != nil {
		return err
	}
	return writer.file.Sync()
}

func (writer *writer) close() error {
	writer.once.Do(func() {
		close(writer.stop)
		<-writer.done
		writer.mu.Lock()
		defer writer.mu.Unlock()
		writer.closed = true
		if err := writer.buffer.Flush(); err != nil {
			writer.closeErr = err
		}
		if writer.closeErr == nil {
			writer.closeErr = writer.file.Sync()
		}
		if err := writer.file.Close(); writer.closeErr == nil {
			writer.closeErr = err
		}
	})
	return writer.closeErr
}

func (writer *writer) flushLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(writer.done)
	for {
		select {
		case <-ticker.C:
			_ = writer.flush()
		case <-writer.stop:
			return
		}
	}
}
