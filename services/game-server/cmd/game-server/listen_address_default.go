//go:build !localpackage

package main

func serverListenAddress() string {
	return runtimeScenarioListenAddress(":8080")
}
