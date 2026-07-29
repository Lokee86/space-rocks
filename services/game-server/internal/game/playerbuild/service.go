package playerbuild

import "fmt"

type InventoryLoader interface {
	Load(identity InventoryIdentity, context InventoryLoadRequest) (InventoryLoadResult, error)
}

type LoadedBuildContext struct {
	InventoryLoad InventoryLoadResult
	Rules         Rules
	Options       EligibleBuildOptions
}

type Service struct {
	loader  InventoryLoader
	catalog Catalog
}

func NewService(loader InventoryLoader, catalog Catalog) (*Service, error) {
	if loader == nil {
		return nil, fmt.Errorf("inventory loader is required")
	}
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	return &Service{loader: loader, catalog: catalog}, nil
}

func (service *Service) LoadOptions(playerID string, identity InventoryIdentity, context InventoryLoadRequest, rules Rules) (LoadedBuildContext, error) {
	if err := ValidateRules(rules); err != nil {
		return LoadedBuildContext{}, err
	}
	loaded, err := service.loader.Load(identity, context)
	if err != nil {
		return LoadedBuildContext{}, err
	}
	normalizedRules := NormalizeRules(rules)
	return LoadedBuildContext{
		InventoryLoad: loaded,
		Rules:         normalizedRules,
		Options:       ComputeEligibility(playerID, loaded.Inventory, service.catalog, normalizedRules),
	}, nil
}

func (service *Service) ResolveSelection(context LoadedBuildContext, selection LoadoutSelection) (ResolvedPlayerBuild, error) {
	return Resolve(selection, context.InventoryLoad.Inventory, service.catalog, context.Rules)
}
