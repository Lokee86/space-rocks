package rooms

import "github.com/Lokee86/space-rocks/services/game-server/internal/game"

var (
	startGameCall        = func(instance *game.Game) { instance.Start() }
	stopGameCall         = func(instance *game.Game) { instance.Stop() }
	matchDecisionCall    = func(instance *game.Game) bool { return instance.MatchDecision().IsOver }
	playerMatchFactsCall = func(instance *game.Game) []game.PlayerMatchFact { return instance.PlayerMatchFacts() }
	finalMatchStateCall  = func(instance *game.Game) (game.FinalMatchState, bool) { return instance.LockFinalMatchState() }
)

func (room *Room) StartGameForMember(playerID string, newGame func() *game.Game) *RoomDomainError {
	room.mu.Lock()
	reservedGame, roomErr := room.reserveStartLocked(playerID)
	room.mu.Unlock()
	if roomErr != nil {
		return roomErr
	}
	return room.finishStart(reservedGame, newGame)
}

func (room *Room) StartGameForSession(sessionID string, newGame func() *game.Game) *RoomDomainError {
	room.mu.Lock()
	member, ok := room.memberForSessionLocked(sessionID)
	if !ok {
		room.mu.Unlock()
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}
	reservedGame, roomErr := room.reserveStartLocked(member.PlayerID)
	room.mu.Unlock()
	if roomErr != nil {
		return roomErr
	}
	return room.finishStart(reservedGame, newGame)
}

func (room *Room) reserveStartLocked(playerID string) (*game.Game, *RoomDomainError) {
	if roomErr := room.validateStartLocked(playerID); roomErr != nil {
		return nil, roomErr
	}
	if roomErr := room.lockTeamAssignmentsLocked(); roomErr != nil {
		return nil, roomErr
	}
	if roomErr := room.markStartingLocked(); roomErr != nil {
		room.unlockTeamAssignmentsLocked()
		return nil, roomErr
	}
	return room.match.Game(), nil
}

func (room *Room) finishStart(reservedGame *game.Game, newGame func() *game.Game) *RoomDomainError {
	created := reservedGame == nil
	startedGame := reservedGame
	if startedGame == nil {
		startedGame = newGame()
	}
	if startedGame == nil {
		room.mu.Lock()
		if room.State == RoomStateStarting && room.match.Game() == reservedGame {
			room.State = RoomStateLobby
			room.unlockTeamAssignmentsLocked()
			room.roomMode.clearMatchResolution()
		}
		room.mu.Unlock()
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Could not create game."}
	}
	room.mu.Lock()
	resolvedRules, resolveErr := room.roomMode.resolve(room.roomTeams.rules)
	room.mu.Unlock()
	if resolveErr != nil || startedGame.ConfigureMatchRules(resolvedRules) != nil {
		room.mu.Lock()
		if room.State == RoomStateStarting && room.match.Game() == reservedGame {
			room.State = RoomStateLobby
			room.unlockTeamAssignmentsLocked()
			room.roomMode.clearMatchResolution()
		}
		room.mu.Unlock()
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Could not resolve match rules."}
	}
	startGameCall(startedGame)

	room.mu.Lock()
	valid := room.State == RoomStateStarting && room.match.Game() == reservedGame
	if valid {
		if created {
			room.match.SetGame(startedGame)
		}
		room.State = RoomStateInGame
		room.match.BeginNextMatch(room.ID)
	}
	room.mu.Unlock()
	if !valid {
		room.mu.Lock()
		ownedActive := room.State == RoomStateInGame && room.match.Game() == startedGame
		if room.State == RoomStateStarting && room.match.Game() == reservedGame {
			room.unlockTeamAssignmentsLocked()
			room.roomMode.clearMatchResolution()
		}
		room.mu.Unlock()
		if !ownedActive {
			stopGameCall(startedGame)
		}
	}
	if !valid {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room start was superseded."}
	}
	return nil
}

func (room *Room) MarkStarting() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.markStartingLocked()
}

func (room *Room) MarkInGame() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.markInGameLocked()
}

func (room *Room) markStartingLocked() *RoomDomainError {
	if room.State != RoomStateLobby {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room can only start from the lobby."}
	}
	room.State = RoomStateStarting
	return nil
}

func (room *Room) markInGameLocked() *RoomDomainError {
	if room.State != RoomStateStarting {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room can only enter in-game from starting."}
	}
	room.State = RoomStateInGame
	return nil
}

func (room *Room) MarkGameOver() *RoomDomainError {
	room.mu.Lock()
	if room.State != RoomStateInGame {
		room.mu.Unlock()
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room can only move to game over from in-game."}
	}
	if room.match.Game() == nil {
		room.State = RoomStateGameOver
		room.mu.Unlock()
		return nil
	}
	capture, err := room.captureGameOverLocked()
	room.mu.Unlock()
	if err != nil {
		return err
	}
	facts := playerMatchFactsCall(capture.Game)
	finalState, locked := finalMatchStateCall(capture.Game)
	if !locked {
		finalState = game.FinalMatchState{Players: facts}
	}
	input := buildEndOfMatchInput(capture, finalState, "manual_game_over")
	room.mu.Lock()
	defer room.mu.Unlock()
	if !room.gameOverCaptureMatchesLocked(capture) {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room game state changed."}
	}
	if _, _, _, flowErr := room.match.ResolveEndOfMatch(input); flowErr != nil {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: flowErr.Error()}
	}
	room.State = RoomStateGameOver
	return nil
}

func (room *Room) MarkGameOverIfComplete() bool {
	if room == nil {
		return false
	}
	room.mu.Lock()
	capture, err := room.captureGameOverLocked()
	room.mu.Unlock()
	if err != nil {
		return false
	}
	if !matchDecisionCall(capture.Game) {
		return false
	}
	facts := playerMatchFactsCall(capture.Game)
	finalState, locked := finalMatchStateCall(capture.Game)
	if !locked {
		finalState = game.FinalMatchState{Players: facts}
	}
	input := buildEndOfMatchInput(capture, finalState, "simulation_complete")
	room.mu.Lock()
	defer room.mu.Unlock()
	if !room.gameOverCaptureMatchesLocked(capture) || room.State != RoomStateInGame {
		return false
	}
	if _, _, _, flowErr := room.match.ResolveEndOfMatch(input); flowErr != nil {
		return false
	}
	room.State = RoomStateGameOver
	return true
}

type gameOverCapture struct {
	State                 RoomState
	Game                  *game.Game
	MatchID               string
	TraceID               string
	Joinable              bool
	Members               map[string]RoomMember
	ParticipantIdentities map[string]matchParticipantIdentity
}

func (room *Room) captureGameOverLocked() (gameOverCapture, *RoomDomainError) {
	if room.State != RoomStateInGame || room.match.Game() == nil {
		return gameOverCapture{}, &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room can only move to game over from in-game."}
	}
	members := make(map[string]RoomMember, len(room.membership.members))
	identities := make(map[string]matchParticipantIdentity, len(room.membership.members)+len(room.match.participantIdentities))
	for playerID, identity := range room.match.participantIdentities {
		identities[playerID] = identity
	}
	for playerID, member := range room.membership.members {
		members[playerID] = *member
		identities[playerID] = matchParticipantIdentity{
			AccountID:      member.AccountID,
			LocalProfileID: member.LocalProfileID,
			IsBot:          member.IsBot,
		}
	}

	return gameOverCapture{
		State:                 room.State,
		Game:                  room.match.Game(),
		MatchID:               room.match.CurrentMatchID(),
		TraceID:               room.match.CurrentTraceID(),
		Joinable:              room.Joinable,
		Members:               members,
		ParticipantIdentities: identities,
	}, nil
}

func (room *Room) gameOverCaptureMatchesLocked(capture gameOverCapture) bool {
	return room.State == capture.State && room.match.Game() == capture.Game && room.match.CurrentMatchID() == capture.MatchID
}

func (room *Room) ResetToLobby(playerID string) *RoomDomainError {
	room.mu.Lock()
	oldGame, roomErr := room.resetToLobbyLocked(playerID)
	room.mu.Unlock()
	if roomErr != nil {
		return roomErr
	}
	if oldGame != nil {
		stopGameCall(oldGame)
	}
	return nil
}

func (room *Room) ResetToLobbyForSession(sessionID string) *RoomDomainError {
	room.mu.Lock()
	member, ok := room.memberForSessionLocked(sessionID)
	if !ok {
		room.mu.Unlock()
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}
	oldGame, roomErr := room.resetToLobbyLocked(member.PlayerID)
	room.mu.Unlock()
	if roomErr != nil {
		return roomErr
	}
	if oldGame != nil {
		stopGameCall(oldGame)
	}
	return nil
}

func (room *Room) resetToLobbyLocked(playerID string) (*game.Game, *RoomDomainError) {
	if _, ok := room.membership.memberByPlayerID(playerID); !ok {
		return nil, &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}
	if room.State != RoomStateGameOver {
		return nil, &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room can only return to lobby from game over."}
	}
	oldGame := room.match.Game()
	room.membership.setAllReady(false)
	room.membership.restoreLobbyPlayerIDs()
	room.match.ClearGame()
	room.match.SetActivePlayers(0)
	room.unlockTeamAssignmentsLocked()
	room.roomMode.clearMatchResolution()
	room.State = RoomStateLobby
	return oldGame, nil
}

func (room *Room) IsGameOver() bool {
	if room == nil {
		return false
	}
	room.mu.Lock()
	if room.State != RoomStateInGame || room.match.Game() == nil {
		room.mu.Unlock()
		return false
	}
	instance := room.match.Game()
	room.mu.Unlock()
	return matchDecisionCall(instance)
}

func (room *Room) StartSinglePlayerGame(newGame func() *game.Game) *RoomDomainError {
	room.mu.Lock()
	if roomErr := room.validateStartPreconditionsLocked(); roomErr != nil {
		room.mu.Unlock()
		return roomErr
	}
	if roomErr := room.lockTeamAssignmentsLocked(); roomErr != nil {
		room.mu.Unlock()
		return roomErr
	}
	if roomErr := room.markStartingLocked(); roomErr != nil {
		room.unlockTeamAssignmentsLocked()
		room.mu.Unlock()
		return roomErr
	}
	reservedGame := room.match.Game()
	room.mu.Unlock()
	return room.finishStart(reservedGame, newGame)
}
