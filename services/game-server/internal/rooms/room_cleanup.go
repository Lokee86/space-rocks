package rooms

import "time"

func (room *Room) StopCleanupTimer() {
	room.mu.Lock()
	defer room.mu.Unlock()

	if timer := room.cleanup.Timer(); timer != nil {
		timer.Stop()
	}
	room.cleanup.ClearTimer()
}

func (room *Room) HasCleanupTimer() bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.cleanup.Timer() != nil
}

func (room *Room) StopGameIfPresent() {
	room.mu.Lock()
	game := room.match.Game()
	room.mu.Unlock()

	if game != nil {
		game.Stop()
	}
}

func (room *Room) CleanupVersionMatches(cleanupVersion int) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.cleanup.VersionMatches(cleanupVersion)
}

func (room *Room) CurrentCleanupVersion() int {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.cleanup.Version()
}

func (room *Room) ScheduleCleanupTimer(cleanupDelay time.Duration, cleanup func(cleanupVersion int)) int {
	room.mu.Lock()
	defer room.mu.Unlock()

	cleanupVersion := room.cleanup.IncrementVersion()
	if timer := room.cleanup.Timer(); timer != nil {
		timer.Stop()
	}
	room.cleanup.SetTimer(time.AfterFunc(cleanupDelay, func() {
		cleanup(cleanupVersion)
	}))
	return cleanupVersion
}

func (room *Room) ShouldCleanup() bool {
	return room != nil && room.IsEmpty()
}
