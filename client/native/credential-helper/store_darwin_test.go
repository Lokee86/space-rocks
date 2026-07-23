//go:build darwin

package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNewPlatformStoreUsesConfiguredKeychainPath(t *testing.T) {
	const keychainPath = "/tmp/space-rocks-smoke.keychain-db"
	t.Setenv(credentialKeychainPathEnvironment, keychainPath)

	store, ok := newPlatformStore().(darwinCredentialStore)
	if !ok {
		t.Fatalf("newPlatformStore() type = %T, want darwinCredentialStore", newPlatformStore())
	}
	if store.keychainPath != keychainPath {
		t.Fatalf("keychain path = %q, want %q", store.keychainPath, keychainPath)
	}
}

func TestDarwinCredentialStoreRoundTripAndUpdate(t *testing.T) {
	store := darwinCredentialStore{}
	req := request{
		Service: fmt.Sprintf(
			"ca.laughingskull.space-rocks.credential-helper-test.%d.%d",
			os.Getpid(),
			time.Now().UnixNano(),
		),
		Account: "session",
		Secret:  "first-token",
	}
	defer store.Clear(req)

	if err := store.Save(req); err != nil {
		t.Fatal(err)
	}
	secret, err := store.Load(req)
	if err != nil {
		t.Fatal(err)
	}
	if secret != req.Secret {
		t.Fatalf("expected %q, got %q", req.Secret, secret)
	}

	req.Secret = "updated-token"
	if err := store.Save(req); err != nil {
		t.Fatal(err)
	}
	secret, err = store.Load(req)
	if err != nil {
		t.Fatal(err)
	}
	if secret != req.Secret {
		t.Fatalf("expected updated secret %q, got %q", req.Secret, secret)
	}

	if err := store.Clear(req); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(req); err != errCredentialNotFound {
		t.Fatalf("expected credential-not-found after clear, got %v", err)
	}
}
