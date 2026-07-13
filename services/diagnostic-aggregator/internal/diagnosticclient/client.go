package diagnosticclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticapi"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
)

var (
	ErrInvalidConfig    = errors.New("diagnosticclient: invalid configuration")
	ErrUnexpectedStatus = errors.New("diagnosticclient: unexpected response status")
	ErrResponseTooLarge = errors.New("diagnosticclient: response too large")
	ErrMalformedJSON    = errors.New("diagnosticclient: malformed response JSON")
	ErrTrailingJSON     = errors.New("diagnosticclient: trailing response JSON")
)

type Config struct {
	BaseURL          string
	BearerToken      string
	HTTPClient       *http.Client
	MaxResponseBytes int64
}

type Client struct {
	baseURL          *url.URL
	bearerToken      string
	httpClient       *http.Client
	maxResponseBytes int64
}

type StatusError struct{ Code int }

func (e *StatusError) Error() string {
	return fmt.Sprintf("diagnosticclient: unexpected response status %d", e.Code)
}
func (e *StatusError) Unwrap() error { return ErrUnexpectedStatus }

func New(config Config) (*Client, error) {
	parsed, err := url.Parse(config.BaseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || config.HTTPClient == nil || config.MaxResponseBytes <= 0 {
		return nil, ErrInvalidConfig
	}
	return &Client{baseURL: parsed, bearerToken: config.BearerToken, httpClient: config.HTTPClient, maxResponseBytes: config.MaxResponseBytes}, nil
}

func (c *Client) Create(ctx context.Context, submission diagnostics.DiagnosticSubmission) (diagnostics.DiagnosticReport, error) {
	body, err := json.Marshal(submission)
	if err != nil {
		return diagnostics.DiagnosticReport{}, err
	}
	return c.do(ctx, http.MethodPost, diagnosticapi.DiagnosticReportsPath, bytes.NewReader(body), http.StatusCreated)
}

func (c *Client) Get(ctx context.Context, reportID string) (diagnostics.DiagnosticReport, error) {
	return c.do(ctx, http.MethodGet, diagnosticapi.DiagnosticReportItemPathPrefix+reportID, nil, http.StatusOK)
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, expectedStatus int) (diagnostics.DiagnosticReport, error) {
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path
	endpoint.RawQuery = ""
	endpoint.Fragment = ""
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return diagnostics.DiagnosticReport{}, err
	}
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return diagnostics.DiagnosticReport{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != expectedStatus {
		return diagnostics.DiagnosticReport{}, &StatusError{Code: response.StatusCode}
	}
	return decodeResponse(response.Body, c.maxResponseBytes)
}

func decodeResponse(body io.Reader, maxBytes int64) (diagnostics.DiagnosticReport, error) {
	limited := io.LimitReader(body, maxBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return diagnostics.DiagnosticReport{}, err
	}
	if int64(len(payload)) > maxBytes {
		return diagnostics.DiagnosticReport{}, ErrResponseTooLarge
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var report diagnostics.DiagnosticReport
	if err := decoder.Decode(&report); err != nil {
		return diagnostics.DiagnosticReport{}, ErrMalformedJSON
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return diagnostics.DiagnosticReport{}, ErrTrailingJSON
	}
	return report, nil
}
