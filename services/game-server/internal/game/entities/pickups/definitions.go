package pickups

import "github.com/Lokee86/space-rocks/server/internal/constants"

type Definition struct {
	Type            PickupType
	Class           PickupClass
	Health          int
	LifespanSeconds float64
}

func DefinitionFor(pickupType PickupType) (Definition, bool) {
	switch pickupType {
	case TypeOneUp:
		return Definition{
			Type:            PickupType(constants.PickupOneUpType),
			Class:           PickupClass(constants.PickupOneUpClass),
			Health:          constants.PickupOneUpHealth,
			LifespanSeconds: constants.PickupOneUpLifespan,
		}, true
	case TypeTorpedo:
		return Definition{
			Type:            PickupType(constants.PickupTorpedoType),
			Class:           PickupClass(constants.PickupTorpedoClass),
			Health:          constants.PickupTorpedoHealth,
			LifespanSeconds: constants.PickupTorpedoLifespan,
		}, true
	}

	return Definition{}, false
}
