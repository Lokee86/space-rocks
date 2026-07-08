package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func buildEventLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) []RealtimeLaneCandidate {
	if len(snapshot.PendingEvents) == 0 {
		return nil
	}

	eventState, _ := state.LaneState(LaneEvent)
	return []RealtimeLaneCandidate{
		{Lane: LaneEvent, Kind: RealtimeLaneCandidateKindEventBatch, Full: BuildEventBatchPacket(snapshot.PendingEvents, eventState.Sequence, snapshot.ServerSentMsec)},
	}
}