package realtime

const (
	LaneWorld              = "world"
	LaneOverlay            = "overlay"
	LaneSession            = "session"
	LaneEvent              = "event"
	LaneControl            = "control"
	LaneAsteroids          = "asteroids"
	LaneBullets            = "bullets"
	LaneAsteroidsLifecycle = "asteroids.lifecycle"
	LaneBulletsLifecycle   = "bullets.lifecycle"
)

func IsBaselineLane(lane Lane) bool {
	return lane == LaneWorld || lane == LaneOverlay || lane == LaneSession
}

const (
	PacketFamilyWorldFull          = "world_full"
	PacketFamilyWorldDelta         = "world_delta"
	PacketFamilyOverlayFull        = "overlay_full"
	PacketFamilyOverlayDelta       = "overlay_delta"
	PacketFamilySessionFull        = "session_full"
	PacketFamilySessionDelta       = "session_delta"
	PacketFamilyAsteroidDelta      = "asteroid_delta"
	PacketFamilyBulletDelta        = "bullet_delta"
	PacketFamilyAsteroidsLifecycle = "asteroids_lifecycle"
	PacketFamilyBulletsLifecycle   = "bullets_lifecycle"
	PacketFamilyEventBatch         = "event_batch"
	PacketFamilyResyncRequest      = "resync_request"
	PacketFamilyResyncRequired     = "resync_required"
)
