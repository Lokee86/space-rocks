module github.com/Lokee86/space-rocks/services/diagnostic-aggregator

go 1.26.3

require (
	github.com/Lokee86/space-rocks/shared/go/observabilityevent v0.0.0
	github.com/Lokee86/space-rocks/shared/go/servicelog v0.0.0
	github.com/google/uuid v1.6.0
)

replace github.com/Lokee86/space-rocks/shared/go/servicelog => ../../shared/go/servicelog

replace github.com/Lokee86/space-rocks/shared/go/observabilityevent => ../../shared/go/observabilityevent
