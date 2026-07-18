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

	sampler.sample = sampler.read()
	sampler.lastAt = now
	sampler.hasValue = true
	return sampler.sample
}

var defaultProcessSampler = NewProcessSampler()

func readProcessSample() ProcessSample {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return ProcessSample{
		HeapAllocatedBytes: int64(stats.HeapAlloc),
		HeapInUseBytes:     int64(stats.HeapInuse),
		SystemBytes:        int64(stats.Sys),
		Goroutines:         runtime.NumGoroutine(),
		GCCycles:           stats.NumGC,
	}
}
