package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"

func (game *Game) playerSessionStateLocked(session *playerSession) PlayerSessionState {
	lives := 0
	respawnCooldown := 0.0
	if state, ok := game.lifeRuntime.ParticipantSnapshot(session.ID); ok {
		lives, _ = game.lifeRuntime.ProjectedLives(session.ID)
		respawnCooldown = state.RespawnCooldown
	}
	return PlayerSessionState{
		ID:                  session.ID,
		ShipType:            session.ShipTypeID,
		Score:               session.Score,
		Lives:               lives,
		RespawnCooldown:     respawnCooldown,
		SpawnX:              session.SpawnPosition.X,
		SpawnY:              session.SpawnPosition.Y,
		PrimaryWeaponID:     string(session.PlayerArmory.Primary.ID),
		PrimaryAmmoPolicy:   string(session.PlayerArmory.Primary.AmmoPolicy),
		SecondaryWeaponID:   string(session.PlayerArmory.Secondary.ID),
		SecondaryAmmoPolicy: string(session.PlayerArmory.Secondary.AmmoPolicy),
		MaxHealth:           session.Stats.MaxHealth,
		MaxShields:          session.Stats.MaxShields,
		ShieldModuleID:      resolvedModuleCatalogID(session, playerbuild.ShieldMod),
		ArmorModuleID:       resolvedModuleCatalogID(session, playerbuild.ArmorMod),
		EngineModuleID:      resolvedModuleCatalogID(session, playerbuild.EngineMod),
		UtilityModuleID:     resolvedModuleCatalogID(session, playerbuild.UtilityMod),
	}
}

func resolvedModuleCatalogID(session *playerSession, slot playerbuild.ModuleSlot) string {
	module, ok := session.ResolvedBuild.EquippedModules[slot]
	if !ok {
		return ""
	}
	return module.CatalogID
}

func (game *Game) playerSessionStatesLocked() map[string]PlayerSessionState {
	playerSessions := make(map[string]PlayerSessionState, len(game.playerSessions))
	for playerID, session := range game.playerSessions {
		if session == nil {
			continue
		}
		playerSessions[playerID] = game.playerSessionStateLocked(session)
	}
	return playerSessions
}
