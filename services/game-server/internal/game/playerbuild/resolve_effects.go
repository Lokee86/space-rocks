package playerbuild

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

func applyWeaponToRuntime(build *ResolvedPlayerBuild, weapon ResolvedWeapon) {
	equipped := weapons.Equipped{ID: weapon.RuntimeID, AmmoPolicy: weapon.AmmoPolicy}
	switch weapon.Point {
	case Primary1:
		build.PlayerArmory.Primary = equipped
		build.StartingEquipmentState.PrimaryAmmo = weapon.StartingAmmo
	case Secondary1:
		build.PlayerArmory.Secondary = equipped
		build.StartingEquipmentState.SecondaryAmmo = weapon.StartingAmmo
	}
}

func applyModule(build *ResolvedPlayerBuild, profile ModuleProfile, hardwired bool) {
	applyAdjustment(&build.ShipStats, profile.Adjustment)
	effectID := profile.ID
	if hardwired {
		effectID = "hardwired:" + effectID
	}
	if profile.Activation == ModulePassive {
		build.AppliedPassiveEffects = append(build.AppliedPassiveEffects, effectID)
	} else if profile.BehaviorID != "" {
		build.ActiveModuleDeclarations = append(build.ActiveModuleDeclarations, profile.BehaviorID)
	}
}

func resolveHardwired(build *ResolvedPlayerBuild, equipment []HardwiredEquipment, catalog Catalog, policy HardwiredPolicy) {
	if policy == HardwiredDisabled {
		return
	}
	for _, owned := range equipment {
		profile, ok := catalog.Modules[owned.EquipmentID]
		if !ok || owned.State != "normal" {
			continue
		}
		resolved := ResolvedModule{
			OwnedModuleID:  owned.HardwiredID,
			CatalogID:      profile.ID,
			Slot:           profile.Slot,
			Activation:     profile.Activation,
			BehaviorID:     profile.BehaviorID,
			Hardwired:      true,
			EffectsApplied: policy == HardwiredAllowed,
		}
		build.HardwiredEquipment = append(build.HardwiredEquipment, resolved)
		if resolved.EffectsApplied {
			applyModule(build, profile, true)
		}
	}
}

func applyAdjustment(stats *runtime.ShipStats, adjustment ShipStatAdjustment) {
	stats.MaxHealth += adjustment.MaxHealthDelta
	stats.MaxShields += adjustment.MaxShieldsDelta
	stats.RotationSpeed *= multiplier(adjustment.RotationMultiplier)
	stats.ThrustForce *= multiplier(adjustment.ThrustMultiplier)
	stats.MaxSpeed *= multiplier(adjustment.MaxSpeedMultiplier)
	stats.Damping *= multiplier(adjustment.DampingMultiplier)
}

func multiplier(value float64) float64 {
	if value == 0 {
		return 1
	}
	return value
}

func clonePointLayout(source map[WeaponPoint]PointKind) map[WeaponPoint]PointKind {
	clone := make(map[WeaponPoint]PointKind, len(source))
	for point, kind := range source {
		clone[point] = kind
	}
	return clone
}

func (build ResolvedPlayerBuild) Clone() ResolvedPlayerBuild {
	clone := build
	clone.WeaponPointLayout = clonePointLayout(build.WeaponPointLayout)
	clone.EquippedWeapons = cloneMap(build.EquippedWeapons)
	clone.EquippedModules = cloneMap(build.EquippedModules)
	clone.HardwiredEquipment = append([]ResolvedModule(nil), build.HardwiredEquipment...)
	clone.AppliedPassiveEffects = append([]string(nil), build.AppliedPassiveEffects...)
	clone.ActiveModuleDeclarations = append([]string(nil), build.ActiveModuleDeclarations...)
	return clone
}

func cloneMap[K comparable, V any](source map[K]V) map[K]V {
	clone := make(map[K]V, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func ValidateResolvedBuild(build ResolvedPlayerBuild) error {
	if build.PlayerID == "" || build.ShipID == "" || build.SelectedOwnedShipID == "" {
		return fmt.Errorf("resolved build identity is incomplete")
	}
	if build.PlayerArmory.Primary.ID == "" {
		return fmt.Errorf("resolved build requires a primary weapon")
	}
	if build.ShipStats.MaxHealth <= 0 || build.ShipStats.MaxShields < 0 {
		return fmt.Errorf("resolved build survivability is invalid")
	}
	return nil
}
