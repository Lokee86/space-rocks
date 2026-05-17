package game

type Ship struct {
	ID            string
	X             float64
	Y             float64
	Rotation      float64
	Velocity      Vector2
	Input         InputState
	Config        ClientConfig
	ShootCooldown float64
}

type Bullet struct {
	ID       string
	OwnerID  string
	X        float64
	Y        float64
	Rotation float64
	Velocity Vector2
	Life     float64
}

type Asteroid struct {
	ID       string
	X        float64
	Y        float64
	Velocity Vector2
	Size     int
	Variant  int
}

type GameState struct {
	Players     map[string]*Ship
	Projectiles map[string]*Bullet
	Asteroids   map[string]*Asteroid
	Enemies     map[string]*Ship
}

func NewGameState() GameState {
	return GameState{
		Players:     make(map[string]*Ship),
		Projectiles: make(map[string]*Bullet),
		Asteroids:   make(map[string]*Asteroid),
		Enemies:     make(map[string]*Ship),
	}
}
