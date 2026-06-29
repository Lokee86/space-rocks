package game

// PendingPresentationEvent is the game-owned wrapper for queued presentation
// events delivered through the event batch projection and snapshot paths.
type PendingPresentationEvent struct {
	EventID string
	Event   EventState
}

// PendingPresentationEvents returns a non-draining copy of the queued
// presentation events for the requested player.
func (game *Game) PendingPresentationEvents(playerID string) []PendingPresentationEvent {
	game.mu.Lock()
	defer game.mu.Unlock()

	pending := game.pendingPresentationEvents[playerID]
	pendingEvents := make([]PendingPresentationEvent, len(pending))
	copy(pendingEvents, pending)
	return pendingEvents
}

// DrainPendingPresentationEvents removes only the requested queued presentation
// events for the given player and returns them in original queue order.
func (game *Game) DrainPendingPresentationEvents(playerID string, eventIDs ...string) []PendingPresentationEvent {
	game.mu.Lock()
	defer game.mu.Unlock()

	if len(eventIDs) == 0 {
		return nil
	}

	pending := game.pendingPresentationEvents[playerID]
	if len(pending) == 0 {
		return nil
	}

	lookup := make(map[string]struct{}, len(eventIDs))
	for _, eventID := range eventIDs {
		lookup[eventID] = struct{}{}
	}

	drained := make([]PendingPresentationEvent, 0, len(eventIDs))
	kept := pending[:0]
	for _, event := range pending {
		if _, ok := lookup[event.EventID]; ok {
			drained = append(drained, event)
			continue
		}
		kept = append(kept, event)
	}

	if len(drained) == 0 {
		return nil
	}

	if len(kept) == 0 {
		delete(game.pendingPresentationEvents, playerID)
		return drained
	}

	game.pendingPresentationEvents[playerID] = kept
	return drained
}
