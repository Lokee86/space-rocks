package tooling

const (
	CapabilityNone           = ""
	CapabilityToolingRead    = "tooling.read"
	CapabilityToolingControl = "tooling.control"
	CapabilityAdminControl   = "admin.control"

	DirectionClientToServer = "client_to_server"
	DirectionServerToClient = "server_to_client"

	InteractionRequest      = "request"
	InteractionCommand      = "command"
	InteractionSubscription = "subscription"
	InteractionResponse     = "response"
	InteractionServerPush   = "server_push"

	AttachmentConnection = "connection"
	AttachmentRoom       = "room"

	PacketTypeDebugStatusSubscribe     = "debug_status_subscribe"
	PacketTypeDebugStatusUnsubscribe   = "debug_status_unsubscribe"
	PacketTypeDebugShapeCatalogRequest = "debug_shape_catalog_request"
	PacketTypeToolingCommandResult     = "tooling_command_result"
)

type PacketPolicy struct {
	PacketType          string
	Direction           string
	Interaction         string
	Capability          string
	Attachment          string
	ParticipantRequired bool
}

var packetPolicies = map[string]PacketPolicy{
	"telemetry_subscribe":                  policy("telemetry_subscribe", DirectionClientToServer, InteractionSubscription, CapabilityNone, AttachmentRoom),
	"telemetry_unsubscribe":                policy("telemetry_unsubscribe", DirectionClientToServer, InteractionSubscription, CapabilityNone, AttachmentRoom),
	"telemetry_ping":                       policy("telemetry_ping", DirectionClientToServer, InteractionRequest, CapabilityNone, AttachmentConnection),
	"measurement_start":                    policy("measurement_start", DirectionClientToServer, InteractionCommand, CapabilityNone, AttachmentRoom),
	"measurement_stop":                     policy("measurement_stop", DirectionClientToServer, InteractionCommand, CapabilityNone, AttachmentRoom),
	"measurement_reset":                    policy("measurement_reset", DirectionClientToServer, InteractionCommand, CapabilityNone, AttachmentRoom),
	"measurement_snapshot_request":         policy("measurement_snapshot_request", DirectionClientToServer, InteractionRequest, CapabilityNone, AttachmentRoom),
	"telemetry_snapshot":                   policy("telemetry_snapshot", DirectionServerToClient, InteractionServerPush, CapabilityNone, AttachmentRoom),
	"telemetry_pong":                       policy("telemetry_pong", DirectionServerToClient, InteractionResponse, CapabilityNone, AttachmentConnection),
	"measurement_started":                  policy("measurement_started", DirectionServerToClient, InteractionResponse, CapabilityNone, AttachmentRoom),
	"measurement_snapshot":                 policy("measurement_snapshot", DirectionServerToClient, InteractionResponse, CapabilityNone, AttachmentRoom),
	"measurement_stopped":                  policy("measurement_stopped", DirectionServerToClient, InteractionResponse, CapabilityNone, AttachmentRoom),
	"tooling_error":                        policy("tooling_error", DirectionServerToClient, InteractionResponse, CapabilityNone, AttachmentConnection),
	"toggle_debug_invincible":              policy("toggle_debug_invincible", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"toggle_debug_infinite_lives":          policy("toggle_debug_infinite_lives", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"toggle_debug_freeze_world":            policy("toggle_debug_freeze_world", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"toggle_debug_freeze_player":           policy("toggle_debug_freeze_player", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_kill_player":                    policy("debug_kill_player", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_spawn_entity":                   policy("debug_spawn_entity", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_spawn_pickup":                   policy("debug_spawn_pickup", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_begin_continuous_bullet_stream": policy("debug_begin_continuous_bullet_stream", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_respawn_player":                 policy("debug_respawn_player", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_set_score":                      policy("debug_set_score", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_add_score":                      policy("debug_add_score", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_set_lives":                      policy("debug_set_lives", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_add_lives":                      policy("debug_add_lives", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_clear_bullets":                  policy("debug_clear_bullets", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_clear_asteroids":                policy("debug_clear_asteroids", DirectionClientToServer, InteractionCommand, CapabilityToolingControl, AttachmentRoom),
	"debug_status":                         policy("debug_status", DirectionServerToClient, InteractionServerPush, CapabilityToolingRead, AttachmentRoom),
	"debug_shape_catalog":                  policy("debug_shape_catalog", DirectionServerToClient, InteractionResponse, CapabilityToolingRead, AttachmentRoom),
	PacketTypeDebugStatusSubscribe:         policy(PacketTypeDebugStatusSubscribe, DirectionClientToServer, InteractionSubscription, CapabilityToolingRead, AttachmentRoom),
	PacketTypeDebugStatusUnsubscribe:       policy(PacketTypeDebugStatusUnsubscribe, DirectionClientToServer, InteractionSubscription, CapabilityToolingRead, AttachmentRoom),
	PacketTypeDebugShapeCatalogRequest:     policy(PacketTypeDebugShapeCatalogRequest, DirectionClientToServer, InteractionRequest, CapabilityToolingRead, AttachmentRoom),
	PacketTypeToolingCommandResult:         policy(PacketTypeToolingCommandResult, DirectionServerToClient, InteractionResponse, CapabilityToolingControl, AttachmentRoom),
}

func PacketPolicyFor(packetType string) (PacketPolicy, bool) {
	policy, ok := packetPolicies[packetType]
	return policy, ok
}

func PacketPolicies() map[string]PacketPolicy {
	copy := make(map[string]PacketPolicy, len(packetPolicies))
	for packetType, policy := range packetPolicies {
		copy[packetType] = policy
	}
	return copy
}

func policy(packetType string, direction string, interaction string, capability string, attachment string) PacketPolicy {
	return PacketPolicy{
		PacketType:  packetType,
		Direction:   direction,
		Interaction: interaction,
		Capability:  capability,
		Attachment:  attachment,
	}
}
