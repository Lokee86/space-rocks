package game

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

type playerSession struct {
	ID            string
	TeamID        teams.ID
	ShipTypeID    string
	Stats         runtime.ShipStats
	SpawnPosition physics.Vector2
	Config        runtime.ClientConfig
	Targeting     PlayerTargeting
	Score         int
	Suspension    runtime.SuspensionState
	DamageOptions runtime.DamageOptions
	PlayerArmory  weapons.PlayerArmory
	betweenLife   *betweenLifeState
}

type betweenLifeState struct {
	HealthValue     int
	ShieldValue     int
	WeaponState     weapons.State
	DamageModifiers []damage.DamageModifier
}

func newPlayerSession(id string, spawnPosition physics.Vector2) *playerSession {
	return &playerSession{
		ID:            id,
		ShipTypeID:    runtime.DefaultShipTypeID,
		Stats:         runtime.ResolveShipStats(runtime.DefaultShipTypeID),
		SpawnPosition: spawnPosition,
		Config:        runtime.DefaultCameraConfig(),
		Targeting:     EmptyPlayerTargeting(),
		PlayerArmory:  weapons.DefaultPlayerArmory(),
	}
}

func (session *playerSession) NewShip(position physics.Vector2) *runtime.Ship {
	ship := &runtime.Ship{
		ID:            session.ID,
		ShipTypeID:    session.ShipTypeID,
		Stats:         session.Stats,
		X:             position.X,
		Y:             position.Y,
		Config:        session.Config,
		Health:        session.Stats.MaxHealth,
		DamageOptions: session.DamageOptions,
	}
	ship.ShipWeapons.Primary = session.PlayerArmory.Primary
	ship.ShipWeapons.Secondary = session.PlayerArmory.Secondary
	session.Targeting.ApplyToShip(ship)
	return ship
}

func (session *playerSession) CaptureBetweenLifeState(ship *runtime.Ship) {
	if ship == nil {
		return
	}
	session.betweenLife = &betweenLifeState{
		HealthValue:     ship.Health,
		ShieldValue:     ship.Shields,
		WeaponState:     ship.WeaponState,
		DamageModifiers: append([]damage.DamageModifier(nil), ship.DamageModifiers...),
	}
}

func (session *playerSession) NewRespawnShip(position physics.Vector2, policy lives.RestorationPolicy) *runtime.Ship {
	if policy.Loadout == lives.LoadoutReset {
		session.PlayerArmory = weapons.DefaultPlayerArmory()
	}
	armory := session.PlayerArmory
	ship := session.NewShip(position)
	ship.ShipWeapons.Primary = armory.Primary
	ship.ShipWeapons.Secondary = armory.Secondary
	if policy.Shields == lives.RestorationFull {
		ship.Shields = min(session.Stats.MaxShields, session.Stats.MaxShields)
	} else if session.betweenLife != nil {
		ship.Shields = min(session.Stats.MaxShields, session.betweenLife.ShieldValue)
	}
	if session.betweenLife != nil {
		if policy.Loadout == lives.LoadoutPersist {
			ship.WeaponState = restoreWeaponState(session.betweenLife.WeaponState, ship.ShipWeapons, policy.ShortCooldownThreshold)
		}
		ship.DamageModifiers = restoreDamageModifiers(session.betweenLife.DamageModifiers, policy.TemporaryEffects)
	}
	if policy.Health == lives.RestorationFull {
		ship.Health = session.Stats.MaxHealth
	} else if session.betweenLife != nil {
		ship.Health = session.betweenLife.HealthValue
	} else {
		ship.Health = 0
	}
	session.betweenLife = nil
	return ship
}

func restoreWeaponState(state weapons.State, armory weapons.ShipWeapons, threshold float64) weapons.State {
	state.Primary = restoreWeaponSlot(state.Primary, armory.Primary, threshold)
	state.Secondary = restoreWeaponSlot(state.Secondary, armory.Secondary, threshold)
	return state
}

func restoreWeaponSlot(state weapons.SlotState, equipped weapons.Equipped, threshold float64) weapons.SlotState {
	profile, ok := weapons.Lookup(equipped.ID)
	if ok && profile.CooldownSeconds < threshold {
		state.CooldownRemaining = 0
	}
	return state
}

func restoreDamageModifiers(modifiers []damage.DamageModifier, policy lives.TemporaryEffectsPolicy) []damage.DamageModifier {
	if policy == lives.TemporaryEffectsPersist {
		return append([]damage.DamageModifier(nil), modifiers...)
	}
	filtered := make([]damage.DamageModifier, 0, len(modifiers))
	for _, modifier := range modifiers {
		if modifier.PersistsThroughDeath {
			filtered = append(filtered, modifier)
		}
	}
	return filtered
}

func (game *Game) planInitialPlayerSpawn(playerIndex int, playerID string) PlayerSpawnPlan {
	shapeID := runtime.ResolveShipStats(runtime.DefaultShipTypeID).CollisionShapeID
	return game.planPlayerSpawn(playerspawn.Request{
		ProfileID:        game.lifeRuntime.Policy().SpawnProfileID,
		PlayerID:         playerID,
		SpawnReason:      string(SpawnReasonInitialPlayer),
		PreferredOrigin:  preferredInitialSpawnPosition(playerIndex),
		CollisionShapeID: shapeID,
	}, SpawnReasonInitialPlayer)
}

func preferredInitialSpawnPosition(playerIndex int) physics.Vector2 {
	return physics.Vector2{
		X: 576 + float64(playerIndex%4)*80,
		Y: 320 + float64(playerIndex/4)*80,
	}
}

func (game *Game) safeRespawnPosition(session *playerSession) physics.Vector2 {
	return game.planPlayerRespawn(session).Position
}

func (game *Game) planPlayerRespawn(session *playerSession) PlayerSpawnPlan {
	return game.planPlayerSpawn(playerspawn.Request{
		ProfileID:        game.lifeRuntime.Policy().SpawnProfileID,
		PlayerID:         session.ID,
		SpawnReason:      string(SpawnReasonPlayerRespawn),
		PreferredOrigin:  session.SpawnPosition,
		CollisionShapeID: session.Stats.CollisionShapeID,
	}, SpawnReasonPlayerRespawn)
}

func (game *Game) planPlayerSpawn(request playerspawn.Request, reason SpawnReason) PlayerSpawnPlan {
	plan, err := playerspawn.PlanBasicSafeSpawnV1(request, game)
	if err != nil {
		panic(err)
	}
	return PlayerSpawnPlan{EntityType: SpawnEntityTypePlayer, Reason: reason, PlayerID: request.PlayerID, Position: plan.Position}
}

func (game *Game) IsSafeSpawn(position physics.Vector2, ignorePlayerID string, collisionShapeID string) bool {
	shape, err := game.collisionShapes.ShipShapeByID(collisionShapeID)
	if err != nil {
		return true
	}

	shipBody := physics.CollisionBody{
		ID:       "respawn",
		Position: position,
		Shape:    shape,
	}
	for _, asteroid := range game.entities.Asteroids {
		if asteroid.IsPendingDespawn() {
			continue
		}

		asteroidBody, ok := asteroid.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		if !hasRespawnClearance(shipBody, asteroidBody, constants.PlayerRespawnBuffer) {
			return false
		}
	}
	for id, player := range game.entities.Players {
		if id == ignorePlayerID || player.IsPendingDespawn() {
			continue
		}

		playerBody, ok := player.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		if !hasRespawnClearance(shipBody, playerBody, constants.PlayerRespawnBuffer) {
			return false
		}
	}

	return true
}

func hasRespawnClearance(shipBody physics.CollisionBody, asteroidBody physics.CollisionBody, buffer float64) bool {
	clearance := collisionShapeRadius(shipBody.Shape) + collisionShapeRadius(asteroidBody.Shape) + buffer
	return space.Distance(shipBody.Position, asteroidBody.Position) > clearance
}

func collisionShapeRadius(shape physics.CollisionShape) float64 {
	switch shape.Type {
	case physics.CollisionShapeCircle:
		return shape.Radius
	case physics.CollisionShapeCapsule:
		return shape.Height * 0.5
	case physics.CollisionShapeRectangle:
		return shape.Size.Multiply(0.5).Length()
	case physics.CollisionShapePolygon:
		var radius float64
		for _, point := range shape.Points {
			radius = math.Max(radius, point.Length())
		}
		return radius
	default:
		return 0
	}
}
