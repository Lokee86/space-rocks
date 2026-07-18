package measurement

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func (clock *testClock) Now() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.now
}

func (clock *testClock) Advance(delta time.Duration) {
	clock.mu.Lock()
	clock.now = clock.now.Add(delta)
	clock.mu.Unlock()
}

func TestRunAggregatesTicksAndPackets(t *testing.T) {
	clock := &testClock{now: time.Unix(100, 0)}
	run := NewRun(RunContext{RunID: "run/one"}, WithClock(clock.Now), WithSampleInterval(time.Hour))
	run.ObserveSimulation(4*time.Millisecond, EntityCounts{Players: 2, AsteroidsSpawnedTotal: 7})
	run.ObserveSimulation(8*time.Millisecond, EntityCounts{Players: 3, AsteroidsSpawnedTotal: 8})
	run.ObservePacketWrite("world", "world_delta", 20)
	run.ObservePacketWrite("world", "world_delta", 40)

	report := run.Snapshot()
	if report.Ticks != (TickSummary{Count: 2, Minimum: 4 * time.Millisecond, Maximum: 8 * time.Millisecond, Total: 12 * time.Millisecond, Average: 6 * time.Millisecond}) {
		t.Fatalf("unexpected tick summary: %#v", report.Ticks)
	}
	if len(report.Packets) != 1 || report.Packets[0].PacketCount != 2 || report.Packets[0].EncodedBytesTotal != 60 || report.Packets[0].MaximumEncodedBytes != 40 {
		t.Fatalf("unexpected packet summary: %#v", report.Packets)
	}
	if report.Context.StartTime != time.Unix(100, 0) || report.Complete {
		t.Fatalf("snapshot should remain active with immutable start context: %#v", report)
	}
}

func TestRunPeriodicSamplesAreBoundedAndCountOverwrites(t *testing.T) {
	clock := &testClock{now: time.Unix(100, 0)}
	process := NewProcessSamplerWithReader(time.Hour, func() ProcessSample {
		return ProcessSample{HeapAllocatedBytes: 10}
	})
	run := NewRun(RunContext{}, WithClock(clock.Now), WithSampleInterval(time.Second), WithSampleCapacity(2), WithProcessSampler(process), WithRoomCountProvider(func() int { return 4 }))
	for i := 1; i <= 4; i++ {
		clock.Advance(time.Second)
		run.ObserveSimulation(time.Duration(i)*time.Millisecond, EntityCounts{Players: i})
	}

	report := run.Snapshot()
	if len(report.Samples) != 2 || report.OverwrittenSampleCount != 2 {
		t.Fatalf("expected a two-entry ring with two overwrites, got %d samples and %d overwrites", len(report.Samples), report.OverwrittenSampleCount)
	}
	if report.Samples[0].Entities.Players != 3 || report.Samples[1].Entities.Players != 4 || report.Samples[0].RoomCount != 4 {
		t.Fatalf("unexpected retained samples: %#v", report.Samples)
	}
}

func TestProcessSamplerCachesAcrossRuns(t *testing.T) {
	reads := 0
	sampler := NewProcessSamplerWithReader(time.Second, func() ProcessSample {
		reads++
		return ProcessSample{Goroutines: reads}
	})
	now := time.Unix(100, 0)
	first := sampler.Sample(now)
	second := sampler.Sample(now.Add(500 * time.Millisecond))
	third := sampler.Sample(now.Add(time.Second))
	if reads != 2 || first.Goroutines != second.Goroutines || third.Goroutines != 2 {
		t.Fatalf("expected cached process sample across runs, reads=%d samples=%#v/%#v/%#v", reads, first, second, third)
	}
}

func TestRunFinalizationIsIdempotentAndIgnoresLaterObservations(t *testing.T) {
	clock := &testClock{now: time.Unix(100, 0)}
	run := NewRun(RunContext{}, WithClock(clock.Now))
	run.ObserveSimulation(time.Millisecond, EntityCounts{})
	first := run.Finalize(StopReasonDisconnected)
	run.ObserveSimulation(time.Second, EntityCounts{Players: 99})
	second := run.Finalize(StopReasonComplete)
	if run.IsActive() || first.StopReason != StopReasonDisconnected || second.StopReason != first.StopReason || second.Ticks != first.Ticks {
		t.Fatalf("finalization should freeze one report: first=%#v second=%#v active=%v", first, second, run.IsActive())
	}
}

func TestRunFinalizationCapturesLatestShortRunState(t *testing.T) {
	clock := &testClock{now: time.Unix(100, 0)}
	process := NewProcessSamplerWithReader(time.Hour, func() ProcessSample {
		return ProcessSample{HeapAllocatedBytes: 42}
	})
	run := NewRun(RunContext{}, WithClock(clock.Now), WithSampleInterval(time.Second), WithProcessSampler(process))
	clock.Advance(100 * time.Millisecond)
	run.ObserveSimulation(time.Millisecond, EntityCounts{Players: 3, Asteroids: 5})

	report := run.Finalize(StopReasonComplete)
	if len(report.Samples) != 1 {
		t.Fatalf("expected final report to include one latest sample, got %#v", report.Samples)
	}
	if report.Samples[0].Entities.Players != 3 || report.Samples[0].Entities.Asteroids != 5 {
		t.Fatalf("final sample lost latest entity state: %#v", report.Samples[0])
	}
	if report.Samples[0].Process.HeapAllocatedBytes != 42 {
		t.Fatalf("final sample lost process state: %#v", report.Samples[0].Process)
	}
}

func TestReportWriterPublishesVersionedJSON(t *testing.T) {
	directory, err := os.MkdirTemp(".", ".measurement-test-")
	if err != nil {
		t.Fatalf("create test report directory: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	report := ServerReport{Context: RunContext{RunID: "unsafe/run"}, Complete: true, StopReason: StopReasonComplete}
	path, err := WriteReport(directory, report)
	if err != nil {
		t.Fatalf("write report: %v", err)
	}
	if filepath.Base(path) == "unsafe/run" || filepath.Ext(path) != ".json" {
		t.Fatalf("report path was not sanitized/versioned: %q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var decoded ServerReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if decoded.Version != ReportVersion || !decoded.Complete {
		t.Fatalf("unexpected decoded report: %#v", decoded)
	}
}
