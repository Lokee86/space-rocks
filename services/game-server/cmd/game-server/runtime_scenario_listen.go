package main

import (
	"net"
	"os"
	"strconv"
	"strings"
)

const runtimeScenarioPortEnv = "SPACE_ROCKS_RUNTIME_SCENARIO_PORT"

func runtimeScenarioListenAddress(defaultAddress string) string {
	rawPort := strings.TrimSpace(os.Getenv(runtimeScenarioPortEnv))
	if rawPort == "" {
		return defaultAddress
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil || port < 1 || port > 65535 {
		return defaultAddress
	}
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
}
