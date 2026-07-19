package playerbuild

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

type ResolvedWeapon struct {
	OwnedWeaponID string
	CatalogID     string
	RuntimeID     weapons.ID
	Point         WeaponPoint
	AmmoPolicy    weapons.AmmoPolicy
	StartingAmmo  int
}

type ResolvedModule struct {
	OwnedModuleID  string
	CatalogID      string
	Slot           ModuleSlot
	Activation     ModuleActivation
	BehaviorID     string
	Hardwired      bool
	EffectsApplied bool
}

type ShieldPolicy struct {
	MaxShields int
	StartsFull bool
}

type StartingEquipmentState struct {
	PrimaryAmmo   int
	SecondaryAmmo int
}

type ResolvedPlayerBuild struct {
	PlayerID                 string
	ModeID                   string
	InventoryVersion         int
	SelectedOwnedShipID      string
	ShipID                   string
	WeightClass              WeightClass
	ShipStats                runtime.ShipStats
	WeaponPointLayout        map[WeaponPoint]PointKind
	EquippedWeapons          map[WeaponPoint]ResolvedWeapon
	EquippedModules          map[ModuleSlot]ResolvedModule
	HardwiredEquipment       []ResolvedModule
	AppliedPassiveEffects    []string
	ActiveModuleDeclarations []string
	ShieldPolicy             ShieldPolicy
	StartingEquipmentState   StartingEquipmentState
	PlayerArmory             weapons.PlayerArmory
}
