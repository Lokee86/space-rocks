extends RefCounted

var self_id = null
var lives = null
var score = null
var health = null
var max_health = null
var shields = null
var max_shields = null
var shield_module_id = null
var armor_module_id = null
var engine_module_id = null
var utility_module_id = null
var respawn_cooldown = null
var primary_weapon_id = null
var secondary_weapon_id = null
var primary_ammo_policy = null
var secondary_ammo_policy = null
var primary_cooldown_remaining = null
var secondary_cooldown_remaining = null
var primary_ammo_remaining = null
var secondary_ammo_remaining = null

func clear_overlay() -> void:
	self_id = null
	lives = null
	score = null
	health = null
	max_health = null
	shields = null
	max_shields = null
	shield_module_id = null
	armor_module_id = null
	engine_module_id = null
	utility_module_id = null
	respawn_cooldown = null
	primary_weapon_id = null
	secondary_weapon_id = null
	primary_ammo_policy = null
	secondary_ammo_policy = null
	primary_cooldown_remaining = null
	secondary_cooldown_remaining = null
	primary_ammo_remaining = null
	secondary_ammo_remaining = null

func apply_full_overlay(overlay_packet: Dictionary) -> void:
	clear_overlay()
	_apply_overlay_fields(overlay_packet)

func apply_overlay_delta(overlay_packet: Dictionary) -> void:
	_apply_overlay_fields(overlay_packet)
	for record in _array_field(overlay_packet, "receiver_creates"):
		if record is Dictionary:
			_apply_overlay_fields(record)
	for record in _array_field(overlay_packet, "receiver_updates"):
		if record is Dictionary:
			_apply_overlay_fields(record)

func _apply_overlay_fields(overlay_packet: Dictionary) -> void:
	if overlay_packet.has("self_id"):
		self_id = overlay_packet.get("self_id")
	if overlay_packet.has("lives"):
		lives = overlay_packet.get("lives")
	if overlay_packet.has("score"):
		score = overlay_packet.get("score")
	if overlay_packet.has("health"):
		health = overlay_packet.get("health")
	if overlay_packet.has("max_health"):
		max_health = overlay_packet.get("max_health")
	if overlay_packet.has("shields"):
		shields = overlay_packet.get("shields")
	if overlay_packet.has("max_shields"):
		max_shields = overlay_packet.get("max_shields")
	if overlay_packet.has("shield_module_id"):
		shield_module_id = overlay_packet.get("shield_module_id")
	if overlay_packet.has("armor_module_id"):
		armor_module_id = overlay_packet.get("armor_module_id")
	if overlay_packet.has("engine_module_id"):
		engine_module_id = overlay_packet.get("engine_module_id")
	if overlay_packet.has("utility_module_id"):
		utility_module_id = overlay_packet.get("utility_module_id")
	if overlay_packet.has("respawn_cooldown"):
		respawn_cooldown = overlay_packet.get("respawn_cooldown")
	if overlay_packet.has("primary_weapon_id"):
		primary_weapon_id = overlay_packet.get("primary_weapon_id")
	if overlay_packet.has("secondary_weapon_id"):
		secondary_weapon_id = overlay_packet.get("secondary_weapon_id")
	if overlay_packet.has("primary_ammo_policy"):
		primary_ammo_policy = overlay_packet.get("primary_ammo_policy")
	if overlay_packet.has("secondary_ammo_policy"):
		secondary_ammo_policy = overlay_packet.get("secondary_ammo_policy")
	if overlay_packet.has("primary_cooldown_remaining"):
		primary_cooldown_remaining = overlay_packet.get("primary_cooldown_remaining")
	if overlay_packet.has("secondary_cooldown_remaining"):
		secondary_cooldown_remaining = overlay_packet.get("secondary_cooldown_remaining")
	if overlay_packet.has("primary_ammo_remaining"):
		primary_ammo_remaining = overlay_packet.get("primary_ammo_remaining")
	if overlay_packet.has("secondary_ammo_remaining"):
		secondary_ammo_remaining = overlay_packet.get("secondary_ammo_remaining")

func _array_field(packet: Dictionary, key: String) -> Array:
	var value = packet.get(key, [])
	if value is Array:
		return value
	return []

