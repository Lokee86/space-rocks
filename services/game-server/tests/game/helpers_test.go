package gametests

import (
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"

	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type scenario struct {
	t    *testing.T
	game *servergame.Game
}

func newScenario(t *testing.T) *scenario {
	t.Helper()

	return &scenario{
		t:    t,
		game: servergame.New(),
	}
}

func (scenario *scenario) addPlayer() string {
	scenario.t.Helper()

	return scenario.game.AddPlayer()
}

func (scenario *scenario) send(playerID string, packet servergame.ClientPacket) {
	scenario.t.Helper()

	scenario.game.HandlePacket(playerID, packet)
}

func (scenario *scenario) step(delta float64) {
	scenario.t.Helper()

	scenario.game.Step(delta)
}

func (scenario *scenario) state(playerID string) servergame.StatePacket {
	scenario.t.Helper()

	rawState := scenario.game.State(playerID)
	if rawState == nil {
		scenario.t.Fatalf("expected state packet for %q, got nil", playerID)
	}

	var packet servergame.StatePacket
	if err := json.Unmarshal(rawState, &packet); err != nil {
		scenario.t.Fatalf("decode state packet for %q: %v", playerID, err)
	}

	return packet
}

func (scenario *scenario) playerState(viewerID string, playerID string) entities.ShipState {
	scenario.t.Helper()

	packet := scenario.state(viewerID)
	player, ok := packet.Players[playerID]
	if !ok {
		scenario.t.Fatalf("expected state packet for %q to include player %q", viewerID, playerID)
	}

	return player
}

func (scenario *scenario) events(playerID string) []servergame.EventState {
	scenario.t.Helper()

	return scenario.state(playerID).Events
}

func (scenario *scenario) useCircleCollisionShapes() {
	scenario.t.Helper()

	scenario.gameField("collisionShapes").Set(reflect.ValueOf(physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{
			Type:   "circle",
			Radius: 5,
		},
		Ship: physics.ImportedCollisionShape{
			Type:   "circle",
			Radius: 20,
		},
		Asteroids: []physics.ImportedCollisionShape{
			{
				Type:   "circle",
				Radius: 20,
			},
		},
	}))
}

func (scenario *scenario) useBulletCapsuleAsteroidPolygonCollisions() {
	scenario.t.Helper()

	scenario.gameField("collisionShapes").Set(reflect.ValueOf(physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{
			Type:   "capsule",
			Radius: 3,
			Height: 24,
		},
		Ship: physics.ImportedCollisionShape{
			Type:   "circle",
			Radius: 20,
		},
		Asteroids: []physics.ImportedCollisionShape{
			{
				Type: "polygon",
				Points: [][]float64{
					{-40, -40},
					{40, -40},
					{40, 40},
					{-40, 40},
				},
			},
		},
	}))
}

func (scenario *scenario) placeAsteroid(id string, position physics.Vector2, size int) {
	scenario.t.Helper()

	scenario.asteroids().SetMapIndex(reflect.ValueOf(id), reflect.ValueOf(&entities.Asteroid{
		ID:   id,
		X:    position.X,
		Y:    position.Y,
		Size: size,
	}))
}

func (scenario *scenario) placeMovingAsteroid(id string, position physics.Vector2, velocity physics.Vector2, size int) {
	scenario.t.Helper()

	scenario.asteroids().SetMapIndex(reflect.ValueOf(id), reflect.ValueOf(&entities.Asteroid{
		ID:       id,
		X:        position.X,
		Y:        position.Y,
		Velocity: velocity,
		Size:     size,
	}))
}

func (scenario *scenario) addCameraView(id string, position physics.Vector2, config entities.ClientConfig) {
	scenario.t.Helper()

	scenario.gameField("cameraViews").SetMapIndex(reflect.ValueOf(id), reflect.ValueOf(&entities.CameraView{
		X:      position.X,
		Y:      position.Y,
		Config: config,
	}))
}

func (scenario *scenario) placeBullet(id string, ownerID string, position physics.Vector2, velocity physics.Vector2) {
	scenario.t.Helper()

	bullet := entities.NewBullet(id, ownerID, position, 0, velocity, entities.DefaultShipStats().BulletLifetime)
	scenario.bullets().SetMapIndex(reflect.ValueOf(id), reflect.ValueOf(bullet))
}

func (scenario *scenario) setPlayerPosition(playerID string, position physics.Vector2) {
	scenario.t.Helper()

	player := scenario.player(playerID)
	player.X = position.X
	player.Y = position.Y
}

func (scenario *scenario) setPlayerPaused(playerID string, paused bool) {
	scenario.t.Helper()

	scenario.player(playerID).Paused = paused
}

func (scenario *scenario) setPlayerInvulnerability(playerID string, seconds float64) {
	scenario.t.Helper()

	scenario.player(playerID).InvulnerabilityRemaining = seconds
}

func (scenario *scenario) setPlayerLives(playerID string, lives int) {
	scenario.t.Helper()

	scenario.player(playerID).Lives = lives
	scenario.sessionField(playerID, "Lives").SetInt(int64(lives))
}

func (scenario *scenario) removePlayerEntity(playerID string) {
	scenario.t.Helper()

	scenario.players().SetMapIndex(reflect.ValueOf(playerID), reflect.Value{})
}

func (scenario *scenario) playerExists(viewerID string, playerID string) bool {
	scenario.t.Helper()

	_, ok := scenario.state(viewerID).Players[playerID]
	return ok
}

func (scenario *scenario) playerEntityExists(playerID string) bool {
	scenario.t.Helper()

	return scenario.players().MapIndex(reflect.ValueOf(playerID)).IsValid()
}

func (scenario *scenario) setSessionSpawnPosition(playerID string, position physics.Vector2) {
	scenario.t.Helper()

	scenario.sessionField(playerID, "SpawnPosition").Set(reflect.ValueOf(position))
}

func (scenario *scenario) sessionSpawnPosition(playerID string) physics.Vector2 {
	scenario.t.Helper()

	return scenario.sessionField(playerID, "SpawnPosition").Interface().(physics.Vector2)
}

func (scenario *scenario) advanceRespawnTimer(playerID string, delta float64) {
	scenario.t.Helper()

	cooldown := scenario.sessionField(playerID, "RespawnCooldown")
	cooldown.SetFloat(max(0, cooldown.Float()-delta))
}

func (scenario *scenario) setAsteroidSpawnElapsed(elapsed float64) {
	scenario.t.Helper()

	scenario.gameField("asteroidSpawnElapsed").SetFloat(elapsed)
}

func (scenario *scenario) asteroidSpawnElapsed() float64 {
	scenario.t.Helper()

	return scenario.gameField("asteroidSpawnElapsed").Float()
}

func (scenario *scenario) pendingEventCount(playerID string) int {
	scenario.t.Helper()

	events := scenario.pendingEvents().MapIndex(reflect.ValueOf(playerID))
	if !events.IsValid() {
		return 0
	}

	return events.Len()
}

func (scenario *scenario) playerPendingDespawn(playerID string) bool {
	scenario.t.Helper()

	return scenario.player(playerID).PendingDespawn
}

func (scenario *scenario) asteroidPendingDespawn(id string) bool {
	scenario.t.Helper()

	asteroid := scenario.asteroids().MapIndex(reflect.ValueOf(id))
	if !asteroid.IsValid() || asteroid.IsNil() {
		scenario.t.Fatalf("expected asteroid %q", id)
	}

	return asteroid.Interface().(*entities.Asteroid).PendingDespawn
}

func (scenario *scenario) asteroidExists(id string) bool {
	scenario.t.Helper()

	return scenario.asteroids().MapIndex(reflect.ValueOf(id)).IsValid()
}

func (scenario *scenario) bulletPendingDespawn(id string) bool {
	scenario.t.Helper()

	return scenario.bullet(id).PendingDespawn
}

func (scenario *scenario) bulletLife(id string) float64 {
	scenario.t.Helper()

	return scenario.bullet(id).Life
}

func (scenario *scenario) bullet(id string) *entities.Bullet {
	scenario.t.Helper()

	bullet := scenario.bullets().MapIndex(reflect.ValueOf(id))
	if !bullet.IsValid() || bullet.IsNil() {
		scenario.t.Fatalf("expected bullet %q", id)
	}

	return bullet.Interface().(*entities.Bullet)
}

func (scenario *scenario) playerDebugBool(playerID string, fieldName string) bool {
	scenario.t.Helper()

	return reflect.ValueOf(scenario.player(playerID)).Elem().FieldByName("DevTools").FieldByName(fieldName).Bool()
}

func (scenario *scenario) worldDebugBool(fieldName string) bool {
	scenario.t.Helper()

	return scenario.gameField("worldDevTools").FieldByName(fieldName).Bool()
}

func (scenario *scenario) worldFrozen() bool {
	scenario.t.Helper()

	return scenario.worldDebugBool("FreezeAsteroids") &&
		scenario.worldDebugBool("FreezeBullets") &&
		scenario.worldDebugBool("FreezeSpawning") &&
		scenario.worldDebugBool("FreezeCollisions")
}

func (scenario *scenario) player(playerID string) *entities.Ship {
	scenario.t.Helper()

	player := scenario.players().MapIndex(reflect.ValueOf(playerID))
	if !player.IsValid() || player.IsNil() {
		scenario.t.Fatalf("expected player %q", playerID)
	}

	return player.Interface().(*entities.Ship)
}

func (scenario *scenario) players() reflect.Value {
	scenario.t.Helper()

	return scenario.stateField("Players")
}

func (scenario *scenario) bullets() reflect.Value {
	scenario.t.Helper()

	return scenario.stateField("Projectiles")
}

func (scenario *scenario) asteroids() reflect.Value {
	scenario.t.Helper()

	return scenario.stateField("Asteroids")
}

func (scenario *scenario) pendingEvents() reflect.Value {
	scenario.t.Helper()

	return scenario.gameField("pendingEvents")
}

func (scenario *scenario) sessionField(playerID string, fieldName string) reflect.Value {
	scenario.t.Helper()

	session := scenario.gameField("playerSessions").MapIndex(reflect.ValueOf(playerID))
	if !session.IsValid() || session.IsNil() {
		scenario.t.Fatalf("expected session for player %q", playerID)
	}

	return exportValue(session.Elem().FieldByName(fieldName))
}

func (scenario *scenario) stateField(fieldName string) reflect.Value {
	scenario.t.Helper()

	return exportValue(scenario.gameField("state").FieldByName(fieldName))
}

func (scenario *scenario) gameField(fieldName string) reflect.Value {
	scenario.t.Helper()

	return exportValue(reflect.ValueOf(scenario.game).Elem().FieldByName(fieldName))
}

func exportValue(value reflect.Value) reflect.Value {
	if value.CanSet() {
		return value
	}

	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem()
}
