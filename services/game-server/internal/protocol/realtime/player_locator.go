package realtime

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

type PlayerLocatorProjection struct {
	Records []game.PlayerLocatorState
}

type PlayerLocatorPacket struct {
	Type           string
	Metadata       Metadata
	PlayerLocators []game.PlayerLocatorState
}

func (packet PlayerLocatorPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet PlayerLocatorPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (packet PlayerLocatorPacket) PacketFamily() string { return PacketFamilyPlayerLocator }
func (packet PlayerLocatorPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet PlayerLocatorPacket) WirePacket() map[string]any {
	return map[string]any{
		"type":             packet.Type,
		"lane":             packet.Metadata.Lane,
		"match_id":         packet.Metadata.MatchID,
		"sequence":         packet.Metadata.Sequence,
		"snapshot_id":      packet.Metadata.SnapshotID,
		"server_sent_msec": packet.Metadata.ServerSentMsec,
		"snapshot_kind":    packet.Metadata.SnapshotKind,
		"chunk_index":      packet.Metadata.ChunkIndex,
		"chunk_count":      packet.Metadata.ChunkCount,
		"is_final_chunk":   packet.Metadata.IsFinalChunk,
		"player_locators":  packet.PlayerLocators,
	}
}
func (PlayerLocatorPacket) realtimeLanePayload() {}

func buildPlayerLocatorCandidate(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) []RealtimeLaneCandidate {
	projection := projectPlayerLocators(snapshot.PlayerLocators)
	if len(projection.Records) == 0 {
		return nil
	}
	stored, hasStored := state.PacketProjection(PacketFamilyPlayerLocator)
	membershipChanged := true
	if hasStored {
		if previous, ok := stored.(PlayerLocatorProjection); ok {
			membershipChanged = playerLocatorMembershipChanged(previous, projection)
		}
	}
	cadence := constants.PlayerLocatorCadenceTicks
	if cadence < 1 {
		cadence = 1
	}
	if hasStored && !membershipChanged && state.HotLaneTick%cadence != 1 {
		return nil
	}

	sequence := state.PacketSequence(PacketFamilyPlayerLocator) + 1
	metadata := Metadata{
		MatchID:        state.MatchID,
		Lane:           LaneShips,
		Sequence:       sequence,
		SnapshotID:     DeltaSnapshotID(LaneShips, sequence),
		ServerSentMsec: snapshot.ServerSentMsec,
		SnapshotKind:   SnapshotKind("delta"),
		ChunkIndex:     0,
		ChunkCount:     1,
		IsFinalChunk:   true,
	}
	packet := PlayerLocatorPacket{
		Type:           PacketFamilyPlayerLocator,
		Metadata:       metadata,
		PlayerLocators: projection.Records,
	}
	return []RealtimeLaneCandidate{mustRealtimeLaneCandidate(packet, projection)}
}

func projectPlayerLocators(locators map[string]game.PlayerLocatorState) PlayerLocatorProjection {
	ids := make([]string, 0, len(locators))
	for id := range locators {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	records := make([]game.PlayerLocatorState, 0, len(ids))
	for _, id := range ids {
		record := locators[id]
		record.ID = id
		records = append(records, record)
	}
	return PlayerLocatorProjection{Records: records}
}

func playerLocatorMembershipChanged(previous PlayerLocatorProjection, current PlayerLocatorProjection) bool {
	if len(previous.Records) != len(current.Records) {
		return true
	}
	for index := range current.Records {
		if previous.Records[index].ID != current.Records[index].ID || previous.Records[index].Active != current.Records[index].Active {
			return true
		}
	}
	return false
}
