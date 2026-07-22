//go:build localpackage

package main

import (
	"net"
	"path/filepath"
	"testing"
)

func TestLocalPackageServerBindsOnlyToLoopback(t *testing.T) {
	t.Setenv(envLocalServerPort, "")
	host, port, err := net.SplitHostPort(serverListenAddress())
	if err != nil {
		t.Fatal(err)
	}
	if host != "127.0.0.1" || port != "8080" {
		t.Fatalf("expected loopback-only 127.0.0.1:8080, got %q", serverListenAddress())
	}
}

func TestLocalPackageServerAllowsIsolatedLoopbackSmokePort(t *testing.T) {
	t.Setenv(envLocalServerPort, "43127")
	if got := serverListenAddress(); got != "127.0.0.1:43127" {
		t.Fatalf("expected isolated loopback port, got %q", got)
	}
}

func TestLocalPackageServerRejectsInvalidPortOverride(t *testing.T) {
	for _, value := range []string{"invalid", "0", "65536"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envLocalServerPort, value)
			if got := serverListenAddress(); got != "127.0.0.1:8080" {
				t.Fatalf("expected default loopback port, got %q", got)
			}
		})
	}
}

func TestLocalPackagePathsUseUserConfigDirectory(t *testing.T) {
	path := runtimePath(filepath.Join("player-data", "player-data.sqlite3"))
	if !filepath.IsAbs(path) {
		t.Fatalf("expected absolute packaged runtime path, got %q", path)
	}
	if filepath.Base(filepath.Dir(filepath.Dir(path))) != localPackageDataDirectoryName {
		t.Fatalf("expected Space Rocks runtime directory, got %q", path)
	}
}
