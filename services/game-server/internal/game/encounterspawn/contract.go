package encounterspawn

import (
	"fmt"
	"math"
)

type ProfileID string

const ProfilePlayercentricAsteroidsV1 ProfileID = "playercentric_asteroids_v1"

type ScheduleKind string

const (
	ScheduleContinuous ScheduleKind = "continuous"
	ScheduleWave       ScheduleKind = "wave"
	ScheduleEvent      ScheduleKind = "event"
	ScheduleObjective  ScheduleKind = "objective"
	ScheduleScripted   ScheduleKind = "scripted"
)

type Priority int
type WeightedPopulation int

type Config struct {
	ID                      ProfileID
	ScheduleKind            ScheduleKind
	IntervalSeconds         float64
	BatchSize               int
	Priority                Priority
	SharedWeightedBudget    WeightedPopulation
	ProfileWeightedLimit    WeightedPopulation
	SpawnTypeWeightedLimits map[string]WeightedPopulation
	RetryCap                int
	InitiallyActive         bool
}

func (config Config) Validate() error {
	if config.ID == "" {
		return fmt.Errorf("encounter spawn profile ID is required")
	}
	if !supportedScheduleKind(config.ScheduleKind) {
		return fmt.Errorf("unsupported encounter spawn schedule kind %q", config.ScheduleKind)
	}
	if math.IsNaN(config.IntervalSeconds) || math.IsInf(config.IntervalSeconds, 0) || config.IntervalSeconds < 0 {
		return fmt.Errorf("encounter spawn interval must be finite and non-negative")
	}
	if config.ScheduleKind == ScheduleContinuous && config.IntervalSeconds <= 0 {
		return fmt.Errorf("continuous encounter spawn interval must be positive")
	}
	if config.BatchSize <= 0 {
		return fmt.Errorf("encounter spawn batch size must be positive")
	}
	if config.Priority < 0 {
		return fmt.Errorf("encounter spawn priority cannot be negative")
	}
	if config.SharedWeightedBudget < 0 || config.ProfileWeightedLimit < 0 {
		return fmt.Errorf("encounter spawn population limits cannot be negative")
	}
	if config.RetryCap < 0 {
		return fmt.Errorf("encounter spawn retry cap cannot be negative")
	}
	for spawnType, limit := range config.SpawnTypeWeightedLimits {
		if spawnType == "" {
			return fmt.Errorf("encounter spawn type limit requires a spawn type")
		}
		if limit < 0 {
			return fmt.Errorf("encounter spawn type limit for %q cannot be negative", spawnType)
		}
	}
	return nil
}

func (config Config) Clone() Config {
	clone := config
	if config.SpawnTypeWeightedLimits != nil {
		clone.SpawnTypeWeightedLimits = make(map[string]WeightedPopulation, len(config.SpawnTypeWeightedLimits))
		for spawnType, limit := range config.SpawnTypeWeightedLimits {
			clone.SpawnTypeWeightedLimits[spawnType] = limit
		}
	}
	return clone
}

func supportedScheduleKind(kind ScheduleKind) bool {
	switch kind {
	case ScheduleContinuous, ScheduleWave, ScheduleEvent, ScheduleObjective, ScheduleScripted:
		return true
	default:
		return false
	}
}
