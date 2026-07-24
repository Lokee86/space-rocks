package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func TestOwnerCanAddReadyBot(t *testing.T) {
	room := NewRoom("ROOM01", RoomStateLobby, nil)
	owner := room.AddMemberSessionID("owner-session")

	bot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add bot: %v", roomErr)
	}
	if !bot.IsBot || !bot.Ready || !bot.Connected {
		t.Fatalf("unexpected bot member: %+v", bot)
	}
	if bot.SessionID == "" || bot.PlayerID == "" {
		t.Fatalf("bot must have stable room identity: %+v", bot)
	}
}

func TestNonOwnerCannotAddBot(t *testing.T) {
	room := NewRoom("ROOM01", RoomStateLobby, nil)
	room.AddMemberSessionID("owner-session")
	guest := room.AddMemberSessionID("guest-session")

	_, roomErr := room.AddBotForOwnerSession(guest.SessionID)
	if roomErr == nil || roomErr.Code != RoomErrorNotRoomOwner {
		t.Fatalf("expected owner error, got %+v", roomErr)
	}
}

func TestOwnerCanRemoveBotAndGuestButNotOwner(t *testing.T) {
	room := NewRoom("ROOM01", RoomStateLobby, nil)
	owner := room.AddMemberSessionID("owner-session")
	guest := room.AddMemberSessionID("guest-session")
	bot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add bot: %v", roomErr)
	}

	removedBot, roomErr := room.RemoveMemberForOwnerSession(owner.SessionID, bot.PlayerID)
	if roomErr != nil || !removedBot.IsBot {
		t.Fatalf("remove bot: removed=%+v err=%v", removedBot, roomErr)
	}
	removedGuest, roomErr := room.RemoveMemberForOwnerSession(owner.SessionID, guest.PlayerID)
	if roomErr != nil || removedGuest.SessionID != guest.SessionID {
		t.Fatalf("remove guest: removed=%+v err=%v", removedGuest, roomErr)
	}
	_, roomErr = room.RemoveMemberForOwnerSession(owner.SessionID, owner.PlayerID)
	if roomErr == nil || roomErr.Code != RoomErrorCannotRemoveOwner {
		t.Fatalf("expected cannot-remove-owner error, got %+v", roomErr)
	}
}

func TestOwnerTransferSkipsBots(t *testing.T) {
	room := NewRoom("ROOM01", RoomStateLobby, nil)
	owner := room.AddMemberSessionID("owner-session")
	guest := room.AddMemberSessionID("guest-session")
	if _, roomErr := room.AddBotForOwnerSession(owner.SessionID); roomErr != nil {
		t.Fatalf("add bot: %v", roomErr)
	}

	if _, _, ok := room.RemoveMemberForSession(owner.SessionID); !ok {
		t.Fatal("expected owner removal")
	}
	if room.OwnerID() != guest.PlayerID {
		t.Fatalf("expected human guest to become owner, got %q", room.OwnerID())
	}
}

func TestLobbyResetKeepsBotsReady(t *testing.T) {
	room := NewRoom("ROOM01", RoomStateLobby, nil)
	owner := room.AddMemberSessionID("owner-session")
	bot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add bot: %v", roomErr)
	}

	room.mu.Lock()
	room.membership.setAllReady(false)
	room.mu.Unlock()

	room.mu.Lock()
	ownerAfterPointer, ownerOK := room.memberForSessionLocked(owner.SessionID)
	botAfterPointer, botOK := room.memberForSessionLocked(bot.SessionID)
	ownerAfter := RoomMember{}
	botAfter := RoomMember{}
	if ownerOK {
		ownerAfter = *ownerAfterPointer
	}
	if botOK {
		botAfter = *botAfterPointer
	}
	room.mu.Unlock()
	if !ownerOK || ownerAfter.Ready {
		t.Fatalf("expected human owner to reset unready, got %+v", ownerAfter)
	}
	if !botOK || !botAfter.Ready {
		t.Fatalf("expected bot to remain ready, got %+v", botAfter)
	}
}

func TestLastHumanLeaveRetiresBotMembershipAndGameplayPlayers(t *testing.T) {
	manager := NewRoomManager()
	defer manager.StopAll()
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	owner := room.AddMemberSessionID("owner-session")
	bot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add bot: %v", roomErr)
	}
	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}
	gameInstance := room.GameInstance()
	ownerPlayerID := gameInstance.AddPlayer()
	botPlayerID := gameInstance.AddBot()
	context := room.GameplayContext()
	if !room.ActivateMemberPlayer(context, owner.SessionID, ownerPlayerID) {
		t.Fatal("activate owner")
	}
	if !room.ActivateMemberPlayer(context, bot.SessionID, botPlayerID) {
		t.Fatal("activate bot")
	}

	result, leaveErr := manager.LeaveMember(room.ID, owner.SessionID, ownerPlayerID)
	if leaveErr != nil {
		t.Fatalf("leave: %v", leaveErr)
	}
	if result.RemainingMembers != 0 || result.ActivePlayers != 0 {
		t.Fatalf("expected abandoned room to be empty, got %+v", result)
	}
	facts := gameInstance.PlayerMatchFacts()
	if len(facts) != 0 {
		t.Fatalf("expected bot and owner gameplay players removed, got %+v", facts)
	}
	if room.OwnerID() != "" {
		t.Fatalf("expected no bot owner, got %q", room.OwnerID())
	}
}
