package codec

import (
	"encoding/json"
	"errors"
)

type Envelope struct {
	Type string `json:"type"`
}

func DecodeType(data []byte) (string, error) {
	var envelope Envelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return "", err
	}
	if envelope.Type == "" {
		return "", errors.New("missing packet type")
	}
	return envelope.Type, nil
}
