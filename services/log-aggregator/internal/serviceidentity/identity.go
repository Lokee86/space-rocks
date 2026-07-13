// Package serviceidentity defines the diagnostic aggregator service identity.
package serviceidentity

const (
	ServiceName         = "diagnostic-aggregator"
	EnvPrefix           = "DIAGNOSTIC_AGGREGATOR_"
	LegacyEnvPrefix     = "LOG_AGGREGATOR_"
	DefaultLogDirectory = "logs/diagnostic-aggregator"
)