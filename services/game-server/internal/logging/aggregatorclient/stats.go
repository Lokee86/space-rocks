package aggregatorclient

import "sync/atomic"

type Stats struct {
	Enqueued, QueueDrops, ClosedDrops, InvalidEvents  uint64
	DroppedBatches, DroppedEvents                     uint64
	SentBatches, SentEvents, SendFailures             uint64
	SpooledBatches, SpooledEvents, SpoolFailures      uint64
	SpoolEvictedBatches, SpoolEvictedEvents           uint64
	ReplayedBatches, ReplayedEvents, ShutdownTimeouts uint64
}

type statsCounters struct {
	enqueued, queueDrops, closedDrops, invalidEvents  atomic.Uint64
	droppedBatches, droppedEvents                     atomic.Uint64
	sentBatches, sentEvents, sendFailures             atomic.Uint64
	spooledBatches, spooledEvents, spoolFailures      atomic.Uint64
	spoolEvictedBatches, spoolEvictedEvents           atomic.Uint64
	replayedBatches, replayedEvents, shutdownTimeouts atomic.Uint64
}

func (s *statsCounters) snapshot() Stats {
	return Stats{
		Enqueued: s.enqueued.Load(), QueueDrops: s.queueDrops.Load(), ClosedDrops: s.closedDrops.Load(), InvalidEvents: s.invalidEvents.Load(),
		DroppedBatches: s.droppedBatches.Load(), DroppedEvents: s.droppedEvents.Load(),
		SentBatches: s.sentBatches.Load(), SentEvents: s.sentEvents.Load(), SendFailures: s.sendFailures.Load(),
		SpooledBatches: s.spooledBatches.Load(), SpooledEvents: s.spooledEvents.Load(), SpoolFailures: s.spoolFailures.Load(),
		SpoolEvictedBatches: s.spoolEvictedBatches.Load(), SpoolEvictedEvents: s.spoolEvictedEvents.Load(),
		ReplayedBatches: s.replayedBatches.Load(), ReplayedEvents: s.replayedEvents.Load(), ShutdownTimeouts: s.shutdownTimeouts.Load(),
	}
}
