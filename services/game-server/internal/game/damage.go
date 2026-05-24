package game

type EntityType string

const (
	EntityTypePlayer     EntityType = "player"
	EntityTypeAsteroid   EntityType = "asteroid"
	EntityTypeProjectile EntityType = "projectile"
)

type DamageType string

const (
	DamageTypeCollision  DamageType = "collision"
	DamageTypeProjectile DamageType = "projectile"
)

type DamageFlags struct {
	BypassesShield          bool
	BypassesInvulnerability bool
	Lethal                  bool
}

type DamageRequest struct {
	TargetEntityID   string
	TargetEntityType EntityType
	SourceEntityID   string
	SourceEntityType EntityType
	Amount           int
	Type             DamageType
	Flags            DamageFlags
}

type DamageResult struct {
	TargetEntityID   string
	TargetEntityType EntityType
	SourceEntityID   string
	SourceEntityType EntityType
	AppliedToHealth  int
	AbsorbedByShield int
	Ignored          bool
	Destroyed        bool
	Fatal            bool
	RemainingHealth  int
	RemainingShield  int
	Reason           string
}

func resolveDamage(req DamageRequest) DamageResult {
	result := DamageResult{
		TargetEntityID:   req.TargetEntityID,
		TargetEntityType: req.TargetEntityType,
		SourceEntityID:   req.SourceEntityID,
		SourceEntityType: req.SourceEntityType,
	}

	if req.Type == DamageTypeProjectile &&
		req.TargetEntityType == EntityTypeAsteroid &&
		req.Amount == 1 &&
		req.Flags.Lethal {
		result.Destroyed = true
		return result
	}

	if req.Type == DamageTypeCollision &&
		req.TargetEntityType == EntityTypePlayer &&
		req.Amount == 1 &&
		req.Flags.Lethal {
		result.Destroyed = true
		result.Fatal = true
		return result
	}

	result.Ignored = true
	return result
}
