//go:build nodevtools

package devtools

import "testing"

func TestEnabled_NodevtoolsBuild(t *testing.T) {
	if Enabled() {
		t.Fatalf("Enabled() = true, want false in nodevtools build")
	}
}

func TestShouldHandleCommand_NodevtoolsBuild(t *testing.T) {
	if got := ShouldHandleCommand(PacketTypeToggleDebugInvincible); got {
		t.Fatalf("ShouldHandleCommand(devtools packet) = true, want false in nodevtools build")
	}

	if got := ShouldHandleCommand("input"); got {
		t.Fatalf("ShouldHandleCommand(non-devtools packet) = true, want false")
	}
}
