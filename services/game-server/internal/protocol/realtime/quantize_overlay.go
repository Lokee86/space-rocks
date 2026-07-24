package realtime

func quantizeOverlayFullPacket(packet OverlayFullPacket) (OverlayWireFullPacket, error) {
	quantized := OverlayWireFullPacket{
		Type:     packet.Type,
		Metadata: packet.Metadata,
		Receiver: OverlayReceiverWireRecord{
			SelfID:                 packet.Receiver.SelfID,
			Lives:                  packet.Receiver.Lives,
			Score:                  packet.Receiver.Score,
			Health:                 packet.Receiver.Health,
			MaxHealth:              packet.Receiver.MaxHealth,
			Shields:                packet.Receiver.Shields,
			MaxShields:             packet.Receiver.MaxShields,
			ShieldModuleID:         packet.Receiver.ShieldModuleID,
			ArmorModuleID:          packet.Receiver.ArmorModuleID,
			EngineModuleID:         packet.Receiver.EngineModuleID,
			UtilityModuleID:        packet.Receiver.UtilityModuleID,
			PrimaryWeaponID:        packet.Receiver.PrimaryWeaponID,
			PrimaryAmmoPolicy:      packet.Receiver.PrimaryAmmoPolicy,
			PrimaryAmmoRemaining:   packet.Receiver.PrimaryAmmoRemaining,
			SecondaryWeaponID:      packet.Receiver.SecondaryWeaponID,
			SecondaryAmmoPolicy:    packet.Receiver.SecondaryAmmoPolicy,
			SecondaryAmmoRemaining: packet.Receiver.SecondaryAmmoRemaining,
		},
	}
	var err error
	quantized.Receiver.RespawnCooldown, err = quantizeTypedFloat("overlay.respawn_cooldown", packet.Receiver.RespawnCooldown)
	if err != nil {
		return OverlayWireFullPacket{}, err
	}
	quantized.Receiver.PrimaryCooldownRemaining, err = quantizeTypedFloat("overlay.primary_cooldown_remaining", packet.Receiver.PrimaryCooldownRemaining)
	if err != nil {
		return OverlayWireFullPacket{}, err
	}
	quantized.Receiver.SecondaryCooldownRemaining, err = quantizeTypedFloat("overlay.secondary_cooldown_remaining", packet.Receiver.SecondaryCooldownRemaining)
	if err != nil {
		return OverlayWireFullPacket{}, err
	}
	return quantized, nil
}
