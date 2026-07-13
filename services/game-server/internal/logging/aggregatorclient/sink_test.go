package aggregatorclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type fakeSender struct {
	mu                    sync.Mutex
	batches               [][]byte
	err                   error
	block                 <-chan struct{}
	started               chan<- struct{}
	canceled              chan<- struct{}
	holdAfterCancellation <-chan struct{}
}

func (f *fakeSender) send(ctx context.Context, batch []byte) error {
	if f.started != nil {
		f.started <- struct{}{}
	}
	if f.block != nil {
		select {
		case <-f.block:
		case <-ctx.Done():
			if f.canceled != nil {
				f.canceled <- struct{}{}
			}
			if f.holdAfterCancellation != nil {
				<-f.holdAfterCancellation
			}
			return ctx.Err()
		}
	}
	f.mu.Lock()
	f.batches = append(f.batches, append([]byte(nil), batch...))
	f.mu.Unlock()
	return f.err
}

func sinkConfig(t *testing.T) Config {
	return Config{
		Enabled: true, EndpointURL: "http://example.test", QueueCapacity: 8,
		BatchSize: 2, FlushInterval: time.Hour, RequestTimeout: time.Second,
		SpoolDirectory: t.TempDir(), SpoolByteCap: 1 << 20, SpoolEnabled: true,
	}
}

func waitFor(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not reached")
}

func (f *fakeSender) batchCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.batches)
}

func (f *fakeSender) payloads() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([][]byte, len(f.batches))
	for i := range f.batches {
		result[i] = append([]byte(nil), f.batches[i]...)
	}
	return result
}

func TestSinkSizeBatching(t *testing.T) {
	s := newSink(sinkConfig(t), &fakeSender{})
	defer s.Close(context.Background())
	s.Write([]byte(`{"a":1}`))
	s.Write([]byte(`{"b":2}`))
	waitFor(t, func() bool { return s.Stats().SentBatches == 1 })
}

func TestSinkIntervalFlush(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 10
	config.FlushInterval = 10 * time.Millisecond
	sender := &fakeSender{}
	s := newSink(config, sender)
	defer s.Close(context.Background())
	s.Write([]byte(`{}`))
	waitFor(t, func() bool { return sender.batchCount() == 1 })
}

func TestSinkQueueFullAndClosedDrops(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	config.QueueCapacity = 1
	block, started := make(chan struct{}), make(chan struct{}, 1)
	sender := &fakeSender{block: block, started: started}
	s := newSink(config, sender)
	s.Write([]byte(`{"a":1}`))
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("send did not start")
	}
	s.Write([]byte(`{"queued":true}`))
	s.Write([]byte(`{"dropped":true}`))
	if s.Stats().QueueDrops != 1 {
		t.Fatalf("queue drops = %d", s.Stats().QueueDrops)
	}
	close(block)
	if err := s.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.Write([]byte(`{}`))
	if s.Stats().ClosedDrops != 1 {
		t.Fatalf("closed drops = %d", s.Stats().ClosedDrops)
	}
}

func TestSinkFailedSendSpoolsBatch(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	s := newSink(config, &fakeSender{err: errors.New("down")})
	s.Write([]byte(`{"x":1}`))
	waitFor(t, func() bool { return s.Stats().SpooledBatches == 1 })
	pending, err := newSpoolStore(config).pending()
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending = %v, %v", pending, err)
	}
	s.Close(context.Background())
}

func TestSinkReplaysOldestSpoolBeforeFresh(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	store := newSpoolStore(config)
	old, _ := encodeBatch([][]byte{[]byte(`{"old":true}`)})
	if _, err := store.save(old); err != nil {
		t.Fatal(err)
	}
	block, started := make(chan struct{}), make(chan struct{}, 2)
	sender := &fakeSender{block: block, started: started}
	s := newSink(config, sender)
	defer s.Close(context.Background())
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("replay did not start")
	}
	s.Write([]byte(`{"fresh":true}`))
	close(block)
	waitFor(t, func() bool { return s.Stats().ReplayedBatches == 1 && s.Stats().SentBatches == 1 })
	payloads := sender.payloads()
	if len(payloads) != 2 {
		t.Fatalf("batches = %d", len(payloads))
	}
	var first, second struct {
		Events []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal(payloads[0], &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(payloads[1], &second); err != nil {
		t.Fatal(err)
	}
	if string(first.Events[0]) != `{"old":true}` || string(second.Events[0]) != `{"fresh":true}` {
		t.Fatalf("wrong replay order")
	}
}

func TestSinkDisabledSpoolDropsBatch(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	config.SpoolEnabled = false
	s := newSink(config, &fakeSender{err: errors.New("down")})
	defer s.Close(context.Background())
	s.Write([]byte(`{}`))
	waitFor(t, func() bool { return s.Stats().SendFailures == 1 })
	if s.Stats().DroppedBatches != 1 || s.Stats().SpooledBatches != 0 {
		t.Fatal("drop accounting incorrect")
	}
}

func TestSinkCloseDrainsQueue(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 10
	s := newSink(config, &fakeSender{})
	s.Write([]byte(`{}`))
	if err := s.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.Stats().SentEvents != 1 {
		t.Fatalf("sent events = %d", s.Stats().SentEvents)
	}
}

func TestSinkWriteAfterCloseBeginsIsDropped(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	block, started := make(chan struct{}), make(chan struct{}, 1)
	sender := &fakeSender{block: block, started: started}
	s := newSink(config, sender)
	s.Write([]byte(`{"first":true}`))
	<-started
	closeDone := make(chan error, 1)
	go func() { closeDone <- s.Close(context.Background()) }()
	waitFor(t, func() bool {
		s.stateMu.Lock()
		closed := s.closed
		s.stateMu.Unlock()
		return closed
	})
	s.Write([]byte(`{"late":true}`))
	if s.Stats().ClosedDrops != 1 {
		t.Fatalf("closed drops = %d", s.Stats().ClosedDrops)
	}
	close(block)
	if err := <-closeDone; err != nil {
		t.Fatal(err)
	}
	for _, payload := range sender.payloads() {
		if bytes.Contains(payload, []byte(`late`)) {
			t.Fatal("late event was delivered")
		}
	}
}

func TestSinkCloseTimeoutCancelsAndSecondCloseWaits(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	block, started := make(chan struct{}), make(chan struct{}, 1)
	canceled, hold := make(chan struct{}, 1), make(chan struct{})
	sender := &fakeSender{block: block, started: started, canceled: canceled, holdAfterCancellation: hold}
	s := newSink(config, sender)
	s.Write([]byte(`{}`))
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := s.Close(ctx); err == nil {
		t.Fatal("expected timeout")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("cancellation not observed")
	}
	secondDone := make(chan error, 1)
	go func() { secondDone <- s.Close(context.Background()) }()
	select {
	case <-secondDone:
		t.Fatal("second Close returned early")
	case <-time.After(20 * time.Millisecond):
	}
	close(hold)
	if err := <-secondDone; err != nil {
		t.Fatal(err)
	}
}

func TestSinkReplaySendFailureRetainsSpool(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	store := newSpoolStore(config)
	payload, _ := encodeBatch([][]byte{[]byte(`{"old":true}`)})
	_, _ = store.save(payload)
	s := newSink(config, &fakeSender{err: errors.New("down")})
	waitFor(t, func() bool { return s.Stats().SendFailures == 1 })
	pending, err := store.pending()
	if err != nil || len(pending) != 1 {
		t.Fatalf("spool removed: %v, %v", pending, err)
	}
	s.Close(context.Background())
}

func TestSinkSpoolEvictionAccounting(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	firstEvent := []byte(`{"same":1}`)
	secondEvent := []byte(`{"same":2}`)
	first, _ := encodeBatch([][]byte{firstEvent})
	config.SpoolByteCap = int64(len(first))
	sink := newSink(config, &fakeSender{err: errors.New("down")})
	sink.Write(firstEvent)
	waitFor(t, func() bool { return sink.Stats().SpooledBatches == 1 })
	sink.Write(secondEvent)
	waitFor(t, func() bool { return sink.Stats().SpoolEvictedBatches == 1 })
	stats := sink.Stats()
	if stats.SpooledBatches != 2 || stats.SpooledEvents != 2 || stats.SpoolEvictedEvents != 1 || stats.DroppedBatches != 1 || stats.DroppedEvents != 1 {
		t.Fatalf("stats = %#v", stats)
	}
	sink.Close(context.Background())
}

func TestSinkReplayDiscardsMalformedOldestAndContinues(t *testing.T) {
	config := sinkConfig(t)
	config.BatchSize = 1
	store := newSpoolStore(config)
	poisonPath := filepath.Join(store.directory, spoolFilePrefix+"0000"+spoolFileSuffix)
	if err := os.WriteFile(poisonPath, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	valid, _ := encodeBatch([][]byte{[]byte(`{"valid":true}`)})
	if _, err := store.save(valid); err != nil {
		t.Fatal(err)
	}
	sender := &fakeSender{}
	sink := newSink(config, sender)
	waitFor(t, func() bool { stats := sink.Stats(); return stats.SpoolFailures == 1 && stats.ReplayedBatches == 1 })
	if _, err := os.Stat(poisonPath); !os.IsNotExist(err) {
		t.Fatalf("poison spool remains: %v", err)
	}
	stats := sink.Stats()
	if stats.DroppedBatches != 1 || stats.DroppedEvents != 0 || stats.ReplayedEvents != 1 {
		t.Fatalf("stats = %#v", stats)
	}
	if err := sink.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSinkReplayPendingFailureCountsSpoolFailure(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "not-a-directory")
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	config := sinkConfig(t)
	config.SpoolDirectory = file.Name()
	s := newSink(config, &fakeSender{})
	waitFor(t, func() bool { return s.Stats().SpoolFailures >= 1 })
	if err := s.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}
