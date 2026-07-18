package rooms

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

type roomMode struct {
	config   modes.RoomModeConfig
	resolved *modes.ResolvedMatchRules
}

func newRoomMode(config modes.RoomModeConfig) roomMode {
	return roomMode{config: modes.NormalizeRoomModeConfig(config)}
}

func (mode *roomMode) resolve(teamConfig teams.Config) (modes.ResolvedMatchRules, error) {
	if mode.resolved != nil {
		return modes.CloneResolvedMatchRules(*mode.resolved), nil
	}
	resolved, err := modes.Resolve(mode.config, teamConfig)
	if err != nil {
		return modes.ResolvedMatchRules{}, err
	}
	mode.resolved = &resolved
	return modes.CloneResolvedMatchRules(resolved), nil
}

func (mode *roomMode) clearMatchResolution() {
	mode.resolved = nil
}

func (room *Room) ModeConfig() modes.RoomModeConfig {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.roomMode.config
}

func (room *Room) ResolvedMatchRules() (modes.ResolvedMatchRules, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()
	if room.roomMode.resolved == nil {
		return modes.ResolvedMatchRules{}, false
	}
	return modes.CloneResolvedMatchRules(*room.roomMode.resolved), true
}
