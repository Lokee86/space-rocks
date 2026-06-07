extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const ProjectileSyncState = preload("res://scripts/world/projectile_sync_state.gd")
const BULLET_SCENE := preload("res://scenes/bullet.tscn")
const TORPEDO_SCENE := preload("res://scenes/projectiles/torpedo.tscn")


static func scene_for_state(state: Dictionary) -> PackedScene:
	var projectile_type := ProjectileSyncState.projectile_type(state)
	match projectile_type:
		"torpedo":
			return TORPEDO_SCENE
		"bullet":
			return BULLET_SCENE
		_:
			return BULLET_SCENE
