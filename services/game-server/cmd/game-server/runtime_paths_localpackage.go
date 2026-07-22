//go:build localpackage

package main

import (
	"os"
	"path/filepath"
)

const localPackageDataDirectoryName = "Space Rocks"

func runtimePath(relativePath string) string {
	configDirectory, err := os.UserConfigDir()
	if err != nil || configDirectory == "" {
		configDirectory = "."
	}
	return filepath.Join(configDirectory, localPackageDataDirectoryName, relativePath)
}
