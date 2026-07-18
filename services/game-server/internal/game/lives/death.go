package lives

import playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
import "github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"

type AttributionCategory string

const (
	AttributionPlayerCaused         AttributionCategory = "player_caused"
	AttributionSelfDestruction      AttributionCategory = "self_destruction"
	AttributionEnvironmental        AttributionCategory = "environmental"
	AttributionUnattributed         AttributionCategory = "unattributed"
	DeathAttributionPlayerCaused                        = AttributionPlayerCaused
	DeathAttributionSelfDestruction                     = AttributionSelfDestruction
	DeathAttributionEnvironmental                       = AttributionEnvironmental
	DeathAttributionUnattributed                        = AttributionUnattributed
)

type DeathInput struct {
	PlayerID        string
	DestroyedShipID string
	TeamID          teams.ID
	MatchID         string
	ModeID          string
	CauseCode       string
	Attribution     AttributionCategory
	KillerPlayerID  string
	AssistPlayerIDs []string
}

func (input DeathInput) clone() DeathInput {
	input.AssistPlayerIDs = append([]string(nil), input.AssistPlayerIDs...)
	return input
}

func (input DeathInput) normalized() DeathInput {
	if input.Attribution == "" {
		input.Attribution = AttributionUnattributed
	}
	return input.clone()
}

type DeathFact struct {
	Accepted        bool
	PlayerID        string
	PreviousStatus  playerstate.Status
	ResultingStatus playerstate.Status
	RemainingLives  int
	RespawnDelay    float64
	DeathCount      int
	ReasonCode      string
	Input           DeathInput
}

func (fact DeathFact) clone() DeathFact {
	fact.Input = fact.Input.clone()
	return fact
}
