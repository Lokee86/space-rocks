package pickups

import "github.com/Lokee86/space-rocks/server/internal/constants"

type Definition struct {
	Type       PickupType
	ScenePath  string
	Collision  CollisionDefinition
}

type CollisionDefinition struct {
	Shape  CollisionShape
	Radius float64
}

func DefinitionFor(pickupType PickupType) (Definition, bool) {
	if pickupType != TypeOneUp {
		return Definition{}, false
	}

	return Definition{
		Type:      PickupType(constants.PickupOneUpType),
		ScenePath: constants.PickupOneUpScenePath,
		Collision: CollisionDefinition{
			Shape:  CollisionShape(constants.PickupCollisionShapeCircle),
			Radius: constants.PickupOneUpCollisionRadius,
		},
	}, true
}
