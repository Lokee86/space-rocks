package rooms

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

type RoomTeamStartSnapshot struct {
	Config      teams.Config
	Assignments teams.Assignments
}

type roomTeams struct {
	rules       teams.Config
	assignments teams.Assignments
	locked      bool
}

func newRoomTeams(config teams.Config) roomTeams {
	return roomTeams{rules: config, assignments: make(teams.Assignments)}
}

func defaultTeamConfig() teams.Config {
	return teams.Config{Structure: teams.StructureFFA}
}

func defaultAssignmentForStructure(structure teams.Structure) teams.ID {
	if structure == teams.StructureCoOp {
		return teams.Team1
	}
	return teams.NoTeam
}

func TeamCountForConfig(config teams.Config) int {
	switch config.Structure {
	case teams.StructureCoOp:
		return 1
	case teams.StructureCustom:
		return len(teams.OrderedIDs())
	case teams.StructureAutoBalanced:
		return config.AutoTeamCount
	default:
		return 0
	}
}

func (room *Room) TeamConfig() teams.Config {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.roomTeams.rules
}

func (room *Room) TeamAssignmentsSnapshot() teams.Assignments {
	room.mu.Lock()
	defer room.mu.Unlock()
	return copyTeamAssignments(room.roomTeams.assignments)
}

func (room *Room) TeamAssignmentsLocked() bool {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.roomTeams.locked
}

func copyTeamAssignments(assignments teams.Assignments) teams.Assignments {
	copy := make(teams.Assignments, len(assignments))
	for participantID, teamID := range assignments {
		copy[participantID] = teamID
	}
	return copy
}

func (room *Room) SetTeamAssignment(requestingSessionID string, targetPlayerID string, teamID teams.ID) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateLobby || room.roomTeams.locked {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Team assignments can only change in the lobby."}
	}
	if room.roomTeams.rules.Structure != teams.StructureCustom {
		return &RoomDomainError{Code: RoomErrorInvalidTeamAssignment, Message: "Manual team assignment is not enabled for this room."}
	}
	if !teams.IsValidTeamID(teamID) {
		return &RoomDomainError{Code: RoomErrorInvalidTeamAssignment, Message: fmt.Sprintf("Team ID %q is invalid.", teamID)}
	}
	requester, ok := room.memberForSessionLocked(requestingSessionID)
	if !ok {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}
	target, ok := room.membership.memberByPlayerID(targetPlayerID)
	if !ok {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Target member is not in the room."}
	}
	switch room.roomTeams.rules.AssignmentMode {
	case teams.AssignmentPlayerSelected:
		if requester != target {
			return &RoomDomainError{Code: RoomErrorInvalidTeamAssignment, Message: "Players may only assign themselves."}
		}
		if requester.Ready {
			return &RoomDomainError{Code: RoomErrorNotReady, Message: "Ready players cannot change team assignment."}
		}
	case teams.AssignmentOwnerAssigned:
		if requester.PlayerID != room.membership.ownerIDValue() {
			return &RoomDomainError{Code: RoomErrorNotRoomOwner, Message: "Only the room owner can assign teams."}
		}
	default:
		return &RoomDomainError{Code: RoomErrorInvalidTeamAssignment, Message: "Room assignment mode is invalid."}
	}

	if room.roomTeams.assignments[target.MemberID] != teamID {
		room.roomTeams.assignments[target.MemberID] = teamID
		if room.roomTeams.rules.AssignmentMode == teams.AssignmentOwnerAssigned {
			target.SetReady(false)
		}
	}
	return nil
}

func (room *Room) rebalanceAutoBalancedLocked() *RoomDomainError {
	if room.State != RoomStateLobby || room.roomTeams.rules.Structure != teams.StructureAutoBalanced {
		return nil
	}
	participantIDs := make([]string, 0, len(room.membership.members))
	for _, member := range room.membership.members {
		participantIDs = append(participantIDs, member.MemberID)
	}
	assignments, err := teams.ResolveAutoBalanced(participantIDs, room.roomTeams.rules.AutoTeamCount)
	if err != nil {
		return &RoomDomainError{Code: RoomErrorInvalidTeamAssignment, Message: err.Error()}
	}
	for _, member := range room.membership.members {
		teamID := assignments[member.MemberID]
		if room.roomTeams.assignments[member.MemberID] != teamID {
			member.SetReady(false)
		}
		room.roomTeams.assignments[member.MemberID] = teamID
	}
	return nil
}

func (room *Room) lockTeamAssignmentsLocked() *RoomDomainError {
	requested := room.roomTeams.assignments
	if room.roomTeams.rules.Structure != teams.StructureCustom {
		requested = nil
	}
	participantIDs := make([]string, 0, len(room.membership.members))
	for _, member := range room.membership.members {
		participantIDs = append(participantIDs, member.MemberID)
	}
	assignments, err := teams.ResolveAssignments(room.roomTeams.rules, participantIDs, requested)
	if err != nil {
		return &RoomDomainError{Code: RoomErrorInvalidTeamAssignment, Message: err.Error()}
	}
	for memberID, teamID := range assignments {
		room.roomTeams.assignments[memberID] = teamID
	}
	room.roomTeams.locked = true
	return nil
}

func (room *Room) unlockTeamAssignmentsLocked() {
	room.roomTeams.locked = false
}

func (room *Room) TeamStartSnapshot() (RoomTeamStartSnapshot, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()
	if !room.roomTeams.locked {
		return RoomTeamStartSnapshot{}, false
	}
	return RoomTeamStartSnapshot{Config: room.roomTeams.rules, Assignments: copyTeamAssignments(room.roomTeams.assignments)}, true
}
