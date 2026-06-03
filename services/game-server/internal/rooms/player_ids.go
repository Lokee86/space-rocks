package rooms

import "github.com/Lokee86/space-rocks/server/internal/playerids"

func formatPlayerID(number int) string {
	return playerids.Format(number)
}

func (room *Room) occupiedPlayerIDsLocked() map[string]bool {
	return room.membership.occupiedPlayerIDs()
}

func (room *Room) nextAvailablePlayerIDLocked() string {
	return room.membership.nextAvailablePlayerID()
}
