package damage

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
	DamageTypeDebug      DamageType = "debug"
)

type DamageRequest struct {
	TargetEntityID   string
	TargetEntityType EntityType
	SourceEntityID   string
	SourceEntityType EntityType
	CurrentHealth    int
	Amount           int
	Type             DamageType
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

func Resolve(req DamageRequest) DamageResult {
	result := DamageResult{
		TargetEntityID:   req.TargetEntityID,
		TargetEntityType: req.TargetEntityType,
		SourceEntityID:   req.SourceEntityID,
		SourceEntityType: req.SourceEntityType,
	}

	if req.Amount <= 0 {
		result.Ignored = true
		return result
	}

	if req.CurrentHealth <= 0 {
		result.Ignored = true
		return result
	}

	remaining := max(req.CurrentHealth-req.Amount, 0)
	result.AppliedToHealth = req.CurrentHealth - remaining
	result.RemainingHealth = remaining
	result.Destroyed = remaining <= 0
	result.Fatal = result.Destroyed && req.TargetEntityType == EntityTypePlayer

	return result
}
