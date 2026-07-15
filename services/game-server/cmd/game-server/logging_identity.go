package main

import (
	"errors"
	"os"
	"strings"

	"github.com/Lokee86/space-rocks/shared/go/servicelog"
	"github.com/google/uuid"
)

const (
	envBuildVersion = "BUILD_VERSION"
	envEnvironment  = "ENVIRONMENT"
)

func loadLoggingIdentity(serviceName string) (servicelog.ServiceIdentity, error) {
	version := strings.TrimSpace(os.Getenv(envBuildVersion))
	environment := strings.TrimSpace(os.Getenv(envEnvironment))
	if version == "" {
		return servicelog.ServiceIdentity{}, errors.New("BUILD_VERSION is required for service logging")
	}
	if environment == "" {
		return servicelog.ServiceIdentity{}, errors.New("ENVIRONMENT is required for service logging")
	}
	return servicelog.ServiceIdentity{
		Name:        serviceName,
		Version:     version,
		Environment: environment,
		InstanceID:  uuid.NewString(),
	}, nil
}
