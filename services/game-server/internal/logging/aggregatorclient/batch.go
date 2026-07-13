package aggregatorclient

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type encodedBatch struct {
	BatchID uuid.UUID         `json:"batch_id"`
	Events  []json.RawMessage `json:"events"`
}

func encodeBatch(events [][]byte) ([]byte, error) {
	rawEvents := make([]json.RawMessage, len(events))
	for index, event := range events {
		if !json.Valid(event) {
			return nil, fmt.Errorf("event %d is invalid JSON", index)
		}
		rawEvents[index] = json.RawMessage(event)
	}
	return json.Marshal(encodedBatch{BatchID: uuid.New(), Events: rawEvents})
}
