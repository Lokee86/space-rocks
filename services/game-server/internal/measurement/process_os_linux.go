//go:build linux

package measurement

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func readOSProcessSample() osProcessSample {
	var sample osProcessSample
	if usage := readLinuxRusage(); usage != nil {
		sample.CPUUserTimeNanos = unix.TimevalToNsec(usage.Utime)
		sample.CPUSystemTimeNanos = unix.TimevalToNsec(usage.Stime)
		sample.PeakResidentSetBytes = usage.Maxrss * 1024
	}
	currentRSS, peakRSS := readLinuxProcMemory()
	if currentRSS > 0 {
		sample.ResidentSetBytes = currentRSS
	}
	if peakRSS > sample.PeakResidentSetBytes {
		sample.PeakResidentSetBytes = peakRSS
	}
	return sample
}

func readLinuxRusage() *unix.Rusage {
	var usage unix.Rusage
	if unix.Getrusage(unix.RUSAGE_SELF, &usage) != nil {
		return nil
	}
	return &usage
}

func readLinuxProcMemory() (int64, int64) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, 0
	}
	var currentRSS int64
	var peakRSS int64
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(line, "VmRSS:"):
			currentRSS = parseLinuxMemoryKilobytes(line)
		case strings.HasPrefix(line, "VmHWM:"):
			peakRSS = parseLinuxMemoryKilobytes(line)
		}
	}
	return currentRSS, peakRSS
}

func parseLinuxMemoryKilobytes(line string) int64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	value, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value * 1024
}
