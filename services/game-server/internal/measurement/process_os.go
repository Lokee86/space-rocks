package measurement

type osProcessSample struct {
	ResidentSetBytes     int64
	PeakResidentSetBytes int64
	CPUUserTimeNanos     int64
	CPUSystemTimeNanos   int64
}
