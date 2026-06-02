package rooms

import "github.com/Lokee86/space-rocks/server/internal/rooms/roomrules"

func roomDomainErrorFromDecision(decision roomrules.Decision) *RoomDomainError {
	if decision.Allowed {
		return nil
	}

	return &RoomDomainError{
		Code:    decision.Code,
		Message: decision.Message,
	}
}
