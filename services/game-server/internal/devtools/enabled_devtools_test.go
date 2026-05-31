//go:build devtools

package devtools

import "testing"

func TestEnabled_DevtoolsBuild(t *testing.T) {
	if !Enabled() {
		t.Fatalf("Enabled() = false, want true in devtools build")
	}
}

func TestShouldHandleCommand_DevtoolsBuild(t *testing.T) {
	if !ShouldHandleCommand(PacketTypeToggleDebugInvincible) {
		t.Fatalf("ShouldHandleCommand(devtools packet) = false, want true in devtools build")
	}
}
