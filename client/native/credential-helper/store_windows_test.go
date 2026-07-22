//go:build windows

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsCredentialStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential.bin")
	store := windowsCredentialStore{}
	req := request{
		Service:  "ca.laughingskull.space-rocks",
		Account:  "session",
		BlobPath: path,
		Secret:   "bearer-token",
	}

	if err := store.Save(req); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(contents, []byte(req.Secret)) {
		t.Fatal("encrypted credential file contains the plaintext token")
	}

	secret, err := store.Load(req)
	if err != nil {
		t.Fatal(err)
	}
	if secret != req.Secret {
		t.Fatalf("expected %q, got %q", req.Secret, secret)
	}

	if err := store.Clear(req); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected credential file removal, got %v", err)
	}
}

func TestWindowsCredentialStoreRejectsDifferentIdentityEntropy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential.bin")
	store := windowsCredentialStore{}
	original := request{Service: "service-a", Account: "account", BlobPath: path, Secret: "bearer-token"}
	if err := store.Save(original); err != nil {
		t.Fatal(err)
	}

	changed := original
	changed.Service = "service-b"
	if _, err := store.Load(changed); err == nil {
		t.Fatal("expected DPAPI decryption to fail with different entropy")
	}
}
