extends RefCounted

const LOCAL_PLAYER_ID := "local-player"
const REMOTE_PLAYER_ID := "remote-player"
const ASTEROID_ID := "asteroid-1"
const BULLET_ID := "bullet-1"


static func snapshot() -> Dictionary:
	return {
		"self_id": LOCAL_PLAYER_ID,
		"players": players(),
		"asteroids": asteroids(),
		"bullets": bullets(),
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
		"x": x,
		"y": y,
		"rotation": rotation,
		"score": score,
	}


static func asteroid_state(x: float, y: float, variant: int, scale: float) -> Dictionary:
	return {
		"x": x,
		"y": y,
		"variant": variant,
		"scale": scale,
	}


static func bullet_state(x: float, y: float, rotation: float) -> Dictionary:
	return {
		"x": x,
		"y": y,
		"rotation": rotation,
		"owner_id": LOCAL_PLAYER_ID,
	}

