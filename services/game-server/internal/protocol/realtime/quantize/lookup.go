package quantize

var fieldPolicies = map[string]Policy{
	"session.elapsed":                  mustPolicy(PolicySeconds),
	"session.duration":                 mustPolicy(PolicySeconds),
	"session.timer":                    mustPolicy(PolicySeconds),
	"session.lifetime":                 mustPolicy(PolicySeconds),
	"session.players.respawn_cooldown": mustPolicy(PolicySeconds),
	"session.players.spawn_x":          mustPolicy(PolicyPosition),
	"session.players.spawn_y":          mustPolicy(PolicyPosition),
	"overlay.respawn_cooldown":         mustPolicy(PolicySeconds),
	"overlay.primary_cooldown_remaining": mustPolicy(PolicySeconds),
	"overlay.secondary_cooldown_remaining": mustPolicy(PolicySeconds),
	"world.ships.x":                    mustPolicy(PolicyPosition),
	"world.ships.y":                    mustPolicy(PolicyPosition),
	"world.ships.rotation":             mustPolicy(PolicyFloatGeneric),
	"world.bullets.x":                  mustPolicy(PolicyPosition),
	"world.bullets.y":                  mustPolicy(PolicyPosition),
	"world.bullets.rotation":           mustPolicy(PolicyFloatGeneric),
	"world.asteroids.x":                mustPolicy(PolicyPosition),
	"world.asteroids.y":                mustPolicy(PolicyPosition),
	"world.asteroids.scale":            mustPolicy(PolicyFloatGeneric),
	"world.pickups.x":                  mustPolicy(PolicyPosition),
	"world.pickups.y":                  mustPolicy(PolicyPosition),
	"world.pickups.age_seconds":        mustPolicy(PolicySeconds),
	"world.pickups.lifespan_seconds":   mustPolicy(PolicySeconds),
	"event.bullet_blast.x":             mustPolicy(PolicyPosition),
	"event.bullet_blast.y":             mustPolicy(PolicyPosition),
	"event.ship_death.x":               mustPolicy(PolicyPosition),
	"event.ship_death.y":               mustPolicy(PolicyPosition),
	"event.ship_death.respawn_delay":   mustPolicy(PolicySeconds),
	"event.damage_applied.x":           mustPolicy(PolicyPosition),
	"event.damage_applied.y":           mustPolicy(PolicyPosition),
	"event.radial_effect_started.x":    mustPolicy(PolicyPosition),
	"event.radial_effect_started.y":    mustPolicy(PolicyPosition),
	"event.pickup_collected.x":         mustPolicy(PolicyPosition),
	"event.pickup_collected.y":         mustPolicy(PolicyPosition),
	"event.pickup_expired.x":           mustPolicy(PolicyPosition),
	"event.pickup_expired.y":           mustPolicy(PolicyPosition),
	"event.pickup_dropped.x":           mustPolicy(PolicyPosition),
	"event.pickup_dropped.y":           mustPolicy(PolicyPosition),
	"event.damage_over_time_tick.x":    mustPolicy(PolicyPosition),
	"event.damage_over_time_tick.y":    mustPolicy(PolicyPosition),
}

func LookupPolicy(fieldPath string) (Policy, bool) {
	policy, ok := fieldPolicies[fieldPath]
	if ok {
		return policy, true
	}
	return mustPolicy(PolicyFloatGeneric), false
}

func mustPolicy(name PolicyName) Policy {
	policy, ok := PolicyByName(name)
	if !ok {
		panic("quantize: missing policy " + string(name))
	}
	return policy
}
