package playerbuild

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

type WeightClass string

const (
	WeightLight    WeightClass = "light"
	WeightStandard WeightClass = "standard"
	WeightHeavy    WeightClass = "heavy"
)

type WeaponPoint string

const (
	Primary1   WeaponPoint = "primary_1"
	Primary2   WeaponPoint = "primary_2"
	Secondary1 WeaponPoint = "secondary_1"
	Secondary2 WeaponPoint = "secondary_2"
)

type PointKind string

const (
	PointNone      PointKind = "none"
	PointHardpoint PointKind = "hardpoint"
	PointSoftpoint PointKind = "softpoint"
)

type ModuleSlot string

const (
	ShieldMod  ModuleSlot = "shield_mod"
	ArmorMod   ModuleSlot = "armor_mod"
	EngineMod  ModuleSlot = "engine_mod"
	UtilityMod ModuleSlot = "utility_mod"
)

type WeaponSize string

const (
	WeaponLight    WeaponSize = "light"
	WeaponStandard WeaponSize = "standard"
	WeaponHeavy    WeaponSize = "heavy"
)

type DeliveryClass string
type TargetingPolicy string
type EffectFlag string
type ModuleActivation string

const (
	ModulePassive ModuleActivation = "passive"
	ModuleActive  ModuleActivation = "active"
)

type HardwiredPolicy string

const (
	HardwiredAllowed    HardwiredPolicy = "allowed"
	HardwiredDisabled   HardwiredPolicy = "disabled"
	HardwiredNormalized HardwiredPolicy = "normalized"
)

type ModuleActivationPolicy string

const (
	ModulesAny         ModuleActivationPolicy = "any"
	ModulesPassiveOnly ModuleActivationPolicy = "passive_only"
)

type ShipVariant struct {
	ID                     string
	WeightClass            WeightClass
	Stats                  runtime.ShipStats
	WeaponPoints           map[WeaponPoint]PointKind
	ModuleSlots            []ModuleSlot
	DefaultPrimaryWeaponID string
}

type WeaponProfile struct {
	ID              string
	RuntimeID       weapons.ID
	Slot            weapons.Slot
	Size            WeaponSize
	DeliveryClass   DeliveryClass
	TargetingPolicy TargetingPolicy
	EffectFlags     []EffectFlag
	AmmoPolicy      weapons.AmmoPolicy
	StartingAmmo    int
}

type ShipStatAdjustment struct {
	MaxHealthDelta     int
	MaxShieldsDelta    int
	RotationMultiplier float64
	ThrustMultiplier   float64
	MaxSpeedMultiplier float64
	DampingMultiplier  float64
}

type ModuleProfile struct {
	ID         string
	Slot       ModuleSlot
	Class      string
	Activation ModuleActivation
	Adjustment ShipStatAdjustment
	BehaviorID string
}

type Catalog struct {
	Ships   map[string]ShipVariant
	Weapons map[string]WeaponProfile
	Modules map[string]ModuleProfile
}

type Rules struct {
	ModeID                    string
	AllowedShipIDs            []string
	BannedShipIDs             []string
	AllowedWeightClasses      []WeightClass
	AllowedWeaponIDs          []string
	BannedWeaponIDs           []string
	AllowedWeaponSlots        []weapons.Slot
	AllowedWeaponSizes        []WeaponSize
	AllowedDeliveryClasses    []DeliveryClass
	AllowedTargetingPolicies  []TargetingPolicy
	RequiredWeaponEffectFlags []EffectFlag
	AllowedModuleSlots        []ModuleSlot
	AllowedModuleClasses      []string
	BannedModuleIDs           []string
	ModuleActivationPolicy    ModuleActivationPolicy
	HardwiredPolicy           HardwiredPolicy
}
