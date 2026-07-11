package inbound

import "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"

type ClientPacketEnvelope struct {
	Type string `json:"type"`
}

func DecodeClientPacketEnvelope(msg []byte) (ClientPacketEnvelope, error) {
	var envelope ClientPacketEnvelope
	err := packetcodec.Decode(msg, &envelope)
	return envelope, err
}
