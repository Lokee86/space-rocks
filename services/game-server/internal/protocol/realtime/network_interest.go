package realtime

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	viewvisibility "github.com/Lokee86/space-rocks/services/game-server/internal/game/visibility"
)

func applyNetworkInterest(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, viewTargetID string) game.GameplayPresentationSnapshot {
	view, ok := interestCameraView(snapshot, viewTargetID)
	if !ok {
		return snapshot
	}

	previousShips, previousAsteroids, previousBullets, previousPickups := previousWorldInterest(state)
	snapshot.Players = interestedShips(snapshot, view, viewTargetID, previousShips)
	snapshot.Asteroids = interestedAsteroids(snapshot, view, previousAsteroids)
	snapshot.Bullets = interestedBullets(snapshot, view, previousBullets)
	snapshot.Pickups = interestedPickups(snapshot, view, previousPickups)
	return snapshot
}

func interestCameraView(snapshot game.GameplayPresentationSnapshot, viewTargetID string) (runtime.CameraView, bool) {
	view := snapshot.CameraView
	hasView := snapshot.HasCameraView
	if !hasView {
		view.Config = runtime.DefaultCameraConfig()
	}
	if viewTargetID != "" {
		if locator, ok := snapshot.PlayerLocators[viewTargetID]; ok {
			view.X = locator.X
			view.Y = locator.Y
			return view, true
		}
	}
	if hasView {
		return view, true
	}
	if locator, ok := snapshot.PlayerLocators[snapshot.SelfID]; ok {
		view.X = locator.X
		view.Y = locator.Y
		return view, true
	}
	return view, false
}

func previousWorldInterest(state RealtimeSessionState) (map[string]bool, map[string]bool, map[string]bool, map[string]bool) {
	ships := make(map[string]bool)
	asteroids := make(map[string]bool)
	bullets := make(map[string]bool)
	pickups := make(map[string]bool)
	projection, ok := state.BaselineProjection(LaneWorld)
	if !ok {
		return ships, asteroids, bullets, pickups
	}
	world, ok := projection.(WorldWireFullPacket)
	if !ok {
		return ships, asteroids, bullets, pickups
	}
	for _, record := range world.Ships {
		ships[record.ID] = true
	}
	for _, record := range world.Asteroids {
		asteroids[record.ID] = true
	}
	for _, record := range world.Bullets {
		bullets[record.ID] = true
	}
	for _, record := range world.Pickups {
		pickups[record.ID] = true
	}
	return ships, asteroids, bullets, pickups
}

func interestMargin(previouslyRelevant bool, entryMargin float64, exitMargin float64) float64 {
	if previouslyRelevant {
		return exitMargin
	}
	return entryMargin
}

func interestedShips(snapshot game.GameplayPresentationSnapshot, view runtime.CameraView, viewTargetID string, previous map[string]bool) map[string]runtime.ShipState {
	result := make(map[string]runtime.ShipState)
	for id, ship := range snapshot.Players {
		alwaysRelevant := id == snapshot.SelfID || id == viewTargetID
		margin := interestMargin(previous[id], constants.NetworkInterestEntryMargin, constants.NetworkInterestExitMargin)
		if alwaysRelevant || viewvisibility.Contains(view, physics.Vector2{X: ship.X, Y: ship.Y}, margin) {
			result[id] = ship
		}
	}
	return result
}

func interestedAsteroids(snapshot game.GameplayPresentationSnapshot, view runtime.CameraView, previous map[string]bool) map[string]runtime.AsteroidState {
	result := make(map[string]runtime.AsteroidState)
	for id, asteroid := range snapshot.Asteroids {
		margin := interestMargin(previous[id], constants.NetworkInterestEntryMargin, constants.NetworkInterestExitMargin)
		if viewvisibility.Contains(view, physics.Vector2{X: asteroid.X, Y: asteroid.Y}, margin) {
			result[id] = asteroid
		}
	}
	return result
}

func interestedBullets(snapshot game.GameplayPresentationSnapshot, view runtime.CameraView, previous map[string]bool) map[string]runtime.BulletState {
	result := make(map[string]runtime.BulletState)
	exitMargin := constants.NetworkInterestProjectileMargin + (constants.NetworkInterestExitMargin - constants.NetworkInterestEntryMargin)
	for id, bullet := range snapshot.Bullets {
		margin := interestMargin(previous[id], constants.NetworkInterestProjectileMargin, exitMargin)
		if viewvisibility.Contains(view, physics.Vector2{X: bullet.X, Y: bullet.Y}, margin) {
			result[id] = bullet
		}
	}
	return result
}

func interestedPickups(snapshot game.GameplayPresentationSnapshot, view runtime.CameraView, previous map[string]bool) map[string]runtime.PickupState {
	result := make(map[string]runtime.PickupState)
	for id, pickup := range snapshot.Pickups {
		margin := interestMargin(previous[id], constants.NetworkInterestEntryMargin, constants.NetworkInterestExitMargin)
		if viewvisibility.Contains(view, physics.Vector2{X: pickup.X, Y: pickup.Y}, margin) {
			result[id] = pickup
		}
	}
	return result
}
