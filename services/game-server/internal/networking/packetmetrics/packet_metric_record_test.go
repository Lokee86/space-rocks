package packetmetrics

import "testing"

func TestPacketMetricRecordCloneReturnsCopy(t *testing.T) {
	record := PacketMetricRecord{
		PacketFamily:         "realtime",
		Lane:                 "overlay",
		Bytes:                128,
		SendReason:           "delta",
		ChunkDecision:        "chunked",
		ResyncDecision:       "not_required",
		Channel:              "sr.overlay",
		EncodedBytes:         128,
		WorldHotCount:        4,
		AsteroidHotCount:     2,
		BulletHotCount:       2,
		AsteroidOffloadedCount: 1,
		BulletOffloadedCount: 1,
		AsteroidMode:         "overflow",
		BulletMode:           "inline",
		Cadence:              "30hz",
		PacketOverTarget:     true,
		PacketOverHardCap:    false,
	}

	clone := record.Clone()

	if clone != record {
		t.Fatalf("expected clone to match record, got %#v want %#v", clone, record)
	}
}

func TestPacketMetricRecordCapturesLaneAndPacketFamily(t *testing.T) {
	record := PacketMetricRecord{PacketFamily: "world_delta", Lane: "world"}
	if record.PacketFamily != "world_delta" || record.Lane != "world" {
		t.Fatalf("record = %#v, want lane and packet family preserved", record)
	}
}


func TestPacketMetricRecordClonePreservesHotLaneDiagnostics(t *testing.T) {
	record := PacketMetricRecord{PacketFamily: "bullet_delta", Lane: "bullets", Channel: "sr.bullets", EncodedBytes: 900, WorldHotCount: 9, AsteroidHotCount: 1, BulletHotCount: 8, AsteroidOffloadedCount: 0, BulletOffloadedCount: 3, AsteroidMode: "overflow", BulletMode: "full_owned_30hz", Cadence: "30hz", PacketOverTarget: true}

	clone := record.Clone()
	if clone != record {
		t.Fatalf("expected clone to preserve hot lane diagnostics, got %#v want %#v", clone, record)
	}
}
