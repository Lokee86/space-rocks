package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/diagnosticclient"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/diagnostics"
	"github.com/google/uuid"
)

const (
	exitOK = iota
	exitArguments
	exitConfiguration
	exitSubmission
	exitRetrieval
	exitOutput
)

type dependencies struct {
	httpClient *http.Client
	now        func() time.Time
	newUUID    func() string
	environ    func(string) string
}

func defaultDependencies() dependencies {
	return dependencies{httpClient: http.DefaultClient, now: time.Now, newUUID: func() string { return uuid.NewString() }, environ: os.Getenv}
}

func run(args []string, stdout, stderr io.Writer, deps dependencies) int {
	flags := flag.NewFlagSet("diagnostic-submit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	baseURL := flags.String("base-url", "", "diagnostic aggregator URL")
	description := flags.String("description", "", "manual diagnostic description")
	timeout := flags.Duration("timeout", 10*time.Second, "request timeout")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(stderr, "diagnostic-submit: invalid arguments")
		return exitArguments
	}
	if *baseURL == "" {
		*baseURL = deps.environ("DIAGNOSTIC_AGGREGATOR_URL")
	}
	token := deps.environ("DIAGNOSTIC_AGGREGATOR_TOKEN")
	if *baseURL == "" || token == "" || *timeout <= 0 {
		fmt.Fprintln(stderr, "diagnostic-submit: missing or invalid configuration")
		return exitConfiguration
	}
	client, err := diagnosticclient.New(diagnosticclient.Config{BaseURL: *baseURL, BearerToken: token, HTTPClient: deps.httpClient, MaxResponseBytes: 4 << 20})
	if err != nil {
		fmt.Fprintln(stderr, "diagnostic-submit: invalid configuration")
		return exitConfiguration
	}
	submission := buildSubmission(deps.now, deps.newUUID, *description)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	report, err := client.Create(ctx, submission)
	if err != nil {
		fmt.Fprintln(stderr, "diagnostic-submit: submission failed")
		return exitSubmission
	}
	report, err = client.Get(ctx, report.DiagnosticReportID)
	if err != nil {
		fmt.Fprintln(stderr, "diagnostic-submit: retrieval failed")
		return exitRetrieval
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(stderr, "diagnostic-submit: output failed")
		return exitOutput
	}
	return exitOK
}

func buildSubmission(now func() time.Time, newUUID func() string, description string) diagnostics.DiagnosticSubmission {
	created := now().UTC()
	serviceInstanceID := newUUID()
	traceID, sessionID := newUUID(), newUUID()
	event := func(name, level string) json.RawMessage {
		payload := map[string]any{
			"event_id": newUUID(), "timestamp": created.Format(time.RFC3339), "level": level,
			"event": name, "service": "diagnostic-submit", "schema_version": 1,
			"service_instance_id": serviceInstanceID, "environment": "development", "build_version": "manual",
			"trace_id": traceID, "session_id": sessionID,
			"fields": map[string]any{"source": "manual_cli"},
		}
		data, _ := json.Marshal(payload)
		return data
	}
	return diagnostics.DiagnosticSubmission{
		ReportVersion: 1, Trigger: diagnostics.DiagnosticTriggerDevelopmentSubmission, SubmittedAt: created,
		Source:          diagnostics.DiagnosticSourceContext{Service: "diagnostic-submit", ServiceInstanceID: serviceInstanceID, Environment: "development", BuildVersion: "manual", Platform: runtime.GOOS + "/" + runtime.GOARCH},
		Correlation:     diagnostics.DiagnosticCorrelationContext{TraceID: traceID, SessionID: sessionID},
		UserDescription: description, Events: []json.RawMessage{event("diagnostic_submission_started", "info"), event("development_failure_observed", "error")},
	}
}
