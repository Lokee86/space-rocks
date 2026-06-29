package realtime

import game "github.com/Lokee86/space-rocks/server/internal/game"

func CompareShadowRealtimeCoverage(snapshot game.GameplayPresentationSnapshot, world WorldFullPacket, overlay OverlayFullPacket, session SessionFullPacket, events EventBatchPacket) []string {
	issues := make([]string, 0)

	if overlay.Receiver.SelfID != snapshot.SelfID {
		issues = append(issues, "self_id not covered by overlay_full")
	}
	if overlay.Receiver.Lives != snapshot.Lives {
		issues = append(issues, "lives not covered by overlay_full")
	}
	if session.TotalAsteroids != snapshot.TotalAsteroids {
		issues = append(issues, "total_asteroids not covered by session_full")
	}
	if world.Metadata.ServerSentMsec != snapshot.ServerSentMsec || overlay.Metadata.ServerSentMsec != snapshot.ServerSentMsec || session.Metadata.ServerSentMsec != snapshot.ServerSentMsec || events.Metadata.ServerSentMsec != snapshot.ServerSentMsec {
		issues = append(issues, "server_sent_msec not propagated to metadata")
	}

	for id, player := range snapshot.Players {
		found := false
		for _, worldPlayer := range world.Ships {
			if worldPlayer.ID == id {
				found = true
				if worldPlayer.ShipType != player.ShipType || worldPlayer.X != player.X || worldPlayer.Y != player.Y || worldPlayer.Rotation != player.Rotation || worldPlayer.Health != player.Health || worldPlayer.Shields != player.Shields || worldPlayer.Thrusting != player.Thrusting || worldPlayer.TargetKind != player.TargetKind || worldPlayer.TargetID != player.TargetID {
					issues = append(issues, "players split by field mismatch in world_full")
				}
			}
		}
		if !found {
			issues = append(issues, "player missing from world_full")
		}
	}

	if len(session.Players) != len(snapshot.PlayerSessions) {
		issues = append(issues, "player_sessions not covered by session_full")
	}
	if len(session.PlayerLifecycle) != len(snapshot.PlayerLifecycle) {
		issues = append(issues, "player_lifecycle not covered by session_full")
	}
	if len(world.Bullets) != len(snapshot.Bullets) {
		issues = append(issues, "bullets not covered by world_full")
	}
	if len(world.Asteroids) != len(snapshot.Asteroids) {
		issues = append(issues, "asteroids not covered by world_full")
	}
	if len(world.Pickups) != len(snapshot.Pickups) {
		issues = append(issues, "pickups not covered by world_full")
	}
	if len(events.Batch.Events) != len(snapshot.PendingEvents) {
		issues = append(issues, "events not covered by event_batch")
	}

	return issues
}
