package rooms

import (
	"crypto/rand"
	"fmt"
)

func newMemberID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(fmt.Errorf("generate member id: %w", err))
	}

	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		uint32(value[0])<<24|uint32(value[1])<<16|uint32(value[2])<<8|uint32(value[3]),
		uint16(value[4])<<8|uint16(value[5]),
		uint16(value[6])<<8|uint16(value[7]),
		uint16(value[8])<<8|uint16(value[9]),
		uint64(value[10])<<40|
			uint64(value[11])<<32|
			uint64(value[12])<<24|
			uint64(value[13])<<16|
			uint64(value[14])<<8|
			uint64(value[15]),
	)
}
