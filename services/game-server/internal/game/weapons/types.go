package weapons

type ID string

const BasicCannon ID = "basic_cannon"
const Torpedo ID = "torpedo"

type Slot string

const Primary Slot = "primary"
const Secondary Slot = "secondary"

type AmmoPolicy string

const InfiniteAmmo AmmoPolicy = "infinite"
const LimitedAmmo AmmoPolicy = "limited"

type Equipped struct {
	ID         ID
	AmmoPolicy  AmmoPolicy
}

type PlayerArmory struct {
	Primary   Equipped
	Secondary Equipped
}

type ShipWeapons struct {
	Primary   Equipped
	Secondary Equipped
}

func EmptyEquipped() Equipped {
	return Equipped{}
}

func DefaultPlayerArmory() PlayerArmory {
	return PlayerArmory{
		Primary: Equipped{
			ID:         BasicCannon,
			AmmoPolicy: InfiniteAmmo,
		},
		Secondary: EmptyEquipped(),
	}
}
