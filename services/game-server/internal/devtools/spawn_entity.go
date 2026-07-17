package devtools

func handleDebugSpawnEntity(target Target, playerID string, command DebugCommand) bool {
	request := SpawnEntityRequest{
		EntityType:     command.EntityType,
		X:              command.X,
		Y:              command.Y,
		HasDirection:   command.HasDirection,
		DirectionX:     command.DirectionX,
		DirectionY:     command.DirectionY,
		TargetPlayerID: command.TargetPlayerID,
	}
	if request.EntityType == EntityTypePlayer {
		_, _, ok := applyDebugSpawnPlayer(target, request)
		if !ok {
			return true
		}
		return true
	}
	if request.EntityType == EntityTypeBullet {
		_, ok := applyDebugSpawnBullet(target, playerID, request)
		if !ok {
			return true
		}
		return true
	}
	if request.EntityType == EntityTypeAsteroid {
		_, _, ok := applyDebugSpawnAsteroid(target, request)
		if !ok {
			return true
		}
		return true
	}
	return true
}
