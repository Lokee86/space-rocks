package game

type ClientPacket struct {
	Type   string       `json:"type"`
	Input  InputState   `json:"input"`
	Config ClientConfig `json:"config"`
}

type ClientConfig struct {
	VisibleWorldWidth  float64 `json:"visible_world_width"`
	VisibleWorldHeight float64 `json:"visible_world_height"`
}

type InputState struct {
	Forward bool `json:"forward"`
	Back    bool `json:"back"`
	Right   bool `json:"right"`
	Left    bool `json:"left"`
	Shoot   bool `json:"shoot"`
}

type ShipState struct {
	ID       string  `json:"id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
}

type AsteroidState struct {
	ID      string  `json:"id"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Size    int     `json:"size"`
	Variant int     `json:"variant"`
}

type BulletState struct {
	ID       string  `json:"id"`
	OwnerID  string  `json:"owner_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
}

type EventState struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type StatePacket struct {
	Type      string                   `json:"type"`
	SelfID    string                   `json:"self_id"`
	Players   map[string]ShipState     `json:"players"`
	Bullets   map[string]BulletState   `json:"bullets"`
	Asteroids map[string]AsteroidState `json:"asteroids"`
	Events    []EventState             `json:"events"`
}
