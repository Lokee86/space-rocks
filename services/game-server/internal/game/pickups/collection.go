package pickups

const EffectTypeAddLives = "add_lives"

type CollectionRequest struct {
	PlayerID   string
	PickupID   string
	PickupType string
	X          float64
	Y          float64
}

type EffectIntent struct {
	PlayerID   string
	PickupID   string
	PickupType string
	EffectType string
	Amount     int
}

type CollectionResult struct {
	Collected   bool
	PlayerID    string
	PickupID    string
	PickupType  string
	X           float64
	Y           float64
	EffectIntent EffectIntent
}

func ResolveCollection(req CollectionRequest) CollectionResult {
	if req.PlayerID == "" || req.PickupID == "" || req.PickupType == "" {
		return CollectionResult{Collected: false}
	}

	if req.PickupType != "1_up" {
		return CollectionResult{Collected: false}
	}

	return CollectionResult{
		Collected:  true,
		PlayerID:   req.PlayerID,
		PickupID:   req.PickupID,
		PickupType: req.PickupType,
		X:          req.X,
		Y:          req.Y,
		EffectIntent: EffectIntent{
			PlayerID:   req.PlayerID,
			PickupID:   req.PickupID,
			PickupType: req.PickupType,
			EffectType: EffectTypeAddLives,
			Amount:     1,
		},
	}
}
