package player

import "testing"

func TestTargetStatusForWorldState(t *testing.T) {
	tests := []struct {
		name  string
		state WorldState
		exists bool
		want  TargetStatus
	}{
		{
			name:   "missing when not exists",
			state:  WorldState{Targetable: true},
			exists: false,
			want:   TargetStatusMissing,
		},
		{
			name:   "active when exists and targetable",
			state:  WorldState{Targetable: true},
			exists: true,
			want:   TargetStatusActive,
		},
		{
			name:   "inactive when exists and not targetable",
			state:  WorldState{Targetable: false},
			exists: true,
			want:   TargetStatusInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TargetStatusForWorldState(tt.state, tt.exists)
			if got != tt.want {
				t.Fatalf("TargetStatusForWorldState() = %q, want %q", got, tt.want)
			}
		})
	}
}
