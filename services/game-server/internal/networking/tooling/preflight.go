package tooling

import (
	"fmt"
	"strings"
)

func (router *Router) preflight(context Context, sender Sender, packet map[string]any) (string, PacketPolicy, bool) {
	packetType, ok := packet["type"].(string)
	if !ok || packetType == "" {
		router.sendError(sender, "", "", "invalid_packet", "tooling packet type is required")
		return "", PacketPolicy{}, false
	}

	policy, ok := PacketPolicyFor(packetType)
	if !ok {
		router.sendError(sender, stringValue(packet, "request_id"), stringValue(packet, "run_id"), "unknown_packet", fmt.Sprintf("unknown tooling packet type %q", packetType))
		return "", PacketPolicy{}, false
	}
	if policy.Direction != DirectionClientToServer {
		router.sendError(sender, stringValue(packet, "request_id"), stringValue(packet, "run_id"), "server_packet_not_allowed", "server-to-client tooling packets cannot be submitted")
		return "", PacketPolicy{}, false
	}
	if requiresRequestID(policy) && strings.TrimSpace(stringValue(packet, "request_id")) == "" {
		router.sendError(sender, "", stringValue(packet, "run_id"), "request_id_required", "request_id is required")
		return "", PacketPolicy{}, false
	}
	if policy.Attachment == AttachmentRoom && strings.TrimSpace(context.RoomID) == "" {
		router.sendError(sender, stringValue(packet, "request_id"), stringValue(packet, "run_id"), "room_required", "room attachment is required")
		return "", PacketPolicy{}, false
	}
	if policy.Capability != CapabilityNone && !context.Capabilities.Has(policy.Capability) {
		router.sendError(sender, stringValue(packet, "request_id"), stringValue(packet, "run_id"), "capability_required", "required tooling capability is not granted")
		return "", PacketPolicy{}, false
	}
	return packetType, policy, true
}

func requiresRequestID(policy PacketPolicy) bool {
	switch policy.Interaction {
	case InteractionRequest, InteractionCommand, InteractionSubscription:
		return true
	default:
		return false
	}
}
