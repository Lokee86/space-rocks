package outbound

import (
	"errors"
	"testing"
	"time"
)

type testWriter struct {
	deadline    time.Time
	deadlineErr error
	writeCalled bool
	deadlineSet bool
}

func (w *testWriter) SetWriteDeadline(deadline time.Time) error {
	w.deadline = deadline
	w.deadlineSet = true
	return w.deadlineErr
}

func (w *testWriter) WriteMessage(int, []byte) error {
	if !w.deadlineSet {
		return errors.New("write occurred before deadline")
	}
	w.writeCalled = true
	return nil
}

func TestWriteServerMessageSetsDeadlineBeforeWrite(t *testing.T) {
	w := &testWriter{}
	if !writeServerMessage(w, []byte("x"), nil) {
		t.Fatal("expected write to succeed")
	}
	if !w.deadline.After(time.Now()) {
		t.Fatal("expected future write deadline")
	}
	if !w.writeCalled {
		t.Fatal("expected write")
	}
}

func TestWriteServerMessageDeadlineFailure(t *testing.T) {
	w := &testWriter{deadlineErr: errors.New("deadline")}
	var callbackErr error
	if writeServerMessage(w, nil, func(err error) { callbackErr = err }) {
		t.Fatal("expected deadline failure")
	}
	if callbackErr == nil {
		t.Fatal("expected close callback")
	}
	if w.writeCalled {
		t.Fatal("did not expect write after deadline failure")
	}
}
