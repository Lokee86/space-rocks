package rooms

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

type RoomCreationConfig struct {
	ModeConfig modes.RoomModeConfig
	TeamConfig teams.Config
	MaxPlayers int
}

func DefaultRoomCreationConfig() RoomCreationConfig {
	return RoomCreationConfig{
		ModeConfig: modes.DefaultRoomModeConfig(),
		TeamConfig: defaultTeamConfig(),
		MaxPlayers: MaxPlayersPerRoom,
	}
}

func normalizeRoomCreationConfig(config RoomCreationConfig) RoomCreationConfig {
	defaults := DefaultRoomCreationConfig()
	config.ModeConfig = modes.NormalizeRoomModeConfig(config.ModeConfig)
	if config.TeamConfig.Structure == "" {
		config.TeamConfig = defaults.TeamConfig
	}
	if config.MaxPlayers == 0 {
		config.MaxPlayers = defaults.MaxPlayers
	}
	return config
}

func validateRoomCreationConfig(config RoomCreationConfig) error {
	if err := modes.ValidateRoomModeConfig(config.ModeConfig); err != nil {
		return fmt.Errorf("invalid mode configuration: %w", err)
	}
	if err := teams.ValidateConfig(config.TeamConfig); err != nil {
		return fmt.Errorf("invalid team configuration: %w", err)
	}
	if config.MaxPlayers < 1 || config.MaxPlayers > MaxPlayersPerRoom {
		return fmt.Errorf("room capacity must be between 1 and %d", MaxPlayersPerRoom)
	}
	return nil
}
