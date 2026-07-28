package main

import (
	"os"

	"github.com/Lokee86/space-rocks/services/game-server/internal/authclient"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func buildAuthVerifierFromEnv(startupTraceID string) networking.TokenVerifier {
	if verifier := runtimeScenarioAuthVerifierFromEnv(); verifier != nil {
		return verifier
	}
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
		logging.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"dependency": "auth_verifier", "failure_mode": "initialization_failed"},
		})
		return nil
	}

	return client
}
