//go:build windows

package measurement

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestFiletimeDurationNanosTreatsProcessTimeAsDuration(t *testing.T) {
	value := windows.Filetime{LowDateTime: 15_000_000}
	if got := filetimeDurationNanos(value); got != 1_500_000_000 {
		t.Fatalf("duration nanos = %d, want 1500000000", got)
	}
}

func TestReadOSProcessSampleReportsWindowsProcessMetrics(t *testing.T) {
	sample := readOSProcessSample()
	if sample.ResidentSetBytes <= 0 || sample.PeakResidentSetBytes < sample.ResidentSetBytes {
		t.Fatalf("unexpected resident memory sample: %#v", sample)
	}
	if sample.CPUUserTimeNanos < 0 || sample.CPUSystemTimeNanos < 0 {
		t.Fatalf("unexpected CPU time sample: %#v", sample)
	}
}
