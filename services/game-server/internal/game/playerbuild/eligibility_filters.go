package playerbuild

import (
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

const (
	ReasonUnavailableState    = "unavailable_state"
	ReasonUnknownCatalog      = "unknown_catalog"
	ReasonShipBanned          = "ship_banned"
	ReasonShipNotAllowed      = "ship_not_allowed"
	ReasonWeightNotAllowed    = "weight_class_not_allowed"
	ReasonWeaponBanned        = "weapon_banned"
	ReasonWeaponNotAllowed    = "weapon_not_allowed"
	ReasonWeaponSlotBlocked   = "weapon_slot_not_allowed"
	ReasonWeaponSizeBlocked   = "weapon_size_not_allowed"
	ReasonDeliveryBlocked     = "delivery_class_not_allowed"
	ReasonTargetingBlocked    = "targeting_policy_not_allowed"
	ReasonEffectFlagMissing   = "required_effect_flag_missing"
	ReasonWeaponPointMissing  = "weapon_point_unavailable"
	ReasonModuleBanned        = "module_banned"
	ReasonModuleSlotBlocked   = "module_slot_not_allowed"
	ReasonModuleClassBlocked  = "module_class_not_allowed"
	ReasonModuleSlotMissing   = "module_slot_unavailable"
	ReasonActiveModuleBlocked = "active_module_not_allowed"
)

func shipBlockReason(owned protocol.OwnedShip, variant ShipVariant, found bool, rules Rules) string {
	if owned.State != "normal" {
		return ReasonUnavailableState
	}
	if !found {
		return ReasonUnknownCatalog
	}
	if contains(rules.BannedShipIDs, owned.ShipID) {
		return ReasonShipBanned
	}
	if len(rules.AllowedShipIDs) > 0 && !contains(rules.AllowedShipIDs, owned.ShipID) {
		return ReasonShipNotAllowed
	}
	if len(rules.AllowedWeightClasses) > 0 && !containsComparable(rules.AllowedWeightClasses, variant.WeightClass) {
		return ReasonWeightNotAllowed
	}
	return ""
}

func weaponBlockReason(owned protocol.OwnedWeapon, profile WeaponProfile, found bool, rules Rules) string {
	if owned.State != "normal" {
		return ReasonUnavailableState
	}
	if !found {
		return ReasonUnknownCatalog
	}
	if contains(rules.BannedWeaponIDs, owned.WeaponID) {
		return ReasonWeaponBanned
	}
	if len(rules.AllowedWeaponIDs) > 0 && !contains(rules.AllowedWeaponIDs, owned.WeaponID) {
		return ReasonWeaponNotAllowed
	}
	if len(rules.AllowedWeaponSlots) > 0 && !containsComparable(rules.AllowedWeaponSlots, profile.Slot) {
		return ReasonWeaponSlotBlocked
	}
	if len(rules.AllowedWeaponSizes) > 0 && !containsComparable(rules.AllowedWeaponSizes, profile.Size) {
		return ReasonWeaponSizeBlocked
	}
	if len(rules.AllowedDeliveryClasses) > 0 && !containsComparable(rules.AllowedDeliveryClasses, profile.DeliveryClass) {
		return ReasonDeliveryBlocked
	}
	if len(rules.AllowedTargetingPolicies) > 0 && !containsComparable(rules.AllowedTargetingPolicies, profile.TargetingPolicy) {
		return ReasonTargetingBlocked
	}
	for _, required := range rules.RequiredWeaponEffectFlags {
		if !containsComparable(profile.EffectFlags, required) {
			return ReasonEffectFlagMissing
		}
	}
	return ""
}

func moduleBlockReason(owned protocol.OwnedModule, profile ModuleProfile, found bool, rules Rules) string {
	if owned.State != "normal" {
		return ReasonUnavailableState
	}
	if !found {
		return ReasonUnknownCatalog
	}
	if contains(rules.BannedModuleIDs, owned.ModuleID) {
		return ReasonModuleBanned
	}
	if len(rules.AllowedModuleSlots) > 0 && !containsComparable(rules.AllowedModuleSlots, profile.Slot) {
		return ReasonModuleSlotBlocked
	}
	if len(rules.AllowedModuleClasses) > 0 && !contains(rules.AllowedModuleClasses, profile.Class) {
		return ReasonModuleClassBlocked
	}
	if rules.ModuleActivationPolicy == ModulesPassiveOnly && profile.Activation == ModuleActive {
		return ReasonActiveModuleBlocked
	}
	return ""
}

func blocked(kind, ownedID, catalogID, reason string) BlockedOption {
	return BlockedOption{Kind: kind, OwnedInstanceID: ownedID, CatalogID: catalogID, ReasonCode: reason}
}

func pointsForSlot(slot weapons.Slot) []WeaponPoint {
	if slot == weapons.Secondary {
		return []WeaponPoint{Secondary1, Secondary2}
	}
	return []WeaponPoint{Primary1, Primary2}
}

func supportedBuildLocations(ships []EligibleShipOption, catalog Catalog) (map[WeaponPoint]bool, map[ModuleSlot]bool) {
	points := map[WeaponPoint]bool{}
	slots := map[ModuleSlot]bool{}
	for _, option := range ships {
		variant, ok := catalog.Ships[option.ShipID]
		if !ok {
			continue
		}
		for point, kind := range variant.WeaponPoints {
			if kind == PointHardpoint {
				points[point] = true
			}
		}
		for _, slot := range variant.ModuleSlots {
			slots[slot] = true
		}
	}
	return points, slots
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func containsComparable[T comparable](values []T, value T) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
