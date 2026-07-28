//go:build localpackage

package main

import (
	"os"
	"strconv"
	"strings"
)

const envLocalServerPort = "SPACE_ROCKS_LOCAL_SERVER_PORT"

func serverListenAddress() string {
	port := strings.TrimSpace(os.Getenv(envLocalServerPort))
	if port == "" {
		port = "8080"
	}
	parsed, err := strconv.Atoi(port)
	if err != nil || parsed < 1 || parsed > 65535 {
		port = "8080"
	}
	return runtimeScenarioListenAddress("127.0.0.1:" + port)
}
