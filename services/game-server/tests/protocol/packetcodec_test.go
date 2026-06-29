package protocoltests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	realtime "github.com/Lokee86/space-rocks/server/internal/protocol/realtime"
)

func TestDecodeClientInputPacket(t *testing.T) {
	raw := []byte(`{"type":"input","input":{"forward":true,"left":true,"primary_fire":true,"secondary_fire":true}}`)

	var packet game.ClientPacket
	if err := packetcodec.Decode(raw, &packet); err != nil {
		t.Fatalf("decode client input packet: %v", err)
	}

	if packet.Type != game.PacketTypeInput {
		t.Fatalf("expected packet type %q, got %q", game.PacketTypeInput, packet.Type)
	}
	if !packet.Input.Forward {
		t.Fatal("expected forward input to decode")
	}
	if !packet.Input.Left {
		t.Fatal("expected left input to decode")
	}
	if !packet.Input.PrimaryFire {
		t.Fatal("expected primary_fire input to decode")
	}
	if !packet.Input.SecondaryFire {
		t.Fatal("expected secondary_fire input to decode")
	}
}

func TestEncodeDecodeLanePackets(t *testing.T) {
	tests := []struct {
		name  string
		packet any
	}{
		{
			name: "world_full",
			packet: realtime.WorldFullPacket{
				Type: realtime.PacketTypeWorldFull,
				Metadata: realtime.Metadata{Lane: realtime.Lane("world"), Sequence: 1, SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true},
				Ships: []realtime.WorldShipRecord{{ID: "player-1"}},
			},
		},
		{
			name: "world_delta",
			packet: realtime.WorldFullPacket{
				Type: realtime.PacketTypeWorldDelta,
				Metadata: realtime.Metadata{Lane: realtime.Lane("world"), Sequence: 2, SnapshotKind: realtime.SnapshotKind("delta"), IsFinalChunk: true},
				Ships: []realtime.WorldShipRecord{{ID: "player-1"}},
			},
		},
		{
			name: "overlay_full",
			packet: realtime.OverlayFullPacket{
				Type: realtime.PacketTypeOverlayFull,
				Metadata: realtime.Metadata{Lane: realtime.Lane("overlay"), Sequence: 1, SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true},
				Receiver: realtime.OverlayReceiverRecord{SelfID: "player-1"},
			},
		},
		{
			name: "overlay_delta",
			packet: realtime.OverlayFullPacket{
				Type: realtime.PacketTypeOverlayDelta,
				Metadata: realtime.Metadata{Lane: realtime.Lane("overlay"), Sequence: 2, SnapshotKind: realtime.SnapshotKind("delta"), IsFinalChunk: true},
				Receiver: realtime.OverlayReceiverRecord{SelfID: "player-1"},
			},
		},
		{
			name: "session_full",
			packet: realtime.SessionFullPacket{
				Type: realtime.PacketTypeSessionFull,
				Metadata: realtime.Metadata{Lane: realtime.Lane("session"), Sequence: 1, SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true},
				Players: []realtime.SessionPlayerRecord{{ID: "player-1"}},
			},
		},
		{
			name: "session_delta",
			packet: realtime.SessionFullPacket{
				Type: realtime.PacketTypeSessionDelta,
				Metadata: realtime.Metadata{Lane: realtime.Lane("session"), Sequence: 2, SnapshotKind: realtime.SnapshotKind("delta"), IsFinalChunk: true},
				Players: []realtime.SessionPlayerRecord{{ID: "player-1"}},
			},
		},
		{
			name: "event_batch",
			packet: realtime.EventBatchPacket{
				Type: realtime.PacketTypeEventBatch,
				Metadata: realtime.Metadata{Lane: realtime.Lane("event"), Sequence: 1, SnapshotKind: realtime.SnapshotKind("batch"), IsFinalChunk: true},
				Batch: realtime.EventBatchRecord{Events: []realtime.EventRecord{{EventID: "event-1"}}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := packetcodec.Encode(tc.packet)
			if err != nil {
				t.Fatalf("encode %s packet: %v", tc.name, err)
			}
			if len(encoded) == 0 {
				t.Fatalf("expected encoded %s packet bytes", tc.name)
			}

			switch want := tc.packet.(type) {
			case realtime.WorldFullPacket:
				var decoded realtime.WorldFullPacket
				if err := packetcodec.Decode(encoded, &decoded); err != nil {
					t.Fatalf("decode %s packet: %v", tc.name, err)
				}
				if decoded.Type != want.Type || decoded.Metadata != want.Metadata || len(decoded.Ships) != len(want.Ships) {
					t.Fatalf("expected %s packet to round-trip, got %+v", tc.name, decoded)
				}
			case realtime.OverlayFullPacket:
				var decoded realtime.OverlayFullPacket
				if err := packetcodec.Decode(encoded, &decoded); err != nil {
					t.Fatalf("decode %s packet: %v", tc.name, err)
				}
				if decoded.Type != want.Type || decoded.Metadata != want.Metadata || decoded.Receiver.SelfID != want.Receiver.SelfID {
					t.Fatalf("expected %s packet to round-trip, got %+v", tc.name, decoded)
				}
			case realtime.SessionFullPacket:
				var decoded realtime.SessionFullPacket
				if err := packetcodec.Decode(encoded, &decoded); err != nil {
					t.Fatalf("decode %s packet: %v", tc.name, err)
				}
				if decoded.Type != want.Type || decoded.Metadata != want.Metadata || len(decoded.Players) != len(want.Players) {
					t.Fatalf("expected %s packet to round-trip, got %+v", tc.name, decoded)
				}
			case realtime.EventBatchPacket:
				var decoded realtime.EventBatchPacket
				if err := packetcodec.Decode(encoded, &decoded); err != nil {
					t.Fatalf("decode %s packet: %v", tc.name, err)
				}
				if decoded.Type != want.Type || decoded.Metadata != want.Metadata || len(decoded.Batch.Events) != len(want.Batch.Events) {
					t.Fatalf("expected %s packet to round-trip, got %+v", tc.name, decoded)
				}
			default:
				t.Fatalf("unexpected packet type %T", tc.packet)
			}
		})
	}
}
