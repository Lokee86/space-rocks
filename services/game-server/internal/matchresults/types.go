package matchresults

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"

type SessionContext string

const (
	SessionSinglePlayer SessionContext = "single_player"
	SessionMultiplayer  SessionContext = "multiplayer"
)

type TerminalStatus string

const (
	TerminalCompleted                  TerminalStatus = "completed"
	TerminalFailed                     TerminalStatus = "failed"
	TerminalCancelled                  TerminalStatus = "cancelled"
	TerminalInvalid                    TerminalStatus = "invalid"
	TerminalAdministrativelyTerminated TerminalStatus = "administratively_terminated"
)

type Outcome string

const (
	OutcomeWon       Outcome = "won"
	OutcomeLost      Outcome = "lost"
	OutcomeDraw      Outcome = "draw"
	OutcomeCompleted Outcome = "completed"
	OutcomeFailed    Outcome = "failed"
	OutcomeAborted   Outcome = "aborted"
)

type ParticipationDisposition string

const (
	DispositionParticipated ParticipationDisposition = "participated"
	DispositionDeparted     ParticipationDisposition = "departed"
	DispositionForfeited    ParticipationDisposition = "forfeited"
	DispositionLateJoin     ParticipationDisposition = "late_join"
)

type PlayerRef struct {
	GamePlayerID   string
	AccountID      string
	LocalProfileID string
	IsBot          bool
}

type ParticipantFact struct {
	PlayerRef   PlayerRef
	TeamID      teams.ID
	Score       int
	ShipDeaths  int
	Disposition ParticipationDisposition
}

type ParticipantResult struct {
	PlayerRef      PlayerRef
	TeamID         teams.ID
	Outcome        Outcome
	Placement      int
	Disposition    ParticipationDisposition
	FinalScore     int
	ShipDeaths     int
	CompletionTime float64
	TargetValue    float64
}

type TeamResult struct {
	TeamID     teams.ID
	Outcome    Outcome
	Placement  int
	FinalScore int
}

type ObjectiveResolution struct {
	DefinitionID  string
	InstanceID    string
	Scope         string
	OwnerID       string
	Status        string
	Progress      float64
	Target        float64
	FailureReason string
	Discovered    bool
}

type MissionResolution struct {
	MissionID string
	Status    string
}

type ChallengeResolutionAggregate struct {
	ChallengeID      string
	AggregationScope string
	Resolution       string
	PlayerRef        PlayerRef
	TeamID           teams.ID
	ObjectiveRef     string
	MissionRef       string
	Values           map[string]float64
}

type BuildInput struct {
	MatchID       string
	TraceID       string
	ModeID        string
	Session       SessionContext
	TeamStructure teams.Structure
	EndReason     string
	Participants  []ParticipantFact
	Objectives    []ObjectiveResolution
	Missions      []MissionResolution
	Challenges    []ChallengeResolutionAggregate
}

type MatchDecision struct {
	TerminalStatus    TerminalStatus
	EndReason         string
	Participants      []ParticipantResult
	Teams             []TeamResult
	WinningPlayerRefs []PlayerRef
	WinningTeamRefs   []teams.ID
}

type MatchSummary struct {
	MatchID           string
	TraceID           string
	ModeID            string
	Session           SessionContext
	TerminalStatus    TerminalStatus
	EndReason         string
	Participants      []ParticipantResult
	Teams             []TeamResult
	WinningPlayerRefs []PlayerRef
	WinningTeamRefs   []teams.ID
	Objectives        []ObjectiveResolution
	Missions          []MissionResolution
	Challenges        []ChallengeResolutionAggregate
}
