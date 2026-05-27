package rooms

import "time"

func (room *Room) StopCleanupTimer() {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.CleanupTimer != nil {
		room.CleanupTimer.Stop()
	}
	room.CleanupTimer = nil
}

func (room *Room) StopGameIfPresent() {
	room.mu.Lock()
	game := room.Game
	room.mu.Unlock()

	if game != nil {
		game.Stop()
	}
}

func (room *Room) CleanupVersionMatches(cleanupVersion int) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.CleanupVersion == cleanupVersion
}

func (room *Room) ScheduleCleanupTimer(cleanupDelay time.Duration, cleanup func(cleanupVersion int)) int {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.CleanupVersion++
	cleanupVersion := room.CleanupVersion
	if room.CleanupTimer != nil {
		room.CleanupTimer.Stop()
	}
	room.CleanupTimer = time.AfterFunc(cleanupDelay, func() {
		cleanup(cleanupVersion)
	})
	return cleanupVersion
}
