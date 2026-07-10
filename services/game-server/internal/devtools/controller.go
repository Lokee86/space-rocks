package devtools

import streamruntime "github.com/Lokee86/space-rocks/server/internal/devtools/streamruntime"

type Dependencies struct {
	Target           Target
	Streams          *streamruntime.Runtime
	ObserverRegistry *ObserverRegistry
}

type Controller struct {
	target           Target
	streams          *streamruntime.Runtime
	observerRegistry *ObserverRegistry
}

var defaultObserverRegistry = NewObserverRegistry()

func NewController(deps Dependencies) *Controller {
	streams := deps.Streams
	if streams == nil {
		streams = streamruntime.DefaultRuntime
	}

	observerRegistry := deps.ObserverRegistry
	if observerRegistry == nil {
		observerRegistry = defaultObserverRegistry
	}

	return &Controller{
		target:           deps.Target,
		streams:          streams,
		observerRegistry: observerRegistry,
	}
}

func (controller *Controller) HandleCommand(playerID string, command DebugCommand) bool {
	if controller == nil || controller.target == nil {
		return false
	}

	switch command.Type {
	case PacketTypeToggleDebugInvincible:
		return handleToggleDebugInvincible(controller.target, playerID, command)
	case PacketTypeToggleDebugInfiniteLives:
		return handleToggleDebugInfiniteLives(controller.target, playerID, command)
	case PacketTypeToggleDebugFreezeWorld:
		return handleToggleDebugFreezeWorld(controller.target, playerID, command)
	case PacketTypeToggleDebugFreezePlayer:
		return handleToggleDebugFreezePlayer(controller.target, playerID, command)
	case PacketTypeDebugKillPlayer:
		return handleDebugKillPlayer(controller.target, playerID, command)
	case PacketTypeDebugSpawnEntity:
		return handleDebugSpawnEntity(controller.target, playerID, command)
	case PacketTypeDebugSpawnPickup:
		return handleDebugSpawnPickup(controller.target, playerID, command)
	case PacketTypeDebugBeginContinuousBulletStream:
		return handleDebugBeginContinuousBulletStream(controller, playerID, command)
	case PacketTypeDebugRespawnPlayer:
		return handleDebugRespawnPlayer(controller.target, playerID, command)
	case PacketTypeDebugSetScore:
		return handleDebugSetScore(controller.target, playerID, command)
	case PacketTypeDebugAddScore:
		return handleDebugAddScore(controller.target, playerID, command)
	case PacketTypeDebugSetLives:
		return handleDebugSetLives(controller.target, playerID, command)
	case PacketTypeDebugAddLives:
		return handleDebugAddLives(controller.target, playerID, command)
	case PacketTypeDebugClearBullets:
		return handleDebugClearBullets(controller, playerID, command)
	case PacketTypeDebugClearAsteroids:
		return handleDebugClearAsteroids(controller.target, playerID, command)
	default:
		return false
	}
}

func (controller *Controller) StatusFor(playerID string) DebugStatus {
	status := DebugStatus{
		WorldFrozen:      controller.target.WorldFrozen(),
		AsteroidsFrozen:  controller.target.AsteroidsFrozen(),
		BulletsFrozen:    controller.target.BulletsFrozen(),
		SpawningFrozen:   controller.target.SpawningFrozen(),
		CollisionsFrozen: controller.target.CollisionsFrozen(),
	}

	if invincible, ok := controller.target.PlayerInvincible(playerID); ok {
		status.Invincible = invincible
	}
	if infiniteLives, ok := controller.target.InfiniteLives(playerID); ok {
		status.InfiniteLives = infiniteLives
	}
	if playerFrozen, ok := controller.target.PlayerFrozen(playerID); ok {
		status.PlayerFrozen = playerFrozen
	}

	return status
}

func (controller *Controller) StatusesForAllPlayers() map[string]DebugStatus {
	statuses := make(map[string]DebugStatus)
	if controller == nil || controller.target == nil {
		return statuses
	}

	decision := controller.target.MatchDecision()
	for _, player := range decision.Players {
		statuses[player.ID] = controller.StatusFor(player.ID)
	}
	return statuses
}
