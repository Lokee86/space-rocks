package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestNewRoomDefaultsToUnconfiguredFFA(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	if got := room.TeamConfig(); got != (teams.Config{Structure: teams.StructureFFA}) {
		t.Fatalf("unexpected default team config: %+v", got)
	}
	if room.TeamAssignmentsLocked() {
		t.Fatal("new room should not have locked assignments")
	}
}

func customRoom(t *testing.T, mode teams.AssignmentMode) *Room {
	t.Helper()
	room, err := NewRoomWithConfig("room", RoomStateLobby, nil, RoomCreationConfig{
		TeamConfig: teams.Config{Structure: teams.StructureCustom, AssignmentMode: mode},
		MaxPlayers: MaxPlayersPerRoom,
	})
	if err != nil {
		t.Fatalf("create custom room: %v", err)
	}
	return room
}

func TestPlayerSelectedAssignmentOnlyMovesRequesterBeforeReady(t *testing.T) {
	room := customRoom(t, teams.AssignmentPlayerSelected)
	requester := room.AddMember(NewRoomMember("session-1"))
	other := room.AddMember(NewRoomMember("session-2"))

	if err := room.SetTeamAssignment(requester.SessionID, other.PlayerID, teams.Team1); err == nil {
		t.Fatal("expected player-selected assignment to reject moving another player")
	}
	if err := room.SetTeamAssignment(requester.SessionID, requester.PlayerID, teams.Team1); err != nil {
		t.Fatalf("assign requester: %v", err)
	}
	requester.SetReady(true)
	if err := room.SetTeamAssignment(requester.SessionID, requester.PlayerID, teams.Team2); err == nil || err.Code != RoomErrorNotReady {
		t.Fatalf("expected ready rejection, got %v", err)
	}
}

func TestOwnerAssignmentClearsTargetReady(t *testing.T) {
	room := customRoom(t, teams.AssignmentOwnerAssigned)
	owner := room.AddMember(NewRoomMember("session-owner"))
	target := room.AddMember(NewRoomMember("session-target"))
	target.SetReady(true)

	if err := room.SetTeamAssignment(owner.SessionID, target.PlayerID, teams.Team2); err != nil {
		t.Fatalf("owner assignment: %v", err)
	}
	if room.TeamAssignmentsSnapshot()[target.MemberID] != teams.Team2 || target.Ready {
		t.Fatalf("expected target assignment and ready reset, got %+v", target)
	}
}

func TestOwnerAssignmentKeepsBotReady(t *testing.T) {
	room := customRoom(t, teams.AssignmentOwnerAssigned)
	owner := room.AddMember(NewRoomMember("session-owner"))
	bot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add bot: %v", roomErr)
	}

	if err := room.SetTeamAssignment(owner.SessionID, bot.PlayerID, teams.Team2); err != nil {
		t.Fatalf("assign bot: %v", err)
	}
	botAfter, ok := room.memberForSessionLocked(bot.SessionID)
	if !ok || !botAfter.Ready {
		t.Fatalf("assigned bot should remain ready, got %+v", botAfter)
	}
}

func TestAutoBalancedMembershipChangesUseCurrentRoster(t *testing.T) {
	room, err := NewRoomWithConfig("room", RoomStateLobby, nil, RoomCreationConfig{
		TeamConfig: teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 2},
		MaxPlayers: MaxPlayersPerRoom,
	})
	if err != nil {
		t.Fatalf("create auto-balanced room: %v", err)
	}
	first := room.AddMember(NewRoomMember("session-1"))
	second := room.AddMember(NewRoomMember("session-2"))
	assignments := room.TeamAssignmentsSnapshot()
	if len(assignments) != 2 || assignments[first.MemberID] == assignments[second.MemberID] ||
		(assignments[first.MemberID] != teams.Team1 && assignments[first.MemberID] != teams.Team2) ||
		(assignments[second.MemberID] != teams.Team1 && assignments[second.MemberID] != teams.Team2) {
		t.Fatalf("unexpected initial assignments: %+v", assignments)
	}
	room.RemoveMemberForSession(second.SessionID)
	assignments = room.TeamAssignmentsSnapshot()
	if assignments[first.MemberID] != teams.Team1 {
		t.Fatalf("expected remaining member to be assigned from current roster, got %+v", assignments)
	}
}

func TestStartLocksAndCopiesTeamAssignments(t *testing.T) {
	room := customRoom(t, teams.AssignmentOwnerAssigned)
	owner := room.AddMember(NewRoomMember("session-owner"))
	if err := room.SetTeamAssignment(owner.SessionID, owner.PlayerID, teams.Team1); err != nil {
		t.Fatalf("assign owner: %v", err)
	}
	owner.SetReady(true)
	if err := room.StartGameForMember(owner.PlayerID, game.New); err != nil {
		t.Fatalf("start room: %v", err)
	}
	defer room.CurrentGame().Stop()

	if !room.TeamAssignmentsLocked() {
		t.Fatal("expected team assignments to lock at start")
	}
	snapshot, ok := room.TeamStartSnapshot()
	if !ok || snapshot.Assignments[owner.MemberID] != teams.Team1 {
		t.Fatalf("unexpected team start snapshot: %+v, %v", snapshot, ok)
	}
	snapshot.Assignments[owner.MemberID] = teams.Team8
	later, _ := room.TeamStartSnapshot()
	if later.Assignments[owner.MemberID] != teams.Team1 {
		t.Fatal("team start snapshot was not defensively copied")
	}
	if err := room.SetTeamAssignment(owner.SessionID, owner.PlayerID, teams.Team1); err == nil {
		t.Fatal("expected assignment mutation after start to fail")
	}
}

func TestFailedStartRollsBackTeamAssignmentLock(t *testing.T) {
	room := customRoom(t, teams.AssignmentOwnerAssigned)
	owner := room.AddMember(NewRoomMember("session-owner"))
	if err := room.SetTeamAssignment(owner.SessionID, owner.PlayerID, teams.Team1); err != nil {
		t.Fatalf("assign owner: %v", err)
	}
	owner.SetReady(true)
	if err := room.StartGameForMember(owner.PlayerID, func() *game.Game { return nil }); err == nil {
		t.Fatal("expected failed start")
	}
	if room.TeamAssignmentsLocked() || room.CurrentState() != RoomStateLobby {
		t.Fatal("failed start should leave lobby assignments unlocked")
	}
}

func TestReturnToLobbyUnlocksTeamAssignments(t *testing.T) {
	room := customRoom(t, teams.AssignmentOwnerAssigned)
	owner := room.AddMember(NewRoomMember("session-owner"))
	if err := room.SetTeamAssignment(owner.SessionID, owner.PlayerID, teams.Team1); err != nil {
		t.Fatalf("assign owner: %v", err)
	}
	owner.SetReady(true)
	if err := room.StartGameForMember(owner.PlayerID, game.New); err != nil {
		t.Fatalf("start room: %v", err)
	}
	gameInstance := room.CurrentGame()
	t.Cleanup(func() {
		if gameInstance != nil {
			gameInstance.Stop()
		}
	})
	room.State = RoomStateGameOver

	if err := room.ResetToLobby(owner.PlayerID); err != nil {
		t.Fatalf("return room to lobby: %v", err)
	}
	if room.TeamAssignmentsLocked() {
		t.Fatal("returning to lobby should unlock team assignments")
	}
	if err := room.SetTeamAssignment(owner.SessionID, owner.PlayerID, teams.Team2); err != nil {
		t.Fatalf("expected assignment mutation after return to lobby: %v", err)
	}
}

func TestSinglePlayerStartLocksTeamAssignmentsAndPublishesSnapshot(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))
	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("start single-player game: %v", err)
	}
	gameInstance := room.CurrentGame()
	t.Cleanup(func() {
		if gameInstance != nil {
			gameInstance.Stop()
		}
	})

	if !room.TeamAssignmentsLocked() {
		t.Fatal("single-player start should lock team assignments")
	}
	if snapshot, ok := room.TeamStartSnapshot(); !ok || snapshot.Config.Structure != teams.StructureFFA {
		t.Fatalf("expected FFA team start snapshot, got %+v, %v", snapshot, ok)
	}
}

func TestFailedSinglePlayerStartRollsBackTeamAssignmentLock(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))
	if err := room.StartSinglePlayerGame(func() *game.Game { return nil }); err == nil {
		t.Fatal("expected failed single-player start")
	}
	if room.TeamAssignmentsLocked() || room.CurrentState() != RoomStateLobby {
		t.Fatal("failed single-player start should leave lobby assignments unlocked")
	}
}

func TestSupersededStartDoesNotUnlockAnotherOwnedReservation(t *testing.T) {
	room := customRoom(t, teams.AssignmentOwnerAssigned)
	owner := room.AddMember(NewRoomMember("session-owner"))
	if err := room.SetTeamAssignment(owner.SessionID, owner.PlayerID, teams.Team1); err != nil {
		t.Fatalf("assign owner: %v", err)
	}
	otherGame := game.New()
	defer otherGame.Stop()
	room.State = RoomStateStarting
	room.roomTeams.locked = true
	room.match.SetGame(otherGame)

	if err := room.finishStart(nil, func() *game.Game { return nil }); err == nil {
		t.Fatal("expected superseded start failure")
	}
	if !room.TeamAssignmentsLocked() {
		t.Fatal("superseded start must not unlock another reservation")
	}
}

func TestTeamAssignmentsSnapshotIsDefensive(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session-1"))
	room.roomTeams.assignments[member.MemberID] = teams.Team1

	snapshot := room.TeamAssignmentsSnapshot()
	snapshot[member.MemberID] = teams.Team2
	snapshot["unknown-member"] = teams.Team3

	later := room.TeamAssignmentsSnapshot()
	if later[member.MemberID] != teams.Team1 || later["unknown-member"] != teams.NoTeam {
		t.Fatalf("assignment mutation leaked into room: %+v", later)
	}
}
