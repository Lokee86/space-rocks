//go:build windows

package measurement

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var getProcessMemoryInfo = windows.NewLazySystemDLL("psapi.dll").NewProc("GetProcessMemoryInfo")

type processMemoryCounters struct {
	Size                       uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

func readOSProcessSample() osProcessSample {
	var sample osProcessSample
	handle := windows.CurrentProcess()
	var creationTime windows.Filetime
	var exitTime windows.Filetime
	var kernelTime windows.Filetime
	var userTime windows.Filetime
	if windows.GetProcessTimes(handle, &creationTime, &exitTime, &kernelTime, &userTime) == nil {
		sample.CPUUserTimeNanos = filetimeDurationNanos(userTime)
		sample.CPUSystemTimeNanos = filetimeDurationNanos(kernelTime)
	}

	counters := processMemoryCounters{Size: uint32(unsafe.Sizeof(processMemoryCounters{}))}
	result, _, _ := getProcessMemoryInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&counters)),
		uintptr(counters.Size),
	)
	if result != 0 {
		sample.ResidentSetBytes = int64(counters.WorkingSetSize)
		sample.PeakResidentSetBytes = int64(counters.PeakWorkingSetSize)
	}
	return sample
}

func filetimeDurationNanos(value windows.Filetime) int64 {
	ticks := uint64(value.HighDateTime)<<32 | uint64(value.LowDateTime)
	return int64(ticks) * 100
}
