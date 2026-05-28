package rooms

import "fmt"

func formatPlayerID(number int) string {
	return fmt.Sprintf("Player-%d", number)
}

func (room *Room) occupiedPlayerIDsLocked() map[string]bool {
	occupied := make(map[string]bool, len(room.Members))
	for _, member := range room.Members {
		if member.PlayerID == "" {
			continue
		}
		occupied[member.PlayerID] = true
	}
	return occupied
}

func (room *Room) nextAvailablePlayerIDLocked() string {
	occupied := room.occupiedPlayerIDsLocked()
	for number := 1; ; number++ {
		playerID := formatPlayerID(number)
		if !occupied[playerID] {
			return playerID
		}
	}
}
