package networking

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
)

func (session *webSocketSession) enqueueLoadoutOptions(traceID string) {
	options, selection := session.playerBuildPacketStates()
	payload, err := packetcodec.Encode(game.LoadoutOptionsPacket{
		Type: game.PacketTypeLoadoutOptions, TraceID: traceID,
		BuildOptions: options, LoadoutSelection: selection,
	})
	if err == nil {
		session.enqueue(payload)
	}
}

func (session *webSocketSession) playerBuildPacketStates() (game.EligibleBuildOptionsState, game.LoadoutSelectionState) {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return buildOptionsState(session.buildState.context.Options), loadoutSelectionState(session.buildState)
}

func buildOptionsState(options playerbuild.EligibleBuildOptions) game.EligibleBuildOptionsState {
	state := game.EligibleBuildOptionsState{ModeID: options.ModeID, PlayerID: options.PlayerID}
	for _, option := range options.EligibleShips {
		state.EligibleShips = append(state.EligibleShips, game.BuildShipOption{
			OwnedShipID: option.OwnedShipID, ShipID: option.ShipID,
			WeightClass: string(option.WeightClass), DefaultPrimaryWeaponID: option.DefaultPrimaryWeaponID,
		})
	}
	points := make([]string, 0, len(options.WeaponsByPoint))
	for point := range options.WeaponsByPoint {
		points = append(points, string(point))
	}
	sort.Strings(points)
	for _, pointID := range points {
		point := playerbuild.WeaponPoint(pointID)
		for _, option := range options.WeaponsByPoint[point] {
			state.EligibleWeapons = append(state.EligibleWeapons, game.BuildWeaponOption{
				OwnedWeaponID: option.OwnedWeaponID, WeaponID: option.WeaponID, WeaponPoint: pointID,
			})
		}
	}
	slots := make([]string, 0, len(options.ModulesBySlot))
	for slot := range options.ModulesBySlot {
		slots = append(slots, string(slot))
	}
	sort.Strings(slots)
	for _, slotID := range slots {
		slot := playerbuild.ModuleSlot(slotID)
		for _, option := range options.ModulesBySlot[slot] {
			state.EligibleModules = append(state.EligibleModules, game.BuildModuleOption{
				OwnedModuleID: option.OwnedModuleID, ModuleID: option.ModuleID, ModuleSlot: slotID,
			})
		}
	}
	for _, option := range options.BlockedOptions {
		state.BlockedOptions = append(state.BlockedOptions, game.BuildBlockedOption{
			Kind: option.Kind, OwnedInstanceID: option.OwnedInstanceID, CatalogID: option.CatalogID,
			WeaponPoint: string(option.WeaponPoint), ModuleSlot: string(option.ModuleSlot), ReasonCode: option.ReasonCode,
		})
	}
	return state
}

func loadoutSelectionState(state sessionBuildState) game.LoadoutSelectionState {
	return game.LoadoutSelectionState{
		SelectedOwnedShipID:    state.selection.SelectedOwnedShipID,
		SelectedWeaponsByPoint: stringWeaponPointMap(state.selection.SelectedWeaponsByPoint),
		SelectedModulesBySlot:  stringModuleSlotMap(state.selection.SelectedModulesBySlot),
		StartingAmmoByPoint:    stringAmmoPointMap(state.selection.StartingAmmoByPoint),
		Valid:                  state.valid, ErrorCode: state.errorCode, Message: state.message,
	}
}

func weaponPointMap(values map[string]string) map[playerbuild.WeaponPoint]string {
	result := make(map[playerbuild.WeaponPoint]string, len(values))
	for key, value := range values {
		result[playerbuild.WeaponPoint(key)] = value
	}
	return result
}

func moduleSlotMap(values map[string]string) map[playerbuild.ModuleSlot]string {
	result := make(map[playerbuild.ModuleSlot]string, len(values))
	for key, value := range values {
		result[playerbuild.ModuleSlot(key)] = value
	}
	return result
}

func ammoPointMap(values map[string]int) map[playerbuild.WeaponPoint]int {
	result := make(map[playerbuild.WeaponPoint]int, len(values))
	for key, value := range values {
		result[playerbuild.WeaponPoint(key)] = value
	}
	return result
}

func stringWeaponPointMap(values map[playerbuild.WeaponPoint]string) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[string(key)] = value
	}
	return result
}

func stringModuleSlotMap(values map[playerbuild.ModuleSlot]string) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[string(key)] = value
	}
	return result
}

func stringAmmoPointMap(values map[playerbuild.WeaponPoint]int) map[string]int {
	result := make(map[string]int, len(values))
	for key, value := range values {
		result[string(key)] = value
	}
	return result
}
