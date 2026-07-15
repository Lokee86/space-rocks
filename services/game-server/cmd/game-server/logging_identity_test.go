package main

import (
	"testing"

	"github.com/google/uuid"
)

func TestLoadLoggingIdentityRequiresBuildAndEnvironment(t *testing.T) {
	t.Setenv(envBuildVersion, "")
	t.Setenv(envEnvironment, "test")
	if _, err := loadLoggingIdentity("game-server"); err == nil {
		t.Fatal("expected missing build version error")
	}
	t.Setenv(envBuildVersion, "test-build")
	t.Setenv(envEnvironment, "")
	if _, err := loadLoggingIdentity("game-server"); err == nil {
		t.Fatal("expected missing environment error")
	}
}

func TestLoadLoggingIdentityPopulatesStartupOwnedFields(t *testing.T) {
	t.Setenv(envBuildVersion, "test-build")
	t.Setenv(envEnvironment, "test")
	identity, err := loadLoggingIdentity("game-server")
	if err != nil {
		t.Fatal(err)
	}
	if identity.Name != "game-server" || identity.Version != "test-build" || identity.Environment != "test" {
		t.Fatalf("identity=%#v", identity)
	}
	if _, err := uuid.Parse(identity.InstanceID); err != nil {
		t.Fatalf("instance id=%q: %v", identity.InstanceID, err)
	}
}
