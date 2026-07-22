//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	cryptProtectUIForbidden = 0x1
	windowsBlobMagic        = "SRDPAPI1\n"
)

var (
	crypt32                = syscall.NewLazyDLL("crypt32.dll")
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procCryptProtectData   = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
	procLocalFree          = kernel32.NewProc("LocalFree")
)

type dataBlob struct {
	Size uint32
	Data *byte
}

type windowsCredentialStore struct{}

func newPlatformStore() credentialStore {
	return windowsCredentialStore{}
}

func (windowsCredentialStore) Load(req request) (string, error) {
	if req.BlobPath == "" {
		return "", errors.New("blob path is required")
	}
	encoded, err := os.ReadFile(req.BlobPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", errCredentialNotFound
	}
	if err != nil {
		return "", err
	}
	if len(encoded) < len(windowsBlobMagic) || string(encoded[:len(windowsBlobMagic)]) != windowsBlobMagic {
		return "", errors.New("credential blob has an unknown format")
	}

	plaintext, err := unprotectData(encoded[len(windowsBlobMagic):], entropyFor(req))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (windowsCredentialStore) Save(req request) error {
	if req.BlobPath == "" {
		return errors.New("blob path is required")
	}
	ciphertext, err := protectData([]byte(req.Secret), entropyFor(req))
	if err != nil {
		return err
	}
	contents := append([]byte(windowsBlobMagic), ciphertext...)
	return atomicWrite(req.BlobPath, contents)
}

func (windowsCredentialStore) Clear(req request) error {
	if req.BlobPath == "" {
		return errors.New("blob path is required")
	}
	if err := os.Remove(req.BlobPath); errors.Is(err, os.ErrNotExist) {
		return errCredentialNotFound
	} else {
		return err
	}
}

func entropyFor(req request) []byte {
	return []byte(req.Service + "\x00" + req.Account)
}

func protectData(plaintext []byte, entropy []byte) ([]byte, error) {
	input := blobFromBytes(plaintext)
	optionalEntropy := blobFromBytes(entropy)
	var output dataBlob

	result, _, callErr := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(&input)),
		0,
		uintptr(unsafe.Pointer(&optionalEntropy)),
		0,
		0,
		cryptProtectUIForbidden,
		uintptr(unsafe.Pointer(&output)),
	)
	runtime.KeepAlive(plaintext)
	runtime.KeepAlive(entropy)
	if result == 0 {
		return nil, windowsCallError("CryptProtectData", callErr)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(output.Data)))
	return copyBlob(output), nil
}

func unprotectData(ciphertext []byte, entropy []byte) ([]byte, error) {
	input := blobFromBytes(ciphertext)
	optionalEntropy := blobFromBytes(entropy)
	var output dataBlob

	result, _, callErr := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&input)),
		0,
		uintptr(unsafe.Pointer(&optionalEntropy)),
		0,
		0,
		cryptProtectUIForbidden,
		uintptr(unsafe.Pointer(&output)),
	)
	runtime.KeepAlive(ciphertext)
	runtime.KeepAlive(entropy)
	if result == 0 {
		return nil, windowsCallError("CryptUnprotectData", callErr)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(output.Data)))
	return copyBlob(output), nil
}

func blobFromBytes(value []byte) dataBlob {
	if len(value) == 0 {
		return dataBlob{}
	}
	return dataBlob{Size: uint32(len(value)), Data: &value[0]}
}

func copyBlob(blob dataBlob) []byte {
	if blob.Size == 0 || blob.Data == nil {
		return nil
	}
	return append([]byte(nil), unsafe.Slice(blob.Data, int(blob.Size))...)
}

func windowsCallError(operation string, err error) error {
	if err == nil || errors.Is(err, syscall.Errno(0)) {
		return fmt.Errorf("%s failed", operation)
	}
	return fmt.Errorf("%s failed: %w", operation, err)
}

func atomicWrite(path string, contents []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".credential-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(temporaryPath, path)
}
