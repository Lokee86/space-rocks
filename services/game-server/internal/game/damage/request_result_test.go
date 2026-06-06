package damage

import "testing"

func TestDamageResolutionRequestConstruction(t *testing.T) {
	req := DamageResolutionRequest{
		Source: DamageSource{
			EntityID:   "bullet-1",
			EntityType: EntityTypeProjectile,
			Cause:      DamageCauseProjectile,
		},
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     3,
			Shield:     1,
		},
		Spec: DamageSpec{
			Amount:       2,
			Type:         DamageTypeExplosive,
			Cause:        DamageCauseProjectile,
			BypassShield: true,
		},
		Modifiers: []DamageModifier{
			{Type: DamageTypeExplosive, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationAdd, Value: 1},
		},
	}

	if req.Source.EntityID != "bullet-1" {
		t.Fatalf("expected source entity id %q, got %q", "bullet-1", req.Source.EntityID)
	}
	if req.Target.EntityID != "asteroid-1" {
		t.Fatalf("expected target entity id %q, got %q", "asteroid-1", req.Target.EntityID)
	}
	if req.Spec.Amount != 2 {
		t.Fatalf("expected spec amount %d, got %d", 2, req.Spec.Amount)
	}
	if len(req.Modifiers) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(req.Modifiers))
	}
}

func TestDamageSpecConstructionKineticCollision(t *testing.T) {
	spec := DamageSpec{
		Amount:       1,
		Type:         DamageTypeKinetic,
		Cause:        DamageCauseCollision,
		BypassShield: false,
	}

	if spec.Amount != 1 {
		t.Fatalf("expected amount %d, got %d", 1, spec.Amount)
	}
	if spec.Type != DamageTypeKinetic {
		t.Fatalf("expected kind %q, got %q", DamageTypeKinetic, spec.Type)
	}
	if spec.Cause != DamageCauseCollision {
		t.Fatalf("expected cause %q, got %q", DamageCauseCollision, spec.Cause)
	}
	if spec.BypassShield {
		t.Fatal("expected bypass shield to be false")
	}
}

func TestDamageSpecConstructionExplosiveProjectile(t *testing.T) {
	spec := DamageSpec{
		Amount:       3,
		Type:         DamageTypeExplosive,
		Cause:        DamageCauseProjectile,
		BypassShield: true,
	}

	if spec.Amount != 3 {
		t.Fatalf("expected amount %d, got %d", 3, spec.Amount)
	}
	if spec.Type != DamageTypeExplosive {
		t.Fatalf("expected kind %q, got %q", DamageTypeExplosive, spec.Type)
	}
	if spec.Cause != DamageCauseProjectile {
		t.Fatalf("expected cause %q, got %q", DamageCauseProjectile, spec.Cause)
	}
	if !spec.BypassShield {
		t.Fatal("expected bypass shield to be true")
	}
}

func TestDamageSourceConstruction(t *testing.T) {
	source := DamageSource{
		EntityID:   "bullet-1",
		EntityType: EntityTypeProjectile,
		Cause:      DamageCauseProjectile,
	}

	if source.EntityID != "bullet-1" {
		t.Fatalf("expected entity id %q, got %q", "bullet-1", source.EntityID)
	}
	if source.EntityType != EntityTypeProjectile {
		t.Fatalf("expected entity type %q, got %q", EntityTypeProjectile, source.EntityType)
	}
	if source.Cause != DamageCauseProjectile {
		t.Fatalf("expected cause %q, got %q", DamageCauseProjectile, source.Cause)
	}
}

func TestDamageTargetConstruction(t *testing.T) {
	target := DamageTarget{
		EntityID:   "player-1",
		EntityType: EntityTypePlayer,
		Health:     3,
		Shield:     1,
	}

	if target.EntityID != "player-1" {
		t.Fatalf("expected entity id %q, got %q", "player-1", target.EntityID)
	}
	if target.EntityType != EntityTypePlayer {
		t.Fatalf("expected entity type %q, got %q", EntityTypePlayer, target.EntityType)
	}
	if target.Health != 3 {
		t.Fatalf("expected health %d, got %d", 3, target.Health)
	}
	if target.Shield != 1 {
		t.Fatalf("expected shield %d, got %d", 1, target.Shield)
	}
}

func TestDamageResultConstruction(t *testing.T) {
	result := DamageResult{
		TargetEntityID:  "player-1",
		TargetEntityType: EntityTypePlayer,
		BaseAmount:      4,
		ModifiedAmount:  5,
	}

	if result.TargetEntityID != "player-1" {
		t.Fatalf("expected result target id %q, got %q", "player-1", result.TargetEntityID)
	}
	if result.BaseAmount != 4 {
		t.Fatalf("expected base amount %d, got %d", 4, result.BaseAmount)
	}
	if result.ModifiedAmount != 5 {
		t.Fatalf("expected modified amount %d, got %d", 5, result.ModifiedAmount)
	}
}
