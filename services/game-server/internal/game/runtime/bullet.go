package runtime

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
)

func NewBullet(
	id string,
	ownerID string,
	position physics.Vector2,
	rotation float64,
	velocity physics.Vector2,
	lifetime float64,
) *Bullet {
	return &Bullet{
		ID:       id,
		OwnerID:  ownerID,
		X:        position.X,
		Y:        position.Y,
		Rotation: rotation,
		Velocity: velocity,
		Life:     lifetime,
		Damage:   constants.BasicCannonDamage,
	}
}

func NewBulletFromWeaponSpawn(id string, ownerID string, spawn weapons.ProjectileSpawn) *Bullet {
	return &Bullet{
		ID:             id,
		OwnerID:        ownerID,
		WeaponID:       spawn.WeaponID,
		ProjectileType: spawn.ProjectileType,
		ImpactEffect:   spawn.ImpactEffect,
		X:              spawn.Position.X,
		Y:              spawn.Position.Y,
		Rotation:       spawn.Rotation,
		Velocity:       spawn.Velocity,
		Life:           spawn.Lifetime,
		Damage:         spawn.Damage.Amount,
		DamageSpec:     spawn.Damage,
	}
}

func (bullet *Bullet) State() BulletState {
	return BulletState{
		ID:             bullet.ID,
		OwnerID:        bullet.OwnerID,
		WeaponID:       string(bullet.WeaponID),
		ProjectileType: bullet.ProjectileType,
		X:              bullet.X,
		Y:              bullet.Y,
		Rotation:       bullet.Rotation,
	}
}

func (bullet *Bullet) Position() physics.Vector2 {
	return physics.Vector2{X: bullet.X, Y: bullet.Y}
}

func (bullet *Bullet) IsPendingDespawn() bool {
	return bullet.PendingDespawn
}

func (bullet *Bullet) ReadyForRemoval() bool {
	return bullet.PendingDespawn && bullet.DespawnDelay <= 0
}

func (bullet *Bullet) IsExpired() bool {
	return bullet.Life <= 0
}

func (bullet *Bullet) MarkPendingDespawn(delay float64) {
	bullet.PendingDespawn = true
	bullet.DespawnDelay = delay
	bullet.Velocity = physics.Vector2{}
}

func (bullet *Bullet) CollisionBody(catalog physics.CollisionShapeCatalog) (physics.CollisionBody, bool) {
	shape, err := catalog.BulletShape()
	if err != nil {
		return physics.CollisionBody{}, false
	}

	return physics.CollisionBody{
		ID:       bullet.ID,
		Position: bullet.Position(),
		Rotation: bullet.Rotation,
		Shape:    shape,
	}, true
}
