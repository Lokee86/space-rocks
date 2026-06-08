extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const PickupSyncState := preload("res://scripts/world/pickup_sync_state.gd")


func test_age_seconds_reads_packet_field() -> void:
	var state := {
		Packets.FIELD_AGE_SECONDS: 3.5,
	}

	assert_eq(PickupSyncState.age_seconds(state), 3.5)


func test_lifespan_seconds_reads_packet_field() -> void:
	var state := {
		Packets.FIELD_LIFESPAN_SECONDS: 12.0,
	}

	assert_eq(PickupSyncState.lifespan_seconds(state), 12.0)


func test_pickup_class_reads_packet_field() -> void:
	var state := {
		Packets.FIELD_PICKUP_CLASS: "powerup",
	}

	assert_eq(PickupSyncState.pickup_class(state), "powerup")


func test_missing_fields_default_to_empty_or_zero() -> void:
	var state := {}

	assert_eq(PickupSyncState.pickup_class(state), "")
	assert_eq(PickupSyncState.age_seconds(state), 0.0)
	assert_eq(PickupSyncState.lifespan_seconds(state), 0.0)
