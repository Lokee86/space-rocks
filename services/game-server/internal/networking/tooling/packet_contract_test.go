package tooling_test

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	networkingtooling "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
)

func TestPacketContractCoversCurrentToolingProtocol(t *testing.T) {
	packetTypes := []string{
		protocol.PacketTypeTelemetrySubscribe,
		protocol.PacketTypeTelemetryUnsubscribe,
		protocol.PacketTypeTelemetryPing,
		protocol.PacketTypeMeasurementStart,
		protocol.PacketTypeMeasurementStop,
		protocol.PacketTypeMeasurementReset,
		protocol.PacketTypeMeasurementSnapshotRequest,
		protocol.PacketTypeTelemetrySnapshot,
		protocol.PacketTypeTelemetryPong,
		protocol.PacketTypeMeasurementStarted,
		protocol.PacketTypeMeasurementSnapshot,
		protocol.PacketTypeMeasurementStopped,
		protocol.PacketTypeToolingError,
	}
	assertCovered(t, packetTypes)
}

func TestPacketContractCoversLegacyDevtoolsMigrationSurface(t *testing.T) {
	packetTypes := []string{
		devtools.PacketTypeToggleDebugInvincible,
		devtools.PacketTypeToggleDebugInfiniteLives,
		devtools.PacketTypeToggleDebugFreezeWorld,
		devtools.PacketTypeToggleDebugFreezePlayer,
		devtools.PacketTypeDebugKillPlayer,
		devtools.PacketTypeDebugSpawnEntity,
		devtools.PacketTypeDebugSpawnPickup,
		devtools.PacketTypeDebugBeginContinuousBulletStream,
		devtools.PacketTypeDebugRespawnPlayer,
		devtools.PacketTypeDebugSetScore,
		devtools.PacketTypeDebugAddScore,
		devtools.PacketTypeDebugSetLives,
		devtools.PacketTypeDebugAddLives,
		devtools.PacketTypeDebugClearBullets,
		devtools.PacketTypeDebugClearAsteroids,
		devtools.PacketTypeDebugStatus,
		devtools.PacketTypeDebugShapeCatalog,
	}
	assertCovered(t, packetTypes)
}

func TestDebugCommandsRequireControlButNotParticipation(t *testing.T) {
	for packetType, policy := range networkingtooling.PacketPolicies() {
		if policy.Direction != networkingtooling.DirectionClientToServer || policy.Interaction != networkingtooling.InteractionCommand {
			continue
		}
		if policy.Capability == networkingtooling.CapabilityNone {
			continue
		}
		if policy.Capability != networkingtooling.CapabilityToolingControl {
			t.Fatalf("%s capability = %q, want %q", packetType, policy.Capability, networkingtooling.CapabilityToolingControl)
		}
		if policy.Attachment != networkingtooling.AttachmentRoom {
			t.Fatalf("%s attachment = %q, want room", packetType, policy.Attachment)
		}
		if policy.ParticipantRequired {
			t.Fatalf("%s unexpectedly requires gameplay participation", packetType)
		}
	}
}

func TestPublicMeasurementAndTelemetryRemainUngated(t *testing.T) {
	packetTypes := []string{
		protocol.PacketTypeTelemetrySubscribe,
		protocol.PacketTypeTelemetryUnsubscribe,
		protocol.PacketTypeTelemetryPing,
		protocol.PacketTypeMeasurementStart,
		protocol.PacketTypeMeasurementStop,
		protocol.PacketTypeMeasurementReset,
		protocol.PacketTypeMeasurementSnapshotRequest,
	}
	for _, packetType := range packetTypes {
		policy, ok := networkingtooling.PacketPolicyFor(packetType)
		if !ok {
			t.Fatalf("missing packet policy for %s", packetType)
		}
		if policy.Capability != networkingtooling.CapabilityNone {
			t.Fatalf("%s capability = %q, want no privileged capability", packetType, policy.Capability)
		}
	}
}

func assertCovered(t *testing.T, packetTypes []string) {
	t.Helper()
	for _, packetType := range packetTypes {
		if _, ok := networkingtooling.PacketPolicyFor(packetType); !ok {
			t.Errorf("missing packet policy for %s", packetType)
		}
	}
}
