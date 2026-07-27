//go:build darwin

package measurement

import "golang.org/x/sys/unix"

func readOSProcessSample() osProcessSample {
	var usage unix.Rusage
	if unix.Getrusage(unix.RUSAGE_SELF, &usage) != nil {
		return osProcessSample{}
	}
	return osProcessSample{
		PeakResidentSetBytes: usage.Maxrss,
		CPUUserTimeNanos:     unix.TimevalToNsec(usage.Utime),
		CPUSystemTimeNanos:   unix.TimevalToNsec(usage.Stime),
	}
}
