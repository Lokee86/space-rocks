package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	targetpolicy "github.com/Lokee86/space-rocks/server/internal/game/targeting"
)

func (game *Game) SelectTargetAtPosition(playerID string, x float64, y float64, target targetpolicy.TargetRef) bool {
	game.mu.Lock()
	defer game.mu.Unlock()

	if !game.playerExistsLocked(playerID) {
		return false
	}

	if target.IsEmpty() {
		return false
	}

	if !game.targetExists(target) {
		return false
	}

	candidates := game.targetCandidatesLocked()
	var matchedCandidate *targetpolicy.TargetCandidate
	for index := range candidates {
		candidate := &candidates[index]
		if candidate.Ref.Kind != target.Kind || candidate.Ref.ID != target.ID {
			continue
		}
		matchedCandidate = candidate
		break
	}
	if matchedCandidate == nil {
		return false
	}

	clickPoint := physics.Vector2{X: x, Y: y}
	if !physics.BodyContainsPoint(matchedCandidate.Body, clickPoint) {
		return false
	}

	player := game.state.Players[playerID]
	player.TargetKind = string(target.Kind)
	player.TargetID = target.ID
	if target.Kind == targetpolicy.TargetKindPlayer {
		player.TargetPlayerID = target.ID
	} else {
		player.TargetPlayerID = ""
	}

	return true
}

func (game *Game) SetTarget(playerID string, target targetpolicy.TargetRef) bool {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, exists := game.state.Players[playerID]
	if !exists || player == nil {
		return false
	}

	if target.IsEmpty() {
		player.TargetKind = ""
		player.TargetID = ""
		player.TargetPlayerID = ""
		return true
	}

	if !game.targetExists(target) {
		return false
	}

	player.TargetKind = string(target.Kind)
	player.TargetID = target.ID
	if target.Kind == targetpolicy.TargetKindPlayer {
		player.TargetPlayerID = target.ID
	} else {
		player.TargetPlayerID = ""
	}

	return true
}

func (game *Game) ClearTarget(playerID string) {
	game.SetTarget(playerID, targetpolicy.EmptyTarget())
}

func (game *Game) Target(playerID string) targetpolicy.TargetRef {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, exists := game.state.Players[playerID]
	if !exists || player == nil {
		return targetpolicy.EmptyTarget()
	}

	return targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKind(player.TargetKind),
		ID:   player.TargetID,
	}
}

func (game *Game) SetPlayerTarget(playerID string, targetPlayerID string) bool {
	if targetPlayerID == "" {
		if !game.playerExists(playerID) {
			return false
		}
		game.ClearTarget(playerID)
		return true
	}
	return game.SetTarget(playerID, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   targetPlayerID,
	})
}

func (game *Game) ClearPlayerTarget(playerID string) bool {
	if !game.playerExists(playerID) {
		return false
	}
	game.ClearTarget(playerID)
	return true
}

func (game *Game) PlayerTarget(playerID string) string {
	target := game.Target(playerID)
	if target.Kind != targetpolicy.TargetKindPlayer {
		return ""
	}

	return target.ID
}

func (game *Game) playerExistsLocked(playerID string) bool {
	player, exists := game.state.Players[playerID]
	return exists && player != nil
}

func (game *Game) playerExists(playerID string) bool {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.playerExistsLocked(playerID)
}

func (game *Game) clearTargetsForMissingPlayersLocked() {
	for _, player := range game.state.Players {
		if player == nil {
			continue
		}

		target := targetpolicy.TargetRef{
			Kind: targetpolicy.TargetKind(player.TargetKind),
			ID:   player.TargetID,
		}
		if target.IsEmpty() {
			continue
		}
		if game.targetExists(target) {
			continue
		}

		player.TargetKind = ""
		player.TargetID = ""
		player.TargetPlayerID = ""
	}
}

func (game *Game) targetExists(target targetpolicy.TargetRef) bool {
	switch target.Kind {
	case targetpolicy.TargetKindPlayer:
		player, exists := game.state.Players[target.ID]
		return exists && player != nil
	case targetpolicy.TargetKindEnemy:
		enemy, exists := game.state.Enemies[target.ID]
		return exists && enemy != nil
	case targetpolicy.TargetKindAsteroid:
		asteroid, exists := game.state.Asteroids[target.ID]
		return exists && asteroid != nil
	case targetpolicy.TargetKindBullet:
		bullet, exists := game.state.Projectiles[target.ID]
		return exists && bullet != nil
	default:
		return false
	}
}

func (game *Game) targetCandidatesLocked() []targetpolicy.TargetCandidate {
	candidates := make([]targetpolicy.TargetCandidate, 0)

	for playerID, player := range game.state.Players {
		if player == nil {
			continue
		}

		body, ok := player.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		candidates = append(candidates, targetpolicy.TargetCandidate{
			Ref: targetpolicy.TargetRef{
				Kind: targetpolicy.TargetKindPlayer,
				ID:   playerID,
			},
			Body: body,
		})
	}

	for asteroidID, asteroid := range game.state.Asteroids {
		if asteroid == nil {
			continue
		}

		body, ok := asteroid.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		candidates = append(candidates, targetpolicy.TargetCandidate{
			Ref: targetpolicy.TargetRef{
				Kind: targetpolicy.TargetKindAsteroid,
				ID:   asteroidID,
			},
			Body: body,
		})
	}

	for projectileID, projectile := range game.state.Projectiles {
		if projectile == nil {
			continue
		}

		body, ok := projectile.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		candidates = append(candidates, targetpolicy.TargetCandidate{
			Ref: targetpolicy.TargetRef{
				Kind: targetpolicy.TargetKindBullet,
				ID:   projectileID,
			},
			Body: body,
		})
	}

	for enemyID, enemy := range game.state.Enemies {
		if enemy == nil {
			continue
		}

		body, ok := enemy.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		candidates = append(candidates, targetpolicy.TargetCandidate{
			Ref: targetpolicy.TargetRef{
				Kind: targetpolicy.TargetKindEnemy,
				ID:   enemyID,
			},
			Body: body,
		})
	}

	return candidates
}
