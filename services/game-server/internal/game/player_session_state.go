package game

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
	}
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
