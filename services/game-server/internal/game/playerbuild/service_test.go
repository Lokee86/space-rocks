package playerbuild

import (
	"errors"
	"testing"
)

type fakeInventoryLoader struct {
	result InventoryLoadResult
	err    error
}

func (loader fakeInventoryLoader) Load(InventoryIdentity, InventoryLoadRequest) (InventoryLoadResult, error) {
	return loader.result, loader.err
}

func TestServiceLoadsHangarOptionsAndResolvesSelection(t *testing.T) {
	loader := fakeInventoryLoader{result: InventoryLoadResult{
		Found:     true,
		Persisted: true,
		Inventory: testInventory(),
	}}
	service, err := NewService(loader, testCatalog())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	context, err := service.LoadOptions(
		"player-1",
		InventoryIdentity{Kind: InventoryIdentityLocalProfile},
		InventoryLoadRequest{},
		Rules{ModeID: "arcade_survival"},
	)
	if err != nil {
		t.Fatalf("load options: %v", err)
	}
	build, err := service.ResolveSelection(context, context.Options.FallbackLoadout)
	if err != nil {
		t.Fatalf("resolve selection: %v", err)
	}
	if build.ShipID != ShipVWing {
		t.Fatalf("expected %q, got %q", ShipVWing, build.ShipID)
	}
}

func TestServicePropagatesInventoryLoadFailure(t *testing.T) {
	service, err := NewService(fakeInventoryLoader{err: errors.New("offline")}, DefaultCatalog())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if _, err := service.LoadOptions("player-1", InventoryIdentity{}, InventoryLoadRequest{}, Rules{}); err == nil {
		t.Fatal("expected inventory load failure")
	}
}
