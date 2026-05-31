//go:build !nodevtools

package devtools

import "testing"

func TestEnabled_DefaultBuild(t *testing.T) {
	if !Enabled() {
		t.Fatalf("Enabled() = false, want true in default build")
	}
}

func TestShouldHandleCommand_DefaultBuild(t *testing.T) {
	if !ShouldHandleCommand(PacketTypeToggleDebugInvincible) {
		t.Fatalf("ShouldHandleCommand(devtools packet) = false, want true in default build")
	}
}
