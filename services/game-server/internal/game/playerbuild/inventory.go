package playerbuild

const (
	InventoryIdentityGuest                = "guest"
	InventoryIdentityLocalProfile         = "local_profile"
	InventoryIdentityAuthenticatedAccount = "authenticated_account"
)

type InventoryIdentity struct {
	Kind           string
	AccountID      string
	LocalProfileID string
}

type InventoryLoadRequest struct {
	PlayMode string
	TraceID  string
}

type InventoryLoadResult struct {
	Found               bool
	Persisted           bool
	SynthesizedFallback bool
	RepairAttempted     bool
	Inventory           Inventory
	ErrorCode           string
	Message             string
}

type HardwiredEquipment struct {
	HardwiredID string
	EquipmentID string
	State       string
}

type OwnedShip struct {
	OwnedShipID        string
	ShipID             string
	HardwiredEquipment []HardwiredEquipment
	State              string
}

type OwnedWeapon struct {
	OwnedWeaponID string
	WeaponID      string
	State         string
}

type OwnedModule struct {
	OwnedModuleID string
	ModuleID      string
	State         string
}

type Inventory struct {
	InventoryVersion   int
	OwnedShips         []OwnedShip
	OwnedWeapons       []OwnedWeapon
	OwnedModules       []OwnedModule
	DefaultOwnedShipID string
}
