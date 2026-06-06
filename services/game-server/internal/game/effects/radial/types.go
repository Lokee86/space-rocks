package radial

type CoverageMode string

const (
	CoverageAnnularWave   CoverageMode = "annular_wave"
	CoverageExpandingFill CoverageMode = "expanding_fill"
)

type ExpirationMode string

const (
	ExpirationSimultaneous ExpirationMode = "simultaneous"
	ExpirationSequential    ExpirationMode = "sequential"
)
