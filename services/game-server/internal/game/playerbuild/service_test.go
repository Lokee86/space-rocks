package playerbuild

import (
	"errors"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type fakeInventoryLoader struct {
	result protocol.PlayerDataLoadHangarInventoryResult
	err    error
}

func (loader fakeInventoryLoader) Load(protocol.PlayerDataIdentity, protocol.PlayerDataRequestContext) (protocol.PlayerDataLoadHangarInventoryResult, error) {
	return loader.result, loader.err
}

func TestServiceLoadsHangarOptionsAndResolvesSelection(t *testing.T) {
	loader := fakeInventoryLoader{result: protocol.PlayerDataLoadHangarInventoryResult{
		Found:     true,
		Persisted: true,
		Inventory: testInventory(),
	}}
	service, err := NewService(loader, testCatalog())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	context, err := service.LoadOptions("player-1", protocol.PlayerDataIdentity{IdentityKind: "local_profile"}, protocol.PlayerDataRequestContext{}, Rules{ModeID: "arcade_survival"})
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
	if _, err := service.LoadOptions("player-1", protocol.PlayerDataIdentity{}, protocol.PlayerDataRequestContext{}, Rules{}); err == nil {
		t.Fatal("expected inventory load failure")
	}
}
