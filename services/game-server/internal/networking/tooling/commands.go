package tooling

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

type CommandController interface {
	HandleCommand(string, devtools.DebugCommand) bool
}

func (router *Router) handleCommand(context Context, sender Sender, packet map[string]any, packetType string) bool {
	if context.CommandController == nil {
		router.sendError(sender, stringValue(packet, "request_id"), stringValue(packet, "run_id"), "command_controller_unavailable", "devtools command controller is not configured")
		return true
	}

	var command devtools.DebugCommand
	if !router.decode(packet, &command, sender, packetType) {
		return true
	}
	if !context.CommandController.HandleCommand(context.GamePlayerID, command) {
		router.sendError(sender, command.RequestID, stringValue(packet, "run_id"), "command_not_applied", "devtools command was not applied")
		return true
	}

	logging.Emit(observability.Request{
		Event: observability.EventNameDevtoolsCommandApplied,
		Context: observability.Context{
			TraceID:    command.TraceID,
			SessionID:  context.SessionID,
			RoomID:     context.RoomID,
			PlayerID:   context.GamePlayerID,
			PacketType: command.Type,
		},
	})
	router.send(sender, protocol.ToolingCommandResult{
		Type:        protocol.PacketTypeToolingCommandResult,
		RequestID:   command.RequestID,
		CommandType: command.Type,
		Applied:     true,
	})
	return true
}
