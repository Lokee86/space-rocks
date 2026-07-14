package servicelog

import "sync"

// Status is the small operational state exposed by Runtime.
type Status struct {
	Closed           bool
	FileEnabled      bool
	FileDegraded     bool
	FileFailureCount int64
	ActivePath       string
	ActiveBytes      int64
}

type statusTracker struct {
	mu     sync.RWMutex
	status Status
}

func (s *statusTracker) snapshot() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *statusTracker) setFileEnabled(enabled bool) {
	s.mu.Lock()
	s.status.FileEnabled = enabled
	s.mu.Unlock()
}

func (s *statusTracker) setActive(path string, bytes int64) {
	s.mu.Lock()
	s.status.ActivePath = path
	s.status.ActiveBytes = bytes
	s.mu.Unlock()
}

func (s *statusTracker) addFailure() {
	s.mu.Lock()
	s.status.FileDegraded = true
	s.status.FileFailureCount++
	s.mu.Unlock()
}
