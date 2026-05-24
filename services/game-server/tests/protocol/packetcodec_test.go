package protocoltests

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func TestDecodeClientInputPacket(t *testing.T) {
	raw := []byte(`{"type":"input","input":{"forward":true,"left":true,"shoot":true}}`)

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
	if !packet.Input.Shoot {
		t.Fatal("expected shoot input to decode")
	}
}

func TestEncodeStatePacket(t *testing.T) {
	raw, err := packetcodec.Encode(game.StatePacket{Type: game.PacketTypeState})
	if err != nil {
		t.Fatalf("encode state packet: %v", err)
	}

	var packet struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &packet); err != nil {
		t.Fatalf("decode encoded state packet: %v", err)
	}
	if packet.Type != game.PacketTypeState {
		t.Fatalf("expected packet type %q, got %q", game.PacketTypeState, packet.Type)
	}
}
