package networking

import "github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"

type clientPacketEnvelope struct {
	Type string `json:"type"`
}

func decodeClientPacketEnvelope(msg []byte) (clientPacketEnvelope, error) {
	var envelope clientPacketEnvelope
	err := packetcodec.Decode(msg, &envelope)
	return envelope, err
}
