package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func buildSessionLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) []RealtimeLaneCandidate {
	candidates := make([]RealtimeLaneCandidate, 0, 1)

	sessionStateLane, sessionSynced := state.LaneState(LaneSession)
	sessionReady := state.LaneBaselineReady(LaneSession)
	sessionSequence := NextLaneSequence(sessionStateLane, sessionSynced)
	sessionFull := BuildSessionFullPacket(snapshot, sessionSequence)
	quantizedSessionFull, err := quantizeSessionFullPacket(sessionFull)
	if err != nil {
		return candidates
	}
	sessionProjection, sessionHasProjection := state.BaselineProjection(LaneSession)
	sessionCanUseProjection := sessionReady && sessionSynced && sessionStateLane.IsFinalChunk && sessionStateLane.BaselineID != "" && sessionHasProjection
	if !sessionCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneSession, Kind: RealtimeLaneCandidateKindFull, Full: quantizedSessionFull, Projection: quantizedSessionFull})
	} else {
		previousSessionFull, ok := sessionProjection.(SessionWireFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneSession, Kind: RealtimeLaneCandidateKindFull, Full: quantizedSessionFull, Projection: quantizedSessionFull})
		} else {
			if !SessionWirePayloadChanged(previousSessionFull, quantizedSessionFull) {
				// No session candidate when the projection is unchanged.
			} else {
				sessionDelta := BuildSessionWireDeltaPacket(previousSessionFull, quantizedSessionFull)
				if SessionWireDeltaHasChanges(sessionDelta) {
					chainedSessionProjection := quantizedSessionFull
					chainedSessionProjection.Metadata = sessionDelta.Metadata
					candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneSession, Kind: RealtimeLaneCandidateKindDelta, Delta: sessionDelta, Projection: chainedSessionProjection})
				}
			}
		}
	}

	return candidates
}