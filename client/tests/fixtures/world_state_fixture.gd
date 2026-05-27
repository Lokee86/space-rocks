extends RefCounted

const Packets := preload("res://scripts/networking/packets/packets.gd")

const LOCAL_PLAYER_ID := "local-player"
const REMOTE_PLAYER_ID := "remote-player"
const ASTEROID_ID := "asteroid-1"
const BULLET_ID := "bullet-1"


static func state() -> Dictionary:
	return {
		Packets.FIELD_TYPE: Packets.TYPE_STATE,
		Packets.FIELD_SELF_ID: LOCAL_PLAYER_ID,
		Packets.FIELD_LIVES: 3,
		Packets.FIELD_PLAYERS: players(),
		Packets.FIELD_ASTEROIDS: asteroids(),
		Packets.FIELD_BULLETS: bullets(),
		Packets.FIELD_EVENTS: [],
	}


static func players() -> Dictionary:
	return {
		LOCAL_PLAYER_ID: player_state(100.0, 120.0, 0.25, 10),
		REMOTE_PLAYER_ID: player_state(220.0, 240.0, 1.5, 20),
	}


static func asteroids() -> Dictionary:
	return {
		ASTEROID_ID: asteroid_state(320.0, 340.0, 1, 1.25),
	}


static func bullets() -> Dictionary:
	return {
		BULLET_ID: bullet_state(420.0, 440.0, 0.75),
	}


static func player_state(x: float, y: float, rotation: float, score: int = 0) -> Dictionary:
	return {
		Packets.FIELD_X: x,
		Packets.FIELD_Y: y,
		Packets.FIELD_ROTATION: rotation,
		Packets.FIELD_SCORE: score,
	}


static func asteroid_state(x: float, y: float, variant: int, scale: float) -> Dictionary:
	return {
		Packets.FIELD_X: x,
		Packets.FIELD_Y: y,
		Packets.FIELD_VARIANT: variant,
		Packets.FIELD_SCALE: scale,
	}


static func bullet_state(x: float, y: float, rotation: float) -> Dictionary:
	return {
		Packets.FIELD_X: x,
		Packets.FIELD_Y: y,
		Packets.FIELD_ROTATION: rotation,
		Packets.FIELD_OWNER_ID: LOCAL_PLAYER_ID,
	}
