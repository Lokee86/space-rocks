//go:build !windows && !darwin

package main

import "errors"

type unsupportedCredentialStore struct{}

func newPlatformStore() credentialStore {
	return unsupportedCredentialStore{}
}

func (unsupportedCredentialStore) Load(request) (string, error) {
	return "", errors.New("secure credential storage is unsupported on this platform")
}

func (unsupportedCredentialStore) Save(request) error {
	return errors.New("secure credential storage is unsupported on this platform")
}

func (unsupportedCredentialStore) Clear(request) error {
	return errors.New("secure credential storage is unsupported on this platform")
}
