package main

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPServerTimeouts(t *testing.T) {
	server := newHTTPServer(http.NewServeMux())
	if server.Addr != ":8080" || server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 15*time.Second || server.WriteTimeout != 15*time.Second || server.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected HTTP server configuration: %+v", server)
	}
}
