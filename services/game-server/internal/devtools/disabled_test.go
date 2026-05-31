//go:build !devtools

package devtools

import "testing"

func TestShouldHandleCommand_DefaultBuild(t *testing.T) {
	if got := ShouldHandleCommand(PacketTypeToggleDebugInvincible); got {
		t.Fatalf("ShouldHandleCommand(devtools packet) = true, want false in default build")
	}

	if got := ShouldHandleCommand("input"); got {
		t.Fatalf("ShouldHandleCommand(non-devtools packet) = true, want false")
	}
}
