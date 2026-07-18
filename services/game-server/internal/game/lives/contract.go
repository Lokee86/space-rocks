package lives

const (
	DefaultSpawnProfileID         = "basic_safe_spawn_v1"
	DefaultShortCooldownThreshold = 10.0
)

type LifeModel string

const (
	LifeModelFinitePerPlayer LifeModel = "finite_per_player"
	LifeModelSharedTeamPool  LifeModel = "shared_team_pool"
	LifeModelInfinite        LifeModel = "infinite"
)

type RespawnTrigger string

const (
	RespawnTriggerAutomatic RespawnTrigger = "automatic"
	RespawnTriggerManual    RespawnTrigger = "manual"
	RespawnTriggerTeam      RespawnTrigger = "team_triggered"
	RespawnTriggerObjective RespawnTrigger = "objective_triggered"
)

type RestorationMode string

const (
	RestorationNone RestorationMode = "none"
	RestorationFull RestorationMode = "full"
)

type TemporaryEffectsPolicy string

const (
	TemporaryEffectsRemove  TemporaryEffectsPolicy = "remove"
	TemporaryEffectsPersist TemporaryEffectsPolicy = "persist"
)

type LoadoutPersistence string

const (
	LoadoutPersist LoadoutPersistence = "persist"
	LoadoutReset   LoadoutPersistence = "reset"
)

type RestorationPolicy struct {
	Health                 RestorationMode
	Shields                RestorationMode
	ShortCooldownThreshold float64
	TemporaryEffects       TemporaryEffectsPolicy
	Loadout                LoadoutPersistence
}

func NewBaselineRestorationPolicy() RestorationPolicy {
	return RestorationPolicy{
		Health:                 RestorationFull,
		Shields:                RestorationFull,
		ShortCooldownThreshold: DefaultShortCooldownThreshold,
		TemporaryEffects:       TemporaryEffectsRemove,
		Loadout:                LoadoutPersist,
	}
}
