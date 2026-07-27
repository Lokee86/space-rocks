//go:build !linux && !windows && !darwin

package measurement

func readOSProcessSample() osProcessSample {
	return osProcessSample{}
}
