package main

import (
	"os"

	"github.com/Lokee86/space-rocks/server/internal/authclient"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking"
)

func buildAuthVerifierFromEnv() networking.TokenVerifier {
	baseURL := os.Getenv("API_SERVER_BASE_URL")
	internalToken := os.Getenv("GAME_SERVER_INTERNAL_TOKEN")
	if baseURL == "" || internalToken == "" {
		return nil
	}

	client, err := authclient.New(authclient.Config{
		BaseURL:       baseURL,
		InternalToken: internalToken,
	})
	if err != nil {
		logging.Server.Error("auth verifier initialization failed", err)
		return nil
	}

	return client
}
