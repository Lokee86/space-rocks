package damage

import (
	"fmt"
	"math"
	"sort"
)

type DamageOverTimeAddOutcome struct {
	EffectID  string
	Added     bool
	Replaced  bool
	Refreshed bool
	Dropped   bool
}

type DamageOverTimeTick struct {
	EffectID string
	Effect   ActiveDamageOverTime
}

type scheduledDamageOverTime struct {
	id                string
	effect            ActiveDamageOverTime
	tickRemaining     float64
	durationRemaining float64
}

type DamageOverTimeRuntime struct {
	nextID  int
	targets map[string][]scheduledDamageOverTime
}

func NewDamageOverTimeRuntime() *DamageOverTimeRuntime {
	return &DamageOverTimeRuntime{targets: make(map[string][]scheduledDamageOverTime)}
}

func (runtime *DamageOverTimeRuntime) Add(effect ActiveDamageOverTime) DamageOverTimeAddOutcome {
	if !validActiveDamageOverTime(effect) {
		return DamageOverTimeAddOutcome{Dropped: true}
	}
	if runtime.targets == nil {
		runtime.targets = make(map[string][]scheduledDamageOverTime)
	}
	policy := effect.StackingPolicy
	if policy == "" {
		policy = DamageOverTimeStack
	}
	stackKey := damageOverTimeStackKey(effect)
	effects := runtime.targets[effect.Target.EntityID]
	matching := matchingDamageOverTimeIndexes(effects, stackKey)

	switch policy {
	case DamageOverTimeReplace:
		if len(matching) > 0 {
			replacement := runtime.newScheduledEffect(effect)
			kept := make([]scheduledDamageOverTime, 0, len(effects)-len(matching)+1)
			for index, existing := range effects {
				if !containsIndex(matching, index) {
					kept = append(kept, existing)
				}
			}
			runtime.targets[effect.Target.EntityID] = append(kept, replacement)
			return DamageOverTimeAddOutcome{EffectID: replacement.id, Added: true, Replaced: true}
		}
	case DamageOverTimeRefresh:
		if len(matching) > 0 {
			index := matching[0]
			refreshed := runtime.newScheduledEffect(effect)
			refreshed.id = effects[index].id
			effects[index] = refreshed
			runtime.targets[effect.Target.EntityID] = effects
			return DamageOverTimeAddOutcome{EffectID: refreshed.id, Added: true, Refreshed: true}
		}
	case DamageOverTimeLimit:
		limit := effect.MaxStacks
		if limit <= 0 {
			limit = 1
		}
		if len(matching) >= limit {
			return DamageOverTimeAddOutcome{Dropped: true}
		}
	case DamageOverTimeStack:
	default:
		return DamageOverTimeAddOutcome{Dropped: true}
	}

	scheduled := runtime.newScheduledEffect(effect)
	runtime.targets[effect.Target.EntityID] = append(effects, scheduled)
	return DamageOverTimeAddOutcome{EffectID: scheduled.id, Added: true}
}

func (runtime *DamageOverTimeRuntime) Step(delta float64, paused bool) []DamageOverTimeTick {
	if paused || delta <= 0 || math.IsNaN(delta) || math.IsInf(delta, 0) {
		return nil
	}
	targetIDs := make([]string, 0, len(runtime.targets))
	for targetID := range runtime.targets {
		targetIDs = append(targetIDs, targetID)
	}
	sort.Strings(targetIDs)

	ticks := make([]DamageOverTimeTick, 0)
	for _, targetID := range targetIDs {
		effects := runtime.targets[targetID]
		remainingEffects := effects[:0]
		for _, scheduled := range effects {
			remainingDelta := delta
			for remainingDelta > 0 && scheduled.durationRemaining > 0 {
				step := min(remainingDelta, scheduled.tickRemaining, scheduled.durationRemaining)
				remainingDelta -= step
				scheduled.tickRemaining -= step
				scheduled.durationRemaining -= step
				if scheduled.tickRemaining <= 0 {
					ticks = append(ticks, DamageOverTimeTick{EffectID: scheduled.id, Effect: scheduled.effect})
					scheduled.tickRemaining = scheduled.effect.TickSeconds
				}
			}
			if scheduled.durationRemaining > 0 {
				remainingEffects = append(remainingEffects, scheduled)
			}
		}
		if len(remainingEffects) == 0 {
			delete(runtime.targets, targetID)
		} else {
			runtime.targets[targetID] = remainingEffects
		}
	}
	return ticks
}

func (runtime *DamageOverTimeRuntime) RemoveTarget(targetID string) int {
	count := len(runtime.targets[targetID])
	delete(runtime.targets, targetID)
	return count
}

func (runtime *DamageOverTimeRuntime) CountTarget(targetID string) int {
	return len(runtime.targets[targetID])
}

func (runtime *DamageOverTimeRuntime) newScheduledEffect(effect ActiveDamageOverTime) scheduledDamageOverTime {
	runtime.nextID++
	return scheduledDamageOverTime{
		id:                fmt.Sprintf("dot-%d", runtime.nextID),
		effect:            effect,
		tickRemaining:     effect.TickSeconds,
		durationRemaining: effect.DurationSeconds,
	}
}

func validActiveDamageOverTime(effect ActiveDamageOverTime) bool {
	return effect.Target.EntityID != "" && effect.AmountPerTick > 0 && effect.TickSeconds > 0 && effect.DurationSeconds > 0
}

func damageOverTimeStackKey(effect ActiveDamageOverTime) string {
	if effect.StackKey != "" {
		return effect.StackKey
	}
	return fmt.Sprintf("%s:%s:%s", effect.Source.EntityID, effect.Type, effect.Target.EntityType)
}

func matchingDamageOverTimeIndexes(effects []scheduledDamageOverTime, stackKey string) []int {
	matching := make([]int, 0)
	for index, effect := range effects {
		if damageOverTimeStackKey(effect.effect) == stackKey {
			matching = append(matching, index)
		}
	}
	return matching
}

func containsIndex(indexes []int, target int) bool {
	for _, index := range indexes {
		if index == target {
			return true
		}
	}
	return false
}
