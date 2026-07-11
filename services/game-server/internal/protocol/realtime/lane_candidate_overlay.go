package realtime

import (
	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func buildOverlayLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) []RealtimeLaneCandidate {
	candidates := make([]RealtimeLaneCandidate, 0, 1)

	overlayState, overlaySynced := state.LaneState(LaneOverlay)
	overlayReady := state.LaneBaselineReady(LaneOverlay)
	overlaySequence := NextLaneSequence(overlayState, overlaySynced)
	overlayFull := BuildOverlayFullPacket(snapshot, state.ReceiverID, overlaySequence)
	quantizedOverlayFull, err := quantizeOverlayFullPacket(overlayFull)
	if err != nil {
		return candidates
	}
	overlayProjection, overlayHasProjection := state.BaselineProjection(LaneOverlay)
	overlayCanUseProjection := overlayReady && overlaySynced && overlayState.IsFinalChunk && overlayState.BaselineID != "" && overlayHasProjection
	if !overlayCanUseProjection {
		candidates = append(candidates, mustRealtimeLaneCandidate(quantizedOverlayFull, quantizedOverlayFull))
	} else {
		previousOverlayFull, ok := overlayProjection.(OverlayWireFullPacket)
		if !ok {
			candidates = append(candidates, mustRealtimeLaneCandidate(quantizedOverlayFull, quantizedOverlayFull))
		} else {
			if !OverlayWirePayloadChanged(previousOverlayFull, quantizedOverlayFull) {
				// No overlay candidate when the projection is unchanged.
			} else {
				overlayDelta := BuildOverlayWireDeltaPacket(previousOverlayFull, quantizedOverlayFull)
				if OverlayWireDeltaHasChanges(overlayDelta) {
					chainedOverlayProjection := quantizedOverlayFull
					chainedOverlayProjection.Metadata = overlayDelta.Metadata
					candidates = append(candidates, mustRealtimeLaneCandidate(overlayDelta, chainedOverlayProjection))
				}
			}
		}
	}

	return candidates
}