package aggregatorclient

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type Sink struct {
	queue    chan []byte
	sender   batchSender
	spool    *spoolStore
	config   Config
	stats    statsCounters
	stop     chan struct{}
	finished chan struct{}
	cancel   context.CancelFunc
	ctx      context.Context
	stateMu  sync.Mutex
	closed   bool
}

func New(config Config) (*Sink, error) {
	if !config.Enabled {
		return nil, nil
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return newSink(config, newHTTPBatchSender(config)), nil
}

func newSink(config Config, sender batchSender) *Sink {
	ctx, cancel := context.WithCancel(context.Background())
	sink := &Sink{
		queue:    make(chan []byte, config.QueueCapacity),
		sender:   sender,
		config:   config,
		stop:     make(chan struct{}),
		finished: make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}
	if config.SpoolEnabled {
		sink.spool = newSpoolStore(config)
	}
	go sink.run()
	return sink
}

func (s *Sink) Write(event []byte) (int, error) {
	if !json.Valid(event) {
		s.stats.invalidEvents.Add(1)
		return len(event), nil
	}
	copyOfEvent := append([]byte(nil), event...)
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.closed {
		s.stats.closedDrops.Add(1)
		return len(event), nil
	}
	select {
	case s.queue <- copyOfEvent:
		s.stats.enqueued.Add(1)
	default:
		s.stats.queueDrops.Add(1)
	}
	return len(event), nil
}

func (s *Sink) Stats() Stats {
	return s.stats.snapshot()
}

func (s *Sink) Close(ctx context.Context) error {
	s.stateMu.Lock()
	if !s.closed {
		s.closed = true
		close(s.stop)
	}
	s.stateMu.Unlock()

	select {
	case <-s.finished:
		return nil
	case <-ctx.Done():
		s.stats.shutdownTimeouts.Add(1)
		s.cancel()
		return ctx.Err()
	}
}

func (s *Sink) run() {
	defer close(s.finished)
	defer s.cancel()

	s.replay()
	ticker := time.NewTicker(s.config.FlushInterval)
	defer ticker.Stop()
	batch := make([][]byte, 0, s.config.BatchSize)
	flush := func(replayFirst bool) {
		if len(batch) == 0 {
			return
		}
		if replayFirst {
			s.replay()
		}
		s.deliver(batch)
		batch = batch[:0]
	}

	for {
		select {
		case event := <-s.queue:
			batch = append(batch, event)
			if len(batch) >= s.config.BatchSize {
				flush(true)
			}
		case <-ticker.C:
			s.replay()
			flush(false)
		case <-s.stop:
			for {
				select {
				case event := <-s.queue:
					batch = append(batch, event)
					if len(batch) >= s.config.BatchSize {
						flush(true)
					}
				default:
					flush(true)
					return
				}
			}
		}
	}
}

func (s *Sink) deliver(events [][]byte) {
	encoded, err := encodeBatch(events)
	if err != nil {
		s.stats.invalidEvents.Add(uint64(len(events)))
		s.stats.droppedBatches.Add(1)
		s.stats.droppedEvents.Add(uint64(len(events)))
		return
	}
	if err = s.sender.send(s.ctx, encoded); err == nil {
		s.stats.sentBatches.Add(1)
		s.stats.sentEvents.Add(uint64(len(events)))
		return
	}
	s.stats.sendFailures.Add(1)
	if s.spool == nil {
		s.stats.droppedBatches.Add(1)
		s.stats.droppedEvents.Add(uint64(len(events)))
		return
	}
	result, saveErr := s.spool.save(encoded)
	if result.EvictedBatches > 0 {
		s.stats.spoolEvictedBatches.Add(uint64(result.EvictedBatches))
		s.stats.spoolEvictedEvents.Add(uint64(result.EvictedEvents))
		s.stats.droppedBatches.Add(uint64(result.EvictedBatches))
		s.stats.droppedEvents.Add(uint64(result.EvictedEvents))
	}
	if saveErr != nil {
		s.stats.spoolFailures.Add(1)
	}
	if !result.Stored {
		s.stats.droppedBatches.Add(1)
		s.stats.droppedEvents.Add(uint64(len(events)))
		return
	}
	s.stats.spooledBatches.Add(1)
	s.stats.spooledEvents.Add(uint64(len(events)))
}

func (s *Sink) replay() {
	if s.spool == nil {
		return
	}
	pending, err := s.spool.pending()
	if err != nil {
		s.stats.spoolFailures.Add(1)
		return
	}
	for _, batch := range pending {
		payload, eventCount, loadErr := s.spool.load(batch)
		if loadErr != nil {
			s.stats.spoolFailures.Add(1)
			if removeErr := s.spool.remove(batch); removeErr != nil {
				s.stats.spoolFailures.Add(1)
				return
			}
			s.stats.droppedBatches.Add(1)
			s.stats.droppedEvents.Add(uint64(eventCount))
			continue
		}
		if err = s.sender.send(s.ctx, payload); err != nil {
			s.stats.sendFailures.Add(1)
			return
		}
		if err = s.spool.remove(batch); err != nil {
			s.stats.spoolFailures.Add(1)
			return
		}
		s.stats.replayedBatches.Add(1)
		s.stats.replayedEvents.Add(uint64(eventCount))
	}
}
