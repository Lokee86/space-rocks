package game

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

func (target *Control) CollisionBodiesByKind() map[string][]physics.CollisionBody {
	bodiesByKind := make(map[string][]physics.CollisionBody, 4)

	for _, player := range target.game.entities.Players {
		body, ok := player.CollisionBody(target.game.collisionShapes)
		if !ok {
			continue
		}
		bodiesByKind["player"] = append(bodiesByKind["player"], body)
	}
	for _, asteroid := range target.game.entities.Asteroids {
		body, ok := asteroid.CollisionBody(target.game.collisionShapes)
		if !ok {
			continue
		}
		bodiesByKind["asteroid"] = append(bodiesByKind["asteroid"], body)
	}
	for _, bullet := range target.game.entities.Projectiles {
		body, ok := bullet.CollisionBody(target.game.collisionShapes)
		if !ok {
			continue
		}
		bodiesByKind["bullet"] = append(bodiesByKind["bullet"], body)
	}
	for _, pickup := range target.game.entities.Pickups {
		if pickup == nil {
			continue
		}
		body, ok := pickup.CollisionBody(target.game.collisionShapes)
		if !ok {
			continue
		}
		bodiesByKind["pickup"] = append(bodiesByKind["pickup"], body)
	}

	return bodiesByKind
}
