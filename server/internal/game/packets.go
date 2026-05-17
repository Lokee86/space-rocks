package game

type InputPacket struct {
	Type  string     `json:"type"`
	Input InputState `json:"input"`
}

type InputState struct {
	Forward bool `json:"forward"`
	Back    bool `json:"back"`
	Right   bool `json:"right"`
	Left    bool `json:"left"`
	Shoot   bool `json:"shoot"`
}

type PlayerState struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
}

type StatePacket struct {
	Type    string                 `json:"type"`
	SelfID  string                 `json:"self_id"`
	Players map[string]PlayerState `json:"players"`
}
