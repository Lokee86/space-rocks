package jsonlstore

import "time"

// Clock supplies time to the rolling store and keeps time-dependent behavior
// deterministic in tests.
type Clock interface {
	Now() time.Time
}

// RealClock reads the process clock.
type RealClock struct{}

// Now returns the current wall-clock time.
func (RealClock) Now() time.Time {
	return time.Now()
}
