package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Lokee86/space-rocks/player-data/httpapi"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/server/internal/networking"
)

func buildPlayerDataRuntime() (*playerdata.Runtime, error) {
	return playerdata.NewRuntimeFromEnv(os.Getenv)
}

func newPlayerDataSink(runtime *playerdata.Runtime) *playerdata.RuntimeSink {
	return playerdata.NewRuntimeSink(runtime)
}

func newPlayerDataProfileHTTPHandler(runtime *playerdata.Runtime, verifier networking.TokenVerifier) http.Handler {
	return httpapi.NewProfileHandler(runtime, httpAPIAuthVerifierAdapter{verifier: verifier})
}

func newPlayerDataLocalProfilesHTTPHandler(runtime *playerdata.Runtime) http.Handler {
	return httpapi.NewLocalProfilesHandler(runtime)
}

type httpAPIAuthVerifierAdapter struct {
	verifier networking.TokenVerifier
}

func (a httpAPIAuthVerifierAdapter) VerifyToken(ctx context.Context, rawToken string) (httpapi.AuthVerificationResult, error) {
	if a.verifier == nil {
		return httpapi.AuthVerificationResult{}, nil
	}

	result, err := a.verifier.VerifyToken(ctx, rawToken)
	if err != nil {
		return httpapi.AuthVerificationResult{}, err
	}

	return httpapi.AuthVerificationResult{
		Valid: result.Valid,
		Identity: httpapi.AuthIdentity{
			AccountID:   result.Identity.AccountID,
			DisplayName: result.Identity.DisplayName,
		},
	}, nil
}
