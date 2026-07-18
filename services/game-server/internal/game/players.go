package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func (game *Game) AddPlayer() string {
	return game.AddPlayerWithTeam(teams.NoTeam)
}

func (game *Game) AddPlayerWithTeam(teamID teams.ID) string {
	game.mu.Lock()
	defer game.mu.Unlock()
	if err := teams.ValidateTeamID(teamID); err != nil {
		teamID = teams.NoTeam
	}

	return game.addPlayerLocked(teamID)
}

func (game *Game) addPlayerLocked(teamID teams.ID) string {
	playerIndex := game.nextID
	game.nextID++

	playerID := fmt.Sprintf("player-%d", game.nextID)
	spawnPlan := game.planInitialPlayerSpawn(playerIndex, playerID)
	spawnPosition := spawnPlan.Position
	if err := game.lifeRuntime.RegisterParticipant(lives.ParticipantRegistration{PlayerID: playerID, TeamID: teamID}); err != nil {
		game.nextID--
		return ""
	}
	if err := game.participationRuntime.RegisterParticipant(playerID); err != nil {
		game.lifeRuntime.RollbackParticipant(playerID)
		game.nextID--
		return ""
	}
	session := newPlayerSession(playerID, spawnPosition)
	session.TeamID = teamID
	player := session.NewShip(spawnPosition)
	game.playerSessions[playerID] = session
	game.participantRecords[playerID] = &participantRecord{ID: playerID, TeamID: teamID}
	game.entities.Players[playerID] = player
	game.setPlayerCameraViewLocked(playerID, player)
	game.pendingPresentationEvents[playerID] = nil
	game.publishPresentationFrameLocked()

	return playerID
}

func (game *Game) setPlayerCameraViewLocked(playerID string, player *runtime.Ship) {
	if playerID == "" || player == nil {
		return
	}

	cameraView, ok := game.cameraViews[playerID]
	if !ok || cameraView == nil {
		cameraView = &runtime.CameraView{}
		game.cameraViews[playerID] = cameraView
	}

	cameraView.X = player.X
	cameraView.Y = player.Y

	// Prefer an existing valid config to avoid flicker. Otherwise seed from session/player.
	cameraConfig := cameraView.Config
	if cameraConfig.VisibleWorldWidth <= 0 || cameraConfig.VisibleWorldHeight <= 0 {
		if session, ok := game.playerSessions[playerID]; ok && session != nil {
			if session.Config.VisibleWorldWidth > 0 && session.Config.VisibleWorldHeight > 0 {
				cameraConfig = session.Config
			}
		}
		if cameraConfig.VisibleWorldWidth <= 0 || cameraConfig.VisibleWorldHeight <= 0 {
			if player.Config.VisibleWorldWidth > 0 && player.Config.VisibleWorldHeight > 0 {
				cameraConfig = player.Config
			}
		}
		if cameraConfig.VisibleWorldWidth <= 0 || cameraConfig.VisibleWorldHeight <= 0 {
			cameraConfig = runtime.DefaultCameraConfig()
		}
	}
	cameraView.Config = runtime.ClampCameraConfig(cameraConfig)
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.lifeRuntime.RemoveParticipant(playerID, "player_removed")
	game.participationRuntime.UnregisterParticipant(playerID)
	game.removeActivePlayerLocked(playerID)
}

// RollbackPlayerAdd removes a provisional player that never completed room activation.
func (game *Game) RollbackPlayerAdd(playerID string) {
	game.DiscardPlayer(playerID)
}

// DiscardPlayer removes both active participation and durable match facts.
// Normal departures should use RemovePlayer so completed-match results remain intact.
func (game *Game) DiscardPlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.lifeRuntime.RollbackParticipant(playerID)
	game.participationRuntime.UnregisterParticipant(playerID)
	game.removeActivePlayerLocked(playerID)
	delete(game.participantRecords, playerID)
}

func (game *Game) removeActivePlayerLocked(playerID string) {
	delete(game.entities.Players, playerID)
	game.inputMu.Lock()
	delete(game.pendingPlayerInputs, playerID)
	game.inputMu.Unlock()
	delete(game.cameraViews, playerID)
	delete(game.playerSessions, playerID)
	delete(game.botControllers, playerID)
	game.clearTargetsForMissingPlayersLocked()
	delete(game.pendingPresentationEvents, playerID)
	game.publishPresentationFrameLocked()
}

func (game *Game) playerLives(playerID string) int {
	if state, ok := game.lifeRuntime.ParticipantSnapshot(playerID); ok {
		return game.projectedPlayerLives(playerID, state)
	}

	return 0
}
