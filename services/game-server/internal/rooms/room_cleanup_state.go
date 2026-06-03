package rooms

import "time"

type roomCleanup struct {
	timer   *time.Timer
	version int
}

func newRoomCleanup() *roomCleanup {
	return &roomCleanup{}
}

func (rc *roomCleanup) Timer() *time.Timer {
	return rc.timer
}

func (rc *roomCleanup) SetTimer(timer *time.Timer) {
	rc.timer = timer
}

func (rc *roomCleanup) ClearTimer() {
	rc.timer = nil
}

func (rc *roomCleanup) Version() int {
	return rc.version
}

func (rc *roomCleanup) IncrementVersion() int {
	rc.version++
	return rc.version
}

func (rc *roomCleanup) VersionMatches(version int) bool {
	return rc.version == version
}
