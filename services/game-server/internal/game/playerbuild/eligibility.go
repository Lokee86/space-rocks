package playerbuild

import "sort"

func ComputeEligibility(playerID string, inventory Inventory, catalog Catalog, rawRules Rules) EligibleBuildOptions {
	rules := NormalizeRules(rawRules)
	options := EligibleBuildOptions{
		ModeID:         rules.ModeID,
		PlayerID:       playerID,
		WeaponsByPoint: map[WeaponPoint][]EligibleWeaponOption{},
		ModulesBySlot:  map[ModuleSlot][]EligibleModuleOption{},
	}

	for _, owned := range inventory.OwnedShips {
		variant, ok := catalog.Ships[owned.ShipID]
		if reason := shipBlockReason(owned, variant, ok, rules); reason != "" {
			options.BlockedOptions = append(options.BlockedOptions, blocked("ship", owned.OwnedShipID, owned.ShipID, reason))
			continue
		}
		options.EligibleShips = append(options.EligibleShips, EligibleShipOption{
			OwnedShipID:            owned.OwnedShipID,
			ShipID:                 owned.ShipID,
			WeightClass:            variant.WeightClass,
			DefaultPrimaryWeaponID: variant.DefaultPrimaryWeaponID,
		})
	}

	supportedPoints, supportedSlots := supportedBuildLocations(options.EligibleShips, catalog)

	for _, owned := range inventory.OwnedWeapons {
		profile, ok := catalog.Weapons[owned.WeaponID]
		if reason := weaponBlockReason(owned, profile, ok, rules); reason != "" {
			options.BlockedOptions = append(options.BlockedOptions, blocked("weapon", owned.OwnedWeaponID, owned.WeaponID, reason))
			continue
		}
		added := false
		for _, point := range pointsForSlot(profile.Slot) {
			if !supportedPoints[point] {
				continue
			}
			added = true
			options.WeaponsByPoint[point] = append(options.WeaponsByPoint[point], EligibleWeaponOption{
				OwnedWeaponID: owned.OwnedWeaponID,
				WeaponID:      owned.WeaponID,
				WeaponPoint:   point,
			})
		}
		if !added {
			options.BlockedOptions = append(options.BlockedOptions, blocked("weapon", owned.OwnedWeaponID, owned.WeaponID, ReasonWeaponPointMissing))
		}
	}

	for _, owned := range inventory.OwnedModules {
		profile, ok := catalog.Modules[owned.ModuleID]
		if reason := moduleBlockReason(owned, profile, ok, rules); reason != "" {
			options.BlockedOptions = append(options.BlockedOptions, blocked("module", owned.OwnedModuleID, owned.ModuleID, reason))
			continue
		}
		if !supportedSlots[profile.Slot] {
			options.BlockedOptions = append(options.BlockedOptions, blocked("module", owned.OwnedModuleID, owned.ModuleID, ReasonModuleSlotMissing))
			continue
		}
		options.ModulesBySlot[profile.Slot] = append(options.ModulesBySlot[profile.Slot], EligibleModuleOption{
			OwnedModuleID: owned.OwnedModuleID,
			ModuleID:      owned.ModuleID,
			ModuleSlot:    profile.Slot,
		})
	}

	sortEligibility(&options)
	options.FallbackLoadout = FallbackSelection(inventory, options)
	return options
}

func sortEligibility(options *EligibleBuildOptions) {
	sort.Slice(options.EligibleShips, func(i, j int) bool {
		return options.EligibleShips[i].OwnedShipID < options.EligibleShips[j].OwnedShipID
	})
	for point := range options.WeaponsByPoint {
		sort.Slice(options.WeaponsByPoint[point], func(i, j int) bool {
			return options.WeaponsByPoint[point][i].OwnedWeaponID < options.WeaponsByPoint[point][j].OwnedWeaponID
		})
	}
	for slot := range options.ModulesBySlot {
		sort.Slice(options.ModulesBySlot[slot], func(i, j int) bool {
			return options.ModulesBySlot[slot][i].OwnedModuleID < options.ModulesBySlot[slot][j].OwnedModuleID
		})
	}
	sort.Slice(options.BlockedOptions, func(i, j int) bool {
		return options.BlockedOptions[i].OwnedInstanceID < options.BlockedOptions[j].OwnedInstanceID
	})
}
