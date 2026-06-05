package pickups

import "github.com/Lokee86/space-rocks/server/internal/constants"

type Definition struct {
	Type      PickupType
	ScenePath string
}

func DefinitionFor(pickupType PickupType) (Definition, bool) {
	if pickupType != TypeOneUp {
		return Definition{}, false
	}

	return Definition{
		Type:      PickupType(constants.PickupOneUpType),
		ScenePath: constants.PickupOneUpScenePath,
	}, true
}
