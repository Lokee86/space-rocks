package diagnostics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type httpBundleService struct {
	created      Bundle
	createErr    error
	getErr       error
	createTrace  string
	createLimit  int
	getID        string
}

func (s *httpBundleService) Create(_ context.Context, traceID string, limit int) (Bundle, error) {
	s.createTrace, s.createLimit = traceID, limit
	return s.created, s.createErr
}

func (s *httpBundleService) Get(_ context.Context, reportID string) (Bundle, error) {
	s.getID = reportID
	return s.created, s.getErr
}

func callHTTP(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
	return response
}

func errorCode(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	return body["error"]
}

func TestHTTPCreateSuccessAndArguments(t *testing.T) {
	service := &httpBundleService{created: Bundle{DiagnosticReportID: "report"}}
	response := callHTTP(NewHTTPHandler(service, 1024), http.MethodPost, BundlesPath, `{"trace_id":"trace","limit":7}`)
	if response.Code != http.StatusCreated || service.createTrace != "trace" || service.createLimit != 7 { t.Fatalf("status=%d service=%#v", response.Code, service) }
	if response.Header().Get("Content-Type") != "application/json" { t.Fatal("missing JSON content type") }
	var bundle Bundle
	if err := json.Unmarshal(response.Body.Bytes(), &bundle); err != nil || bundle.DiagnosticReportID != "report" { t.Fatalf("body=%s err=%v", response.Body, err) }
}

func TestHTTPGetSuccessAndArgument(t *testing.T) {
	service := &httpBundleService{created: Bundle{DiagnosticReportID: "report"}}
	response := callHTTP(NewHTTPHandler(service, 1024), http.MethodGet, BundleItemPathPrefix+"report", "")
	if response.Code != http.StatusOK || service.getID != "report" { t.Fatalf("status=%d id=%q", response.Code, service.getID) }
}

func TestHTTPStrictBodyValidation(t *testing.T) {
	for _, test := range []struct { name, body, want string }{
		{"unknown field", `{"trace_id":"x","other":1}`, "invalid_json"},
		{"trailing document", `{"trace_id":"x"}{}`, "invalid_json"},
		{"empty body", "", "invalid_json"},
		{"missing trace", `{}`, "invalid_trace_id"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := callHTTP(NewHTTPHandler(&httpBundleService{}, 1024), http.MethodPost, BundlesPath, test.body)
			if response.Code != http.StatusBadRequest || errorCode(t, response) != test.want { t.Fatalf("status=%d code=%s", response.Code, errorCode(t, response)) }
		})
	}
	response := callHTTP(NewHTTPHandler(&httpBundleService{}, 4), http.MethodPost, BundlesPath, `{"trace_id":"x"}`)
	if response.Code != http.StatusBadRequest || errorCode(t, response) != "request_too_large" { t.Fatalf("status=%d code=%s", response.Code, errorCode(t, response)) }
}

func TestHTTPPathAndMethodValidation(t *testing.T) {
	handler := NewHTTPHandler(&httpBundleService{}, 1024)
	response := callHTTP(handler, http.MethodGet, BundlesPath, "")
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodPost || errorCode(t, response) != "method_not_allowed" { t.Fatal(response.Code) }
	response = callHTTP(handler, http.MethodPost, BundleItemPathPrefix+"report", "")
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodGet { t.Fatal(response.Code) }
	response = callHTTP(handler, http.MethodGet, BundleItemPathPrefix+"report/extra", "")
	if response.Code != http.StatusBadRequest || errorCode(t, response) != "invalid_diagnostic_report_id" { t.Fatal(response.Code) }
	response = callHTTP(handler, http.MethodGet, "/unknown", "")
	if response.Code != http.StatusNotFound || errorCode(t, response) != "not_found" { t.Fatal(response.Code) }
}

func TestHTTPServiceErrorMapping(t *testing.T) {
	service := &httpBundleService{}
	handler := NewHTTPHandler(service, 1024)
	cases := []struct { name, path, body string
		method                   string
		serviceErr               error
		wantStatus               int
		wantCode                 string }{
		{"invalid trace", BundlesPath, `{"trace_id":"x"}`, http.MethodPost, ErrInvalidTraceID, 400, "invalid_trace_id"},
		{"invalid limit", BundlesPath, `{"trace_id":"x"}`, http.MethodPost, ErrInvalidLimit, 400, "invalid_limit"},
		{"invalid report", BundleItemPathPrefix+"bad", "", http.MethodGet, ErrInvalidDiagnosticReportID, 400, "invalid_diagnostic_report_id"},
		{"trace missing", BundlesPath, `{"trace_id":"x"}`, http.MethodPost, ErrNoEvents, 404, "not_found"},
		{"bundle missing", BundleItemPathPrefix+"report", "", http.MethodGet, ErrBundleNotFound, 404, "not_found"},
		{"internal create", BundlesPath, `{"trace_id":"x"}`, http.MethodPost, errors.New("secret"), 503, "service_unavailable"},
		{"internal get", BundleItemPathPrefix+"report", "", http.MethodGet, errors.New("secret"), 503, "service_unavailable"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			service.createErr, service.getErr = nil, nil
			if test.method == http.MethodPost { service.createErr = test.serviceErr } else { service.getErr = test.serviceErr }
			response := callHTTP(handler, test.method, test.path, test.body)
			if response.Code != test.wantStatus || errorCode(t, response) != test.wantCode { t.Fatalf("status=%d code=%s", response.Code, errorCode(t, response)) }
		})
	}
}

func TestHTTPNilServiceUnavailable(t *testing.T) {
	for _, request := range [][2]string{{http.MethodPost, BundlesPath}, {http.MethodGet, BundleItemPathPrefix + "report"}} {
		response := callHTTP(NewHTTPHandler(nil, 1024), request[0], request[1], `{"trace_id":"x"}`)
		if response.Code != http.StatusServiceUnavailable || errorCode(t, response) != "service_unavailable" { t.Fatal(response.Code) }
	}
}
