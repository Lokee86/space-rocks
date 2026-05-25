extends RefCounted

const PLAYER_ACTIVE := "active"


static func is_active(status) -> bool:
	return str(status) == PLAYER_ACTIVE
