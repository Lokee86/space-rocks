//go:build darwin

package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

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
