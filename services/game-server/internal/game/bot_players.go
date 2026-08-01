package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/bots"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func (game *Game) AddBot() string {
	return game.AddBotWithTeam(teams.NoTeam)
}

func (game *Game) AddBotWithTeam(teamID teams.ID) string {
	game.mu.Lock()
	defer game.mu.Unlock()
	if err := teams.ValidateTeamID(teamID); err != nil {
		teamID = teams.NoTeam
	}

	playerID := game.addPlayerLocked(teamID)
	game.enableBotPlayerLocked(playerID)
	return playerID
}

func (game *Game) enableBotPlayerLocked(playerID string) bool {
	session := game.playerSessions[playerID]
	if playerID == "" || session == nil {
		return false
	}

	config := runtime.DefaultCameraConfig()
	session.Config = config
	if ship := game.entities.Players[playerID]; ship != nil {
		ship.SetConfig(config)
		game.setPlayerCameraViewLocked(playerID, ship)
	}
	if cameraView := game.cameraViews[playerID]; cameraView != nil {
		cameraView.Config = config
	}

	game.botControllers[playerID] = bots.NewController()
	return true
}

func (game *Game) stepBots() {
	for playerID, controller := range game.botControllers {
		ship, ok := game.entities.Players[playerID]
		if !ok || ship == nil {
			session := game.playerSessions[playerID]
			if session == nil || !session.CanRespawn() {
				continue
			}
			game.respawnPlayer(playerID)
			ship = game.entities.Players[playerID]
		}
		if ship == nil || !game.playerCanReceiveInput(playerID, ship) {
			continue
		}

		cameraView := game.cameraViews[playerID]
		observation := bots.Observation{
			Position:  ship.Position(),
			Velocity:  ship.Velocity,
			Rotation:  ship.Rotation,
			Asteroids: make([]bots.AsteroidObservation, 0),
			Players:   make([]bots.PlayerObservation, 0),
		}
		if game.resolvedMatchRules.PlayerDamageEnabled {
			for targetID, target := range game.entities.Players {
				if target == nil || target.IsPendingDespawn() || !game.projectileCanDamagePlayerLocked(playerID, targetID) {
					continue
				}
				observation.Players = append(observation.Players, bots.PlayerObservation{Position: target.Position()})
			}
		}
		for _, asteroid := range game.entities.Asteroids {
			if asteroid == nil || asteroid.PendingDespawn {
				continue
			}
			if cameraView == nil || !isInsideCameraView(cameraView, asteroid.Position()) {
				continue
			}
			observation.Asteroids = append(observation.Asteroids, bots.AsteroidObservation{
				Position: physics.Vector2{X: asteroid.X, Y: asteroid.Y},
				Velocity: asteroid.Velocity,
				Size:     asteroid.Size,
			})
		}
		ship.SetInput(controller.Decide(observation))
	}
}
