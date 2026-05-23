package rooms

import (
	"crypto/rand"
	"strings"
)

func NormalizeRoomID(roomID string) string {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return DefaultRoomID
	}

	return roomID
}

func NormalizeRoomCode(roomCode string) string {
	return strings.ToUpper(strings.TrimSpace(roomCode))
}

func IsValidRoomCode(roomCode string) bool {
	if len(roomCode) != RoomCodeLength {
		return false
	}
	for _, character := range roomCode {
		if !strings.ContainsRune(RoomCodeAlphabet, character) {
			return false
		}
	}

	return true
}

func GenerateRoomCode() (string, error) {
	bytes := make([]byte, RoomCodeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.Grow(RoomCodeLength)
	for _, value := range bytes {
		builder.WriteByte(RoomCodeAlphabet[int(value)%len(RoomCodeAlphabet)])
	}

	return builder.String(), nil
}
