extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ProjectileSyncState := preload("res://scripts/world/projectile_sync_state.gd")


func test_server_position_reads_packet_coordinates() -> void:
	assert_eq(
		ProjectileSyncState.server_position({
			Packets.FIELD_X: 420.5,
			Packets.FIELD_Y: 840.25,
		}),
		Vector2(420.5, 840.25)
	)


func test_projectile_type_defaults_to_bullet_when_missing() -> void:
	assert_eq(
		ProjectileSyncState.projectile_type({}),
		"bullet"
	)


func test_projectile_type_defaults_to_bullet_when_empty() -> void:
	assert_eq(
		ProjectileSyncState.projectile_type({
			Packets.FIELD_PROJECTILE_TYPE: "",
		}),
		"bullet"
	)


func test_projectile_type_returns_bullet_value() -> void:
	assert_eq(
		ProjectileSyncState.projectile_type({
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		}),
		"bullet"
	)


func test_projectile_type_returns_torpedo_value() -> void:
	assert_eq(
		ProjectileSyncState.projectile_type({
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		}),
		"torpedo"
	)

