package main

import (
	"net/http"
	"testing"
)

func TestRegisterRuntimeScenarioPprofRequiresExplicitEnablement(t *testing.T) {
	t.Setenv(runtimeScenarioPprofEnv, "")
	mux := http.NewServeMux()
	registerRuntimeScenarioPprof(mux)

	_, pattern := mux.Handler(mustPprofRequest(t))
	if pattern != "" {
		t.Fatalf("expected pprof route to stay disabled, got %q", pattern)
	}
}

func TestRegisterRuntimeScenarioPprofRegistersHeapRouteWhenEnabled(t *testing.T) {
	t.Setenv(runtimeScenarioPprofEnv, "1")
	mux := http.NewServeMux()
	registerRuntimeScenarioPprof(mux)

	_, pattern := mux.Handler(mustPprofRequest(t))
	if pattern != "GET /debug/pprof/heap" {
		t.Fatalf("expected heap pprof route, got %q", pattern)
	}
}

func mustPprofRequest(t *testing.T) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/debug/pprof/heap", nil)
	if err != nil {
		t.Fatalf("create pprof request: %v", err)
	}
	return request
}
