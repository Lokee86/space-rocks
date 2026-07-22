package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const maxRequestBytes = 1 << 20

var errCredentialNotFound = errors.New("credential not found")

type request struct {
	Action   string `json:"action"`
	Service  string `json:"service"`
	Account  string `json:"account"`
	BlobPath string `json:"blob_path,omitempty"`
	Secret   string `json:"secret,omitempty"`
}

type response struct {
	OK     bool   `json:"ok"`
	Secret string `json:"secret,omitempty"`
	Error  string `json:"error,omitempty"`
}

type credentialStore interface {
	Load(request) (string, error)
	Save(request) error
	Clear(request) error
}

func main() {
	if err := run(os.Stdin, os.Stdout, newPlatformStore()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(input io.Reader, output io.Writer, store credentialStore) error {
	decoder := json.NewDecoder(io.LimitReader(input, maxRequestBytes))
	decoder.DisallowUnknownFields()

	var req request
	if err := decoder.Decode(&req); err != nil {
		return writeResponse(output, response{OK: false, Error: "invalid_request"})
	}
	if req.Service == "" || req.Account == "" {
		return writeResponse(output, response{OK: false, Error: "invalid_identity"})
	}

	switch req.Action {
	case "load":
		secret, err := store.Load(req)
		if errors.Is(err, errCredentialNotFound) {
			return writeResponse(output, response{OK: true})
		}
		if err != nil {
			return writeResponse(output, response{OK: false, Error: "load_failed"})
		}
		return writeResponse(output, response{OK: true, Secret: secret})
	case "save":
		if req.Secret == "" {
			return writeResponse(output, response{OK: false, Error: "empty_secret"})
		}
		if err := store.Save(req); err != nil {
			return writeResponse(output, response{OK: false, Error: "save_failed"})
		}
		return writeResponse(output, response{OK: true})
	case "clear":
		if err := store.Clear(req); err != nil && !errors.Is(err, errCredentialNotFound) {
			return writeResponse(output, response{OK: false, Error: "clear_failed"})
		}
		return writeResponse(output, response{OK: true})
	default:
		return writeResponse(output, response{OK: false, Error: "unsupported_action"})
	}
}

func writeResponse(output io.Writer, value response) error {
	writer := bufio.NewWriter(output)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		return err
	}
	return writer.Flush()
}
