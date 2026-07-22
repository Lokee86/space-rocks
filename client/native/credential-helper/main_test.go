package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

type fakeCredentialStore struct {
	secret   string
	loadErr  error
	saveErr  error
	clearErr error
}

func (store *fakeCredentialStore) Load(request) (string, error) {
	return store.secret, store.loadErr
}

func (store *fakeCredentialStore) Save(req request) error {
	if store.saveErr == nil {
		store.secret = req.Secret
	}
	return store.saveErr
}

func (store *fakeCredentialStore) Clear(request) error {
	if store.clearErr == nil {
		store.secret = ""
	}
	return store.clearErr
}

func TestRunSaveLoadAndClear(t *testing.T) {
	store := &fakeCredentialStore{}
	assertResponse(t, store, request{Action: "save", Service: "service", Account: "account", Secret: "token"}, response{OK: true})
	assertResponse(t, store, request{Action: "load", Service: "service", Account: "account"}, response{OK: true, Secret: "token"})
	assertResponse(t, store, request{Action: "clear", Service: "service", Account: "account"}, response{OK: true})
	assertResponse(t, store, request{Action: "load", Service: "service", Account: "account"}, response{OK: true})
}

func TestRunTreatsMissingCredentialAsSignedOut(t *testing.T) {
	store := &fakeCredentialStore{loadErr: errCredentialNotFound}
	assertResponse(t, store, request{Action: "load", Service: "service", Account: "account"}, response{OK: true})
}

func TestRunDoesNotExposeBackendErrors(t *testing.T) {
	store := &fakeCredentialStore{loadErr: errors.New("sensitive backend detail")}
	assertResponse(t, store, request{Action: "load", Service: "service", Account: "account"}, response{OK: false, Error: "load_failed"})
}

func TestRunRejectsUnknownFields(t *testing.T) {
	var output bytes.Buffer
	payload, err := json.Marshal(map[string]string{
		"action":  "load",
		"service": "service",
		"account": "account",
		"token":   "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := run(bytes.NewReader(payload), &output, &fakeCredentialStore{}); err != nil {
		t.Fatal(err)
	}
	var actual response
	if err := json.Unmarshal(output.Bytes(), &actual); err != nil {
		t.Fatal(err)
	}
	if actual != (response{OK: false, Error: "invalid_request"}) {
		t.Fatalf("unexpected response: %#v", actual)
	}
}

func assertResponse(t *testing.T, store credentialStore, req request, expected response) {
	t.Helper()
	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(bytes.NewReader(payload), &output, store); err != nil {
		t.Fatal(err)
	}
	var actual response
	if err := json.Unmarshal(output.Bytes(), &actual); err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}
