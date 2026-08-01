package game

func (game *Game) playerSessionStateLocked(session *playerSession) PlayerSessionState {
	displayScore := session.Score
	if game.resolvedMatchRules.TeamScoreEnabled {
		displayScore = game.teamScoreLocked(session.TeamID)
	}
	return PlayerSessionState{
		ID:                  session.ID,
		ShipType:            session.ShipTypeID,
		Score:               displayScore,
		Lives:               session.Lives,
		RespawnCooldown:     session.RespawnCooldown,
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
