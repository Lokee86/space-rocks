package matchreporting

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	serverplayerdata "github.com/Lokee86/space-rocks/server/internal/playerdata"
)

type fakePlayerDataSink struct {
	payloads      [][]byte
	response      []byte
	responseErr    error
}

func (s *fakePlayerDataSink) HandlePlayerDataCommand(payload []byte) ([]byte, error) {
	s.payloads = append(s.payloads, append([]byte(nil), payload...))
	if s.responseErr != nil {
		return nil, s.responseErr
	}
	return append([]byte(nil), s.response...), nil
}

func encodeRecordMatchResultResponse(t *testing.T, accepted bool, duplicate bool) []byte {
	t.Helper()

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResultResult{
		Type:      protocol.PacketTypePlayerDataRecordMatchResultResult,
		Accepted:  accepted,
		Duplicate: duplicate,
	})
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}

	return payload
}

func decodeRecordMatchResultCommand(t *testing.T, payload []byte) protocol.PlayerDataRecordMatchResult {
	t.Helper()

	var command protocol.PlayerDataRecordMatchResult
	if err := json.Unmarshal(payload, &command); err != nil {
		t.Fatalf("decode command: %v", err)
	}

	return command
}

func testReporterSummary() serverplayerdata.MatchResultSummary {
	return serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
				AccountID:    "acct-1",
				Score:        10,
				ShipDeaths:   1,
				Won:          false,
			},
		},
	}
}

func TestNewRuntimeReporterRejectsNilSink(t *testing.T) {
	reporter, err := NewRuntimeReporter(nil)
	if err == nil {
		t.Fatal("expected NewRuntimeReporter to reject nil sink")
	}
	if reporter != nil {
		t.Fatal("expected reporter to be nil on error")
	}
}

func TestRuntimeReporterReportsMultipleCommands(t *testing.T) {
	sink := &fakePlayerDataSink{
		response: encodeRecordMatchResultResponse(t, true, false),
	}
	reporter, err := NewRuntimeReporter(sink)
	if err != nil {
		t.Fatalf("NewRuntimeReporter returned error: %v", err)
	}

	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
				AccountID:    "acct-1",
				Score:        10,
				ShipDeaths:   1,
				Won:          false,
			},
			{
				GamePlayerID:   "Player-2",
				LocalProfileID: "local-2",
				Score:          20,
				ShipDeaths:     2,
				Won:            true,
			},
		},
	}

	if err := reporter.ReportMatchResult(summary); err != nil {
		t.Fatalf("ReportMatchResult returned error: %v", err)
	}
	if len(sink.payloads) != 2 {
		t.Fatalf("expected 2 payloads, got %d", len(sink.payloads))
	}

	firstCommand := decodeRecordMatchResultCommand(t, sink.payloads[0])
	if firstCommand.Type != protocol.PacketTypePlayerDataRecordMatchResult {
		t.Fatalf("expected first command type %q, got %q", protocol.PacketTypePlayerDataRecordMatchResult, firstCommand.Type)
	}
	if firstCommand.ResultID != "room-1-match-1:Player-1" {
		t.Fatalf("expected first ResultID %q, got %q", "room-1-match-1:Player-1", firstCommand.ResultID)
	}

	secondCommand := decodeRecordMatchResultCommand(t, sink.payloads[1])
	if secondCommand.Type != protocol.PacketTypePlayerDataRecordMatchResult {
		t.Fatalf("expected second command type %q, got %q", protocol.PacketTypePlayerDataRecordMatchResult, secondCommand.Type)
	}
	if secondCommand.ResultID != "room-1-match-1:Player-2" {
		t.Fatalf("expected second ResultID %q, got %q", "room-1-match-1:Player-2", secondCommand.ResultID)
	}
}

func TestRuntimeReporterTreatsDuplicateResponseAsSuccess(t *testing.T) {
	sink := &fakePlayerDataSink{
		response: encodeRecordMatchResultResponse(t, true, true),
	}
	reporter, err := NewRuntimeReporter(sink)
	if err != nil {
		t.Fatalf("NewRuntimeReporter returned error: %v", err)
	}

	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
			},
		},
	}

	if err := reporter.ReportMatchResult(summary); err != nil {
		t.Fatalf("ReportMatchResult returned error for duplicate response: %v", err)
	}
	if len(sink.payloads) != 1 {
		t.Fatalf("expected 1 payload, got %d", len(sink.payloads))
	}
}

func TestRuntimeReporterReturnsSinkError(t *testing.T) {
	sink := &fakePlayerDataSink{
		responseErr: errors.New("sink failed"),
	}
	reporter, err := NewRuntimeReporter(sink)
	if err != nil {
		t.Fatalf("NewRuntimeReporter returned error: %v", err)
	}

	if err := reporter.ReportMatchResult(testReporterSummary()); err == nil {
		t.Fatal("expected ReportMatchResult to return sink error")
	}
}

func TestRuntimeReporterReturnsJSONError(t *testing.T) {
	sink := &fakePlayerDataSink{
		response: []byte("{not-json"),
	}
	reporter, err := NewRuntimeReporter(sink)
	if err != nil {
		t.Fatalf("NewRuntimeReporter returned error: %v", err)
	}

	if err := reporter.ReportMatchResult(testReporterSummary()); err == nil {
		t.Fatal("expected ReportMatchResult to return JSON error")
	}
}

func TestRuntimeReporterReturnsRejectedResponseError(t *testing.T) {
	sink := &fakePlayerDataSink{
		response: encodeRecordMatchResultResponse(t, false, false),
	}
	reporter, err := NewRuntimeReporter(sink)
	if err != nil {
		t.Fatalf("NewRuntimeReporter returned error: %v", err)
	}

	if err := reporter.ReportMatchResult(testReporterSummary()); err == nil {
		t.Fatal("expected ReportMatchResult to return rejected response error")
	}
}
