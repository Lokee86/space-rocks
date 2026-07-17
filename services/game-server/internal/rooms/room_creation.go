package rooms

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

type RoomCreationConfig struct {
	TeamConfig teams.Config
	MaxPlayers int
}

func DefaultRoomCreationConfig() RoomCreationConfig {
	return RoomCreationConfig{
		TeamConfig: defaultTeamConfig(),
		MaxPlayers: MaxPlayersPerRoom,
	}
}

func normalizeRoomCreationConfig(config RoomCreationConfig) RoomCreationConfig {
	defaults := DefaultRoomCreationConfig()
	if config.TeamConfig.Structure == "" {
		config.TeamConfig = defaults.TeamConfig
	}
	if config.MaxPlayers == 0 {
		config.MaxPlayers = defaults.MaxPlayers
	}
	return config
}

func validateRoomCreationConfig(config RoomCreationConfig) error {
	if err := teams.ValidateConfig(config.TeamConfig); err != nil {
		return fmt.Errorf("invalid team configuration: %w", err)
	}
	if config.MaxPlayers < 1 || config.MaxPlayers > MaxPlayersPerRoom {
		return fmt.Errorf("room capacity must be between 1 and %d", MaxPlayersPerRoom)
	}
	return nil
}
