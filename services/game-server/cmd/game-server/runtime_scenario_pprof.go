package main

import (
	"net/http"
	"net/http/pprof"
	"os"
)

const runtimeScenarioPprofEnv = "SPACE_ROCKS_RUNTIME_SCENARIO_PPROF"

func registerRuntimeScenarioPprof(mux *http.ServeMux) {
	if mux == nil || os.Getenv(runtimeScenarioPprofEnv) != "1" {
		return
	}
	mux.Handle("GET /debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("GET /debug/pprof/goroutine", pprof.Handler("goroutine"))
}
