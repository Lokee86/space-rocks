extends RefCounted

var self_id = null
var lives = null
var score = null
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

func _apply_overlay_fields(overlay_packet: Dictionary) -> void:
	if overlay_packet.has("self_id"):
		self_id = overlay_packet.get("self_id")
	if overlay_packet.has("lives"):
		lives = overlay_packet.get("lives")
	if overlay_packet.has("score"):
		score = overlay_packet.get("score")
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

