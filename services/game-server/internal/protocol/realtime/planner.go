package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

type RealtimeLaneCandidateKind string

const (
	RealtimeLaneCandidateKindFull       RealtimeLaneCandidateKind = "full"
	RealtimeLaneCandidateKindDelta      RealtimeLaneCandidateKind = "delta"
	RealtimeLaneCandidateKindEventBatch RealtimeLaneCandidateKind = "event_batch"
)

type RealtimeLaneCandidate struct {
	Lane       Lane
	Kind       RealtimeLaneCandidateKind
	Full       any
	Projection any
	Delta      any
}

type RealtimeLanePlan struct {
	Candidates []RealtimeLaneCandidate
}

type RealtimeSendPrepared struct {
	CandidatePlan RealtimeLanePlan
	Records       []ScheduleRecord
	SendPlan      SendPlan
}

func AssembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) RealtimeLanePlan {
	candidates := make([]RealtimeLaneCandidate, 0, 4)


	worldState, worldSynced := state.LaneState(LaneWorld)
	worldReady := state.LaneBaselineReady(LaneWorld)
	worldSequence := NextLaneSequence(worldState, worldSynced)
	worldFull := BuildWorldFullPacket(snapshot, worldSequence)
	quantizedWorldFull, err := quantizeWorldFullPacket(worldFull)
	if err != nil {
		return RealtimeLanePlan{Candidates: candidates}
	}
	worldProjection, worldHasProjection := state.BaselineProjection(LaneWorld)
	worldCanUseProjection := worldReady && worldSynced && worldState.IsFinalChunk && worldState.BaselineID != "" && worldHasProjection
if !worldCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:       LaneWorld,
			Kind:       RealtimeLaneCandidateKindFull,
			Full:       quantizedWorldFull,
			Projection: quantizedWorldFull,
		})
	} else {
		previousWorldFull, ok := worldProjection.(WorldWireFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{
				Lane:       LaneWorld,
				Kind:       RealtimeLaneCandidateKindFull,
				Full:       quantizedWorldFull,
				Projection: quantizedWorldFull,
			})
		} else if !ProjectionChanged(previousWorldFull, quantizedWorldFull) {
			// No world candidate when the projection is unchanged.
		} else {
			worldDelta := BuildWorldWireDeltaPacket(previousWorldFull, quantizedWorldFull)
			if WorldWireDeltaHasChanges(worldDelta) {
				candidates = append(candidates, RealtimeLaneCandidate{
					Lane:       LaneWorld,
					Kind:       RealtimeLaneCandidateKindDelta,
					Delta:      worldDelta,
					Projection: quantizedWorldFull,
				})
			}
		}
	}

	overlayState, overlaySynced := state.LaneState(LaneOverlay)
	overlayReady := state.LaneBaselineReady(LaneOverlay)
	overlaySequence := NextLaneSequence(overlayState, overlaySynced)
	overlayFull := BuildOverlayFullPacket(snapshot, state.ReceiverID, overlaySequence)
	quantizedOverlayFull, err := quantizeOverlayFullPacket(overlayFull)
	if err != nil {
		return RealtimeLanePlan{Candidates: candidates}
	}
	overlayProjection, overlayHasProjection := state.BaselineProjection(LaneOverlay)
	overlayCanUseProjection := overlayReady && overlaySynced && overlayState.IsFinalChunk && overlayState.BaselineID != "" && overlayHasProjection
	if !overlayCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:       LaneOverlay,
			Kind:       RealtimeLaneCandidateKindFull,
			Full:       quantizedOverlayFull,
			Projection: quantizedOverlayFull,
		})
	} else {
		previousOverlayFull, ok := overlayProjection.(OverlayWireFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{
				Lane:       LaneOverlay,
				Kind:       RealtimeLaneCandidateKindFull,
				Full:       quantizedOverlayFull,
				Projection: quantizedOverlayFull,
			})
		} else {
			if !ProjectionChanged(previousOverlayFull, quantizedOverlayFull) {
				// No overlay candidate when the projection is unchanged.
			} else {
				overlayDelta := BuildOverlayWireDeltaPacket(previousOverlayFull, quantizedOverlayFull)
				if OverlayWireDeltaHasChanges(overlayDelta) {
					candidates = append(candidates, RealtimeLaneCandidate{
					Lane:       LaneOverlay,
					Kind:       RealtimeLaneCandidateKindDelta,
					Delta:      overlayDelta,
					Projection: quantizedOverlayFull,
				})
				}
			}
		}
	}

	sessionState, sessionSynced := state.LaneState(LaneSession)
	sessionReady := state.LaneBaselineReady(LaneSession)
	sessionSequence := NextLaneSequence(sessionState, sessionSynced)
	sessionFull := BuildSessionFullPacket(snapshot, sessionSequence)
	quantizedSessionFull, err := quantizeSessionFullPacket(sessionFull)
	if err != nil {
		return RealtimeLanePlan{Candidates: candidates}
	}
	sessionProjection, sessionHasProjection := state.BaselineProjection(LaneSession)
	sessionCanUseProjection := sessionReady && sessionSynced && sessionState.IsFinalChunk && sessionState.BaselineID != "" && sessionHasProjection
	if !sessionCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:       LaneSession,
			Kind:       RealtimeLaneCandidateKindFull,
			Full:       quantizedSessionFull,
			Projection: quantizedSessionFull,
		})
	} else {
		previousSessionFull, ok := sessionProjection.(SessionWireFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{
				Lane:       LaneSession,
				Kind:       RealtimeLaneCandidateKindFull,
				Full:       quantizedSessionFull,
				Projection: quantizedSessionFull,
			})
		} else {
			if !ProjectionChanged(previousSessionFull, quantizedSessionFull) {
				// No session candidate when the projection is unchanged.
			} else {
				sessionDelta := BuildSessionWireDeltaPacket(previousSessionFull, quantizedSessionFull)
				if SessionWireDeltaHasChanges(sessionDelta) {
					candidates = append(candidates, RealtimeLaneCandidate{
					Lane:       LaneSession,
					Kind:       RealtimeLaneCandidateKindDelta,
					Delta:      sessionDelta,
					Projection: quantizedSessionFull,
				})
				}
			}
		}
	}

	if len(snapshot.PendingEvents) > 0 {
		eventState, _ := state.LaneState(LaneEvent)
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneEvent,
			Kind: RealtimeLaneCandidateKindEventBatch,
			Full: BuildEventBatchPacket(snapshot.PendingEvents, eventState.Sequence, snapshot.ServerSentMsec),
		})
	}

	return RealtimeLanePlan{Candidates: candidates}
}

func prepareRealtimeSendPlan(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) RealtimeSendPrepared {
	candidatePlan := AssembleRealtimeLaneCandidates(snapshot, state)

	records := make([]ScheduleRecord, 0, len(candidatePlan.Candidates))
	for i, candidate := range candidatePlan.Candidates {
		records = append(records, scheduleRecordForCandidate(i, candidate))
	}

	return RealtimeSendPrepared{
		CandidatePlan: candidatePlan,
		Records:       records,
		SendPlan:      SelectSendPlan(records),
	}
}

func packetFamilyForCandidate(candidate RealtimeLaneCandidate) string {
	switch candidate.Lane {
	case LaneWorld:
		switch candidate.Kind {
		case RealtimeLaneCandidateKindFull:
			return PacketFamilyWorldFull
		case RealtimeLaneCandidateKindDelta:
			return PacketFamilyWorldDelta
		}
	case LaneOverlay:
		switch candidate.Kind {
		case RealtimeLaneCandidateKindFull:
			return PacketFamilyOverlayFull
		case RealtimeLaneCandidateKindDelta:
			return PacketFamilyOverlayDelta
		}
	case LaneSession:
		switch candidate.Kind {
		case RealtimeLaneCandidateKindFull:
			return PacketFamilySessionFull
		case RealtimeLaneCandidateKindDelta:
			return PacketFamilySessionDelta
		}
	case LaneEvent:
		if candidate.Kind == RealtimeLaneCandidateKindEventBatch {
			return PacketFamilyEventBatch
		}
	}

	return ""
}

func deliveryClassForCandidate(candidate RealtimeLaneCandidate) DeliveryClass {
	switch candidate.Kind {
	case RealtimeLaneCandidateKindEventBatch:
		return DeliveryClassEventOnce
	case RealtimeLaneCandidateKindDelta:
		switch candidate.Lane {
		case LaneSession:
			return DeliveryClassDeferrable
		case LaneWorld, LaneOverlay:
			return DeliveryClassHotSupersedable
		}
	default:
		return DeliveryClassRequired
	}

	return DeliveryClassRequired
}

func priorityForCandidate(candidate RealtimeLaneCandidate) Priority {
	switch candidate.Kind {
	case RealtimeLaneCandidateKindEventBatch, RealtimeLaneCandidateKindFull:
		return PriorityCritical
	case RealtimeLaneCandidateKindDelta:
		switch candidate.Lane {
		case LaneSession:
			return PriorityMedium
		case LaneWorld, LaneOverlay:
			return PriorityHigh
		}
	}

	return PriorityCritical
}

func scheduleRecordForCandidate(candidateIndex int, candidate RealtimeLaneCandidate) ScheduleRecord {
	packetFamily := packetFamilyForCandidate(candidate)
	record := ScheduleRecord{
		Lane:           candidate.Lane,
		CandidateIndex: candidateIndex,
		PacketFamily:   packetFamily,
		RecordKind:     string(candidate.Kind),
		Priority:       priorityForCandidate(candidate),
		DeliveryClass:  deliveryClassForCandidate(candidate),
		EstimatedBytes: EstimatePacketBytes(packetFamily, 1, 0),
		ChunkCount:     1,
		IsFinalChunk:   true,
	}

	switch candidate.Kind {
	case RealtimeLaneCandidateKindFull, RealtimeLaneCandidateKindEventBatch:
		record.PayloadRef = candidate.Full
	case RealtimeLaneCandidateKindDelta:
		record.PayloadRef = candidate.Delta
	}

	return record
}

func CandidateProjection(candidate RealtimeLaneCandidate) (any, bool) {
	if candidate.Kind == RealtimeLaneCandidateKindEventBatch {
		return nil, false
	}
	if candidate.Projection == nil {
		return nil, false
	}
	return candidate.Projection, true
}


func quantizeOverlayFullPacket(packet OverlayFullPacket) (OverlayWireFullPacket, error) {
	quantized := OverlayWireFullPacket{
		Type: packet.Type,
		Metadata: packet.Metadata,
		Receiver: OverlayReceiverWireRecord{
			SelfID:                  packet.Receiver.SelfID,
			Lives:                   packet.Receiver.Lives,
			Score:                   packet.Receiver.Score,
			PrimaryWeaponID:         packet.Receiver.PrimaryWeaponID,
			PrimaryAmmoPolicy:       packet.Receiver.PrimaryAmmoPolicy,
			PrimaryAmmoRemaining:    packet.Receiver.PrimaryAmmoRemaining,
			SecondaryWeaponID:       packet.Receiver.SecondaryWeaponID,
			SecondaryAmmoPolicy:     packet.Receiver.SecondaryAmmoPolicy,
			SecondaryAmmoRemaining:  packet.Receiver.SecondaryAmmoRemaining,
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

func quantizeSessionFullPacket(packet SessionFullPacket) (SessionWireFullPacket, error) {
	quantized := SessionWireFullPacket{
		Type: packet.Type,
		Metadata: packet.Metadata,
		Players: make([]SessionPlayerWireRecord, 0, len(packet.Players)),
		PlayerLifecycle: packet.PlayerLifecycle,
		TotalAsteroids: packet.TotalAsteroids,
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


