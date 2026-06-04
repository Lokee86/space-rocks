package game

func (game *Game) playerSessionStateLocked(session *playerSession) PlayerSessionState {
	return PlayerSessionState{
		ID:               session.ID,
		ShipType:         session.ShipTypeID,
		Score:            session.Score,
		Lives:            session.Lives,
		RespawnCooldown:  session.RespawnCooldown,
		SpawnX:           session.SpawnPosition.X,
		SpawnY:           session.SpawnPosition.Y,
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
