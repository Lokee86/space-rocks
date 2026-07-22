//go:build localpackage

package main

import "os"

var packagedBuildVersion = "local-alpha-dev"

func init() {
	if os.Getenv(envBuildVersion) == "" {
		_ = os.Setenv(envBuildVersion, packagedBuildVersion)
	}
	if os.Getenv(envEnvironment) == "" {
		_ = os.Setenv(envEnvironment, "local-packaged-alpha")
	}
}
