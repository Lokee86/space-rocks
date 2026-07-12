package realtime

func quantizeSessionFullPacket(packet SessionFullPacket) (SessionWireFullPacket, error) {
	quantized := SessionWireFullPacket{
		Type:            packet.Type,
		Metadata:        packet.Metadata,
		Players:         make([]SessionPlayerWireRecord, 0, len(packet.Players)),
		PlayerLifecycle: packet.PlayerLifecycle,
		TotalAsteroids:  packet.TotalAsteroids,
	}
	var err error
	for _, player := range packet.Players {
		wirePlayer := SessionPlayerWireRecord{
			ID:                  player.ID,
			ShipType:            player.ShipType,
			Score:               player.Score,
			Lives:               player.Lives,
			PrimaryWeaponID:     player.PrimaryWeaponID,
			PrimaryAmmoPolicy:   player.PrimaryAmmoPolicy,
			SecondaryWeaponID:   player.SecondaryWeaponID,
			SecondaryAmmoPolicy: player.SecondaryAmmoPolicy,
		}
		wirePlayer.RespawnCooldown, err = quantizeTypedFloat("session.players.respawn_cooldown", player.RespawnCooldown)
		if err != nil {
			return SessionWireFullPacket{}, err
		}
		wirePlayer.SpawnX, err = quantizeTypedFloat("session.players.spawn_x", player.SpawnX)
		if err != nil {
			return SessionWireFullPacket{}, err
		}
		wirePlayer.SpawnY, err = quantizeTypedFloat("session.players.spawn_y", player.SpawnY)
		if err != nil {
			return SessionWireFullPacket{}, err
		}
		quantized.Players = append(quantized.Players, wirePlayer)
	}
	return quantized, nil
}
