package playerbuild

type BlockedOption struct {
	Kind            string
	OwnedInstanceID string
	CatalogID       string
	WeaponPoint     WeaponPoint
	ModuleSlot      ModuleSlot
	ReasonCode      string
}

type EligibleShipOption struct {
	OwnedShipID            string
	ShipID                 string
	WeightClass            WeightClass
	DefaultPrimaryWeaponID string
}

type EligibleWeaponOption struct {
	OwnedWeaponID string
	WeaponID      string
	WeaponPoint   WeaponPoint
}

type EligibleModuleOption struct {
	OwnedModuleID string
	ModuleID      string
	ModuleSlot    ModuleSlot
}

type EligibleBuildOptions struct {
	ModeID          string
	PlayerID        string
	EligibleShips   []EligibleShipOption
	WeaponsByPoint  map[WeaponPoint][]EligibleWeaponOption
	ModulesBySlot   map[ModuleSlot][]EligibleModuleOption
	BlockedOptions  []BlockedOption
	FallbackLoadout LoadoutSelection
}

type LoadoutSelection struct {
	PlayerID               string
	ModeID                 string
	SelectedOwnedShipID    string
	SelectedWeaponsByPoint map[WeaponPoint]string
	SelectedModulesBySlot  map[ModuleSlot]string
	StartingAmmoByPoint    map[WeaponPoint]int
}
