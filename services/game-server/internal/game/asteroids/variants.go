package asteroids

import "math/rand"

type Variant struct {
	ID                  string
	Index               int
	CollisionShape      string
	StatsProfile        string
	DropTable           string
	TimedSpawnWeight    float64
	FragmentSpawnWeight float64
	DebugSpawnWeight    float64
}

var Variants = []Variant{
	{
		ID:                  "asteroid_1",
		Index:               0,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
	{
		ID:                  "asteroid_2",
		Index:               1,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
	{
		ID:                  "asteroid_3",
		Index:               2,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
	{
		ID:                  "asteroid_4",
		Index:               3,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
	{
		ID:                  "asteroid_5",
		Index:               4,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
	{
		ID:                  "asteroid_6",
		Index:               5,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
	{
		ID:                  "asteroid_7",
		Index:               6,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
	{
		ID:                  "asteroid_8",
		Index:               7,
		CollisionShape:      "asteroid:0",
		StatsProfile:        "standard",
		DropTable:           "basicasteroids",
		TimedSpawnWeight:    1.0,
		FragmentSpawnWeight: 1.0,
		DebugSpawnWeight:    1.0,
	},
}

func Count() int {
	return len(Variants)
}

func ByIndex(index int) Variant {
	if len(Variants) == 0 {
		return Variant{}
	}

	return Variants[wrapIndex(index, len(Variants))]
}

func TimedSpawnVariants() []Variant {
	return spawnVariants(func(variant Variant) float64 {
		return variant.TimedSpawnWeight
	})
}

func FragmentSpawnVariants() []Variant {
	return spawnVariants(func(variant Variant) float64 {
		return variant.FragmentSpawnWeight
	})
}

func DebugSpawnVariants() []Variant {
	return spawnVariants(func(variant Variant) float64 {
		return variant.DebugSpawnWeight
	})
}

func RandomTimedSpawnVariantIndex() int {
	return randomWeightedVariantIndex(func(variant Variant) float64 {
		return variant.TimedSpawnWeight
	})
}

func RandomFragmentSpawnVariantIndex() int {
	return randomWeightedVariantIndex(func(variant Variant) float64 {
		return variant.FragmentSpawnWeight
	})
}

func RandomDebugSpawnVariantIndex() int {
	return randomWeightedVariantIndex(func(variant Variant) float64 {
		return variant.DebugSpawnWeight
	})
}

func spawnVariants(weightFn func(Variant) float64) []Variant {
	spawnable := make([]Variant, 0, len(Variants))
	for _, variant := range Variants {
		if weightFn(variant) > 0.0 {
			spawnable = append(spawnable, variant)
		}
	}
	return spawnable
}

func randomWeightedVariantIndex(weightFn func(Variant) float64) int {
	var totalWeight float64
	for _, variant := range Variants {
		weight := weightFn(variant)
		if weight > 0.0 {
			totalWeight += weight
		}
	}

	if totalWeight <= 0.0 {
		return 0
	}

	threshold := rand.Float64() * totalWeight
	var cumulativeWeight float64
	for _, variant := range Variants {
		weight := weightFn(variant)
		if weight <= 0.0 {
			continue
		}

		cumulativeWeight += weight
		if threshold < cumulativeWeight {
			return variant.Index
		}
	}

	return Variants[len(Variants)-1].Index
}

func wrapIndex(index int, size int) int {
	if size <= 0 {
		return 0
	}

	wrapped := index % size
	if wrapped < 0 {
		wrapped += size
	}
	return wrapped
}
