package measurement

import (
	"runtime"
	"sync"
	"time"
)

type ProcessSampler struct {
	mu       sync.Mutex
	interval time.Duration
	lastAt   time.Time
	sample   ProcessSample
	hasValue bool
	read     func() ProcessSample
}

func NewProcessSampler(interval ...time.Duration) *ProcessSampler {
	value := time.Second
	if len(interval) > 0 && interval[0] > 0 {
		value = interval[0]
	}
	return NewProcessSamplerWithReader(value, readProcessSample)
}

func NewProcessSamplerWithReader(interval time.Duration, reader func() ProcessSample) *ProcessSampler {
	if interval <= 0 {
		interval = time.Second
	}
	if reader == nil {
		reader = readProcessSample
	}
	return &ProcessSampler{interval: interval, read: reader}
}

func (sampler *ProcessSampler) Sample(now time.Time) ProcessSample {
	sampler.mu.Lock()
	defer sampler.mu.Unlock()
	if sampler.hasValue && !now.Before(sampler.lastAt) && now.Sub(sampler.lastAt) < sampler.interval {
		return sampler.sample
	}

	next := sampler.read()
	if sampler.hasValue && now.After(sampler.lastAt) {
		previousCPU := sampler.sample.CPUUserTimeNanos + sampler.sample.CPUSystemTimeNanos
		currentCPU := next.CPUUserTimeNanos + next.CPUSystemTimeNanos
		if currentCPU >= previousCPU {
			next.CPUUtilizationCores = float64(currentCPU-previousCPU) / float64(now.Sub(sampler.lastAt))
		}
		if next.GCPauseTotalNanos >= sampler.sample.GCPauseTotalNanos {
			next.GCPauseWindowNanos = next.GCPauseTotalNanos - sampler.sample.GCPauseTotalNanos
		}
	}

	sampler.sample = next
	sampler.lastAt = now
	sampler.hasValue = true
	return sampler.sample
}

var defaultProcessSampler = NewProcessSampler()

func readProcessSample() ProcessSample {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	process := readOSProcessSample()
	lastPauseNanos := uint64(0)
	if stats.NumGC > 0 {
		lastPauseNanos = stats.PauseNs[(stats.NumGC-1)%uint32(len(stats.PauseNs))]
	}
	return ProcessSample{
		HeapAllocatedBytes:   int64(stats.HeapAlloc),
		HeapInUseBytes:       int64(stats.HeapInuse),
		SystemBytes:          int64(stats.Sys),
		ResidentSetBytes:     process.ResidentSetBytes,
		PeakResidentSetBytes: process.PeakResidentSetBytes,
		CPUUserTimeNanos:     process.CPUUserTimeNanos,
		CPUSystemTimeNanos:   process.CPUSystemTimeNanos,
		Goroutines:           runtime.NumGoroutine(),
		GCCycles:             stats.NumGC,
		GCPauseTotalNanos:    stats.PauseTotalNs,
		GCLastPauseNanos:     lastPauseNanos,
	}
}
