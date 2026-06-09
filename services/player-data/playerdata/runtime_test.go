package playerdata

import (
	"bytes"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestNewRuntimeRejectsNilStore(t *testing.T) {
	if _, err := NewRuntime(Config{}); err == nil {
		t.Fatal("NewRuntime returned nil error for nil store")
	}
}

func TestRuntimeHandleDelegatesLoadStats(t *testing.T) {
	store := NewMemoryStore()
	identity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}
	if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      8,
		ShipDeaths: 1,
		Won:        true,
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	runtime, err := NewRuntime(Config{Store: store})
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	got, err := runtime.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	want, err := NewDispatcher(store).Handle(payload)
	if err != nil {
		t.Fatalf("dispatcher handle returned error: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("Handle() = %s, want %s", got, want)
	}
}
