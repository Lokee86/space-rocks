package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/Lokee86/space-rocks/services/game-server/internal/authclient"
	"github.com/google/uuid"
)

const (
	runtimeScenarioAuthEnv     = "SPACE_ROCKS_RUNTIME_SCENARIO_AUTH"
	runtimeScenarioTokenPrefix = "runtime-scenario:"
)

type runtimeScenarioTokenVerifier struct{}

func runtimeScenarioAuthVerifierFromEnv() *runtimeScenarioTokenVerifier {
	if strings.TrimSpace(os.Getenv(runtimeScenarioAuthEnv)) != "1" {
		return nil
	}
	return &runtimeScenarioTokenVerifier{}
}

func (runtimeScenarioTokenVerifier) VerifyToken(_ context.Context, rawToken string) (authclient.VerifyResult, error) {
	normalizedToken := strings.TrimSpace(rawToken)
	markerIndex := strings.Index(normalizedToken, runtimeScenarioTokenPrefix)
	if markerIndex < 0 {
		return authclient.VerifyResult{Valid: false}, nil
	}
	clientID := strings.TrimSpace(normalizedToken[markerIndex+len(runtimeScenarioTokenPrefix):])
	if clientID == "" {
		return authclient.VerifyResult{Valid: false}, nil
	}
	accountUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("space-rocks-runtime-scenario:"+clientID))
	userID := int64(binary.BigEndian.Uint64(accountUUID[:8]) & 0x7fffffffffffffff)
	if userID == 0 {
		return authclient.VerifyResult{}, fmt.Errorf("derive runtime scenario user id")
	}
	return authclient.VerifyResult{
		Valid: true,
		Identity: authclient.Identity{
			UserID:      userID,
			AccountID:   accountUUID.String(),
			DisplayName: "Scenario " + clientID,
		},
	}, nil
}
