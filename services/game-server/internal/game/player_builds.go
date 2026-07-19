package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
)

func (game *Game) ApplyPlayerBuild(playerID string, build playerbuild.ResolvedPlayerBuild) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.lockedFinalMatchState != nil || game.matchElapsed > 0 {
		return fmt.Errorf("player builds cannot change after match start")
	}
	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return fmt.Errorf("player session %q does not exist", playerID)
	}
	if err := session.ApplyResolvedBuild(build); err != nil {
		return err
	}

	if current := game.entities.Players[playerID]; current != nil {
		replacement := session.NewShip(current.Position())
		replacement.Rotation = current.Rotation
		replacement.Velocity = current.Velocity
		replacement.Config = current.Config
		game.entities.Players[playerID] = replacement
		game.setPlayerCameraViewLocked(playerID, replacement)
	}
	game.publishPresentationFrameLocked()
	return nil
}

func (game *Game) PlayerResolvedBuild(playerID string) (playerbuild.ResolvedPlayerBuild, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()

	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return playerbuild.ResolvedPlayerBuild{}, false
	}
	return session.ResolvedBuild.Clone(), true
}
