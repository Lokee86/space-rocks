package encounterspawn

import (
	"math"
	"testing"
)

func validConfig() Config {
	return Config{
		ID:                      ProfilePlayercentricAsteroidsV1,
		ScheduleKind:            ScheduleContinuous,
		IntervalSeconds:         3,
		BatchSize:               3,
		Priority:                1,
		SharedWeightedBudget:    100,
		ProfileWeightedLimit:    80,
		SpawnTypeWeightedLimits: map[string]WeightedPopulation{"asteroid": 60},
		RetryCap:                8,
		InitiallyActive:         true,
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"missing ID", func(config *Config) { config.ID = "" }},
		{"unsupported schedule", func(config *Config) { config.ScheduleKind = "unknown" }},
		{"zero continuous interval", func(config *Config) { config.IntervalSeconds = 0 }},
		{"non-finite interval", func(config *Config) { config.IntervalSeconds = math.Inf(1) }},
		{"zero batch", func(config *Config) { config.BatchSize = 0 }},
		{"negative priority", func(config *Config) { config.Priority = -1 }},
		{"negative shared budget", func(config *Config) { config.SharedWeightedBudget = -1 }},
		{"negative profile limit", func(config *Config) { config.ProfileWeightedLimit = -1 }},
		{"negative retry cap", func(config *Config) { config.RetryCap = -1 }},
		{"empty spawn type", func(config *Config) { config.SpawnTypeWeightedLimits = map[string]WeightedPopulation{"": 1} }},
		{"negative spawn limit", func(config *Config) { config.SpawnTypeWeightedLimits = map[string]WeightedPopulation{"asteroid": -1} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validConfig()
			test.mutate(&config)
			if err := config.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestConfigSupportsNonContinuousScheduleContracts(t *testing.T) {
	for _, kind := range []ScheduleKind{ScheduleWave, ScheduleEvent, ScheduleObjective, ScheduleScripted} {
		config := validConfig()
		config.ScheduleKind = kind
		config.IntervalSeconds = 0
		if err := config.Validate(); err != nil {
			t.Fatalf("schedule %q rejected: %v", kind, err)
		}
	}
}

func TestConfigCloneDefensivelyCopiesLimits(t *testing.T) {
	config := validConfig()
	clone := config.Clone()
	clone.SpawnTypeWeightedLimits["asteroid"] = 1
	if config.SpawnTypeWeightedLimits["asteroid"] != 60 {
		t.Fatal("clone mutated source spawn type limits")
	}
}
