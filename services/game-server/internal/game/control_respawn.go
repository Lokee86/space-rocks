package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func (target *Control) SafeRespawnPosition(playerID string) (physics.Vector2, bool) {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	session, ok := target.game.playerSessions[playerID]
	if !ok || session == nil {
		return physics.Vector2{}, false
	}
	return target.game.safeRespawnPosition(session), true
}

func (target *Control) ForceRespawnPlayer(playerID string, position physics.Vector2, cameraConfig runtimepkg.ClientConfig) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	session, ok := target.game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}

	player := session.NewShip(position)
	target.game.entities.Players[playerID] = player
	fact := target.game.lifeRuntime.ForceActivateForDevtools(playerID)
	if !fact.Accepted {
		delete(target.game.entities.Players, playerID)
		return false
	}
	cameraView := target.game.cameraViews[playerID]
	if cameraView == nil {
		cameraView = &runtimepkg.CameraView{}
		target.game.cameraViews[playerID] = cameraView
	}
	cameraView.X = player.X
	cameraView.Y = player.Y
	cameraView.Config = cameraConfig

	target.game.publishPresentationFrameLocked()
	return true
}

func (game *Game) respawnPlayer(playerID string) bool {
	session, ok := game.playerSessions[playerID]
	if !ok {
		fact := lives.RejectedRespawn(playerID, "session_missing")
		game.emitRespawnBlocked(playerID, fact)
		return false
	}
	respawnFact := game.lifeRuntime.EvaluateRespawn(playerID)
	if !respawnFact.Accepted {
		game.emitRespawnBlocked(playerID, respawnFact)
		return false
	}
	if _, ok := game.entities.Players[playerID]; ok {
		game.emitRespawnBlocked(playerID, lives.RejectedRespawn(playerID, "already_active"))
		return false
	}

	spawnPlan := game.planPlayerRespawn(session)
	player := session.NewRespawnShip(spawnPlan.Position, game.lifeRuntime.Policy().Restoration)
	game.entities.Players[playerID] = player
	commitFact := game.lifeRuntime.CommitRespawn(playerID)
	if !commitFact.Accepted {
		delete(game.entities.Players, playerID)
		game.emitRespawnBlocked(playerID, commitFact)
		return false
	}
	game.setPlayerCameraViewLocked(playerID, player)
	return true
}

func (game *Game) emitRespawnBlocked(playerID string, fact lives.RespawnFact) {
	if game.matchID == "" || game.matchTraceID == "" {
		return
	}
	lives := fact.RemainingLives
	if state, ok := game.lifeRuntime.ParticipantSnapshot(playerID); ok {
		lives = game.projectedPlayerLives(playerID, state)
	}
	logging.Emit(observability.Request{
		Event: observability.EventNameRespawnBlocked,
		Context: observability.Context{
			TraceID:  game.matchTraceID,
			MatchID:  game.matchID,
			PlayerID: playerID,
		},
		Fields: observability.Fields{
			"reason_code":      fact.ReasonCode,
			"lives":            lives,
			"respawn_cooldown": fact.RespawnDelay,
		},
	})
}
