package devtools

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/server/internal/playerids"
)

func formatDebugGamePlayerID(number int) string {
	return fmt.Sprintf("player-%d", number)
}

func parseDebugGamePlayerIDNumber(playerID string) (int, bool) {
	trimmed := strings.TrimSpace(playerID)
	parts := strings.Split(trimmed, "-")
	if len(parts) != 2 {
		return 0, false
	}

	if parts[0] != "player" && parts[0] != "Player" {
		return 0, false
	}

	number, err := strconv.Atoi(parts[1])
	if err != nil || number <= 0 {
		return 0, false
	}

	return number, true
}

func normalizeDebugSpawnPlayerID(playerID string) (string, bool) {
	number, ok := parseDebugGamePlayerIDNumber(playerID)
	if !ok {
		return "", false
	}

	return formatDebugGamePlayerID(number), true
}

func allocateDebugGameplayPlayerID(
	isOccupied func(string) bool,
	reserve func(string) bool,
) (string, bool) {
	for number := 1; number <= playerids.MaxPlayers; number++ {
		candidate := formatDebugGamePlayerID(number)
		if isOccupied(candidate) {
			continue
		}
		if !reserve(candidate) {
			continue
		}
		return candidate, true
	}

	return "", false
}

func NormalizeDebugSpawnPlayerID(playerID string) (string, bool) {
	return normalizeDebugSpawnPlayerID(playerID)
}

func AllocateDebugGameplayPlayerID(
	isOccupied func(string) bool,
	reserve func(string) bool,
) (string, bool) {
	return allocateDebugGameplayPlayerID(isOccupied, reserve)
}
