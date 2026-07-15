package diagnosticapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/operationcontext"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeReportService struct {
	created               DiagnosticReport
	got                   DiagnosticReport
	createErr, getErr     error
	createCalls, getCalls int
	createContext         context.Context
	getContext            context.Context
}

type recordingEmitter struct {
	requests []observability.Request
	result   observability.Result
}

func (e *recordingEmitter) Emit(request observability.Request) observability.Result {
	e.requests = append(e.requests, request)
	return e.result
}

func (f *fakeReportService) Create(ctx context.Context, request DiagnosticReportCreateRequest) (DiagnosticReport, error) {
	f.createContext = ctx
	f.createCalls++
	if f.createErr != nil {
		return DiagnosticReport{}, f.createErr
	}
	f.created.Trigger = request.Trigger
	return f.created, nil
}
func (f *fakeReportService) Get(ctx context.Context, _ string) (DiagnosticReport, error) {
	f.getContext = ctx
	f.getCalls++
	return f.got, f.getErr
}

func newTestHandler(t *testing.T, service ReportService, config HandlerConfig) http.Handler {
	t.Helper()
	if config.Emitter == nil {
		config.Emitter = &recordingEmitter{result: observability.Result{Accepted: true}}
	}
	h, err := NewHandler(service, config)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNewHandlerConstructionGuards(t *testing.T) {
	service := &fakeReportService{}
	for name, config := range map[string]HandlerConfig{
		"zero limit":         {Authorize: func(*http.Request) bool { return true }},
		"missing authorizer": {MaxRequestBytes: 10},
		"missing emitter":    {MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return true }},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewHandler(service, config); err == nil {
				t.Fatal("expected construction error")
			}
		})
	}
	if _, err := NewHandler(nil, HandlerConfig{MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return true }}); err == nil {
		t.Fatal("expected nil service error")
	}
}

func TestAuthorizationPrecedesBodyAndService(t *testing.T) {
	service := &fakeReportService{}
	h := newTestHandler(t, service, HandlerConfig{MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return false }})
	r := httptest.NewRequest(http.MethodPost, DiagnosticReportsPath, failingReader{})
	r.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	h.ServeHTTP(response, r)
	if response.Code != http.StatusUnauthorized || service.createCalls != 0 {
		t.Fatalf("status=%d calls=%d", response.Code, service.createCalls)
	}
	if response.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("WWW-Authenticate=%q", response.Header().Get("WWW-Authenticate"))
	}
	if strings.Contains(response.Body.String(), "configured-secret") || strings.Contains(response.Body.String(), "received-secret") {
		t.Fatal("unauthorized response exposed bearer credentials")
	}
}

func TestDiagnosticResponsesSetSecurityHeaders(t *testing.T) {
	h := newTestHandler(t, &fakeReportService{}, HandlerConfig{MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return true }})
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/outside", nil))
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("security headers: Cache-Control=%q X-Content-Type-Options=%q", response.Header().Get("Cache-Control"), response.Header().Get("X-Content-Type-Options"))
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("body read") }

func TestPostMediaTypeAndBodyLimits(t *testing.T) {
	service := &fakeReportService{created: DiagnosticReport{DiagnosticReportID: "report-1"}}
	h := newTestHandler(t, service, HandlerConfig{MaxRequestBytes: 20, Authorize: func(*http.Request) bool { return true }})
	cases := []struct {
		name, contentType, body string
		length                  int64
		status                  int
		code                    string
	}{
		{"media type", "text/plain", "{}", -1, http.StatusUnsupportedMediaType, ErrorCodeUnsupportedMediaType},
		{"content length", "application/json", "{}", 21, http.StatusRequestEntityTooLarge, ErrorCodeRequestTooLarge},
		{"limited reader", "application/json", strings.Repeat("x", 21), -1, http.StatusRequestEntityTooLarge, ErrorCodeRequestTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, DiagnosticReportsPath, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.contentType)
			r.ContentLength = tc.length
			response := httptest.NewRecorder()
			h.ServeHTTP(response, r)
			assertError(t, response, tc.status, tc.code)
		})
	}
}

func TestPostAcceptsJSONCharset(t *testing.T) {
	service := &fakeReportService{created: DiagnosticReport{DiagnosticReportID: "report-1"}}
	h := newTestHandler(t, service, HandlerConfig{MaxRequestBytes: 20, Authorize: func(*http.Request) bool { return true }})
	r := httptest.NewRequest(http.MethodPost, DiagnosticReportsPath, strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	response := httptest.NewRecorder()
	h.ServeHTTP(response, r)
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}

func TestPostStrictAndTrailingJSON(t *testing.T) {
	h := newTestHandler(t, &fakeReportService{}, HandlerConfig{MaxRequestBytes: 100, Authorize: func(*http.Request) bool { return true }})
	for body, code := range map[string]string{"{\"unknown\":true}": ErrorCodeMalformedJSON, "{} {}": ErrorCodeTrailingJSON} {
		r := httptest.NewRequest(http.MethodPost, DiagnosticReportsPath, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		h.ServeHTTP(response, r)
		assertError(t, response, http.StatusBadRequest, code)
	}
}

func TestSuccessResponsesAndMethods(t *testing.T) {
	service := &fakeReportService{created: DiagnosticReport{DiagnosticReportID: "report-1"}, got: DiagnosticReport{DiagnosticReportID: "report-1"}}
	h := newTestHandler(t, service, HandlerConfig{MaxRequestBytes: 100, Authorize: func(*http.Request) bool { return true }})
	post := httptest.NewRequest(http.MethodPost, DiagnosticReportsPath, strings.NewReader(`{}`))
	post.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	h.ServeHTTP(response, post)
	if response.Code != http.StatusCreated || response.Header().Get("Location") != DiagnosticReportItemPathPrefix+"report-1" {
		t.Fatalf("post response: %d location=%q", response.Code, response.Header().Get("Location"))
	}
	get := httptest.NewRequest(http.MethodGet, DiagnosticReportItemPathPrefix+"report-1", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, get)
	if response.Code != http.StatusOK || !json.Valid(response.Body.Bytes()) {
		t.Fatalf("get response: %d %s", response.Code, response.Body)
	}
	method := httptest.NewRequest(http.MethodDelete, DiagnosticReportsPath, nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, method)
	assertError(t, response, http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed)
	if response.Header().Get("Allow") != http.MethodPost {
		t.Fatal("wrong collection Allow header")
	}
}

func TestGetAuthorizationPrecedesServiceAccess(t *testing.T) {
	service := &fakeReportService{}
	h := newTestHandler(t, service, HandlerConfig{MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return false }})
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, DiagnosticReportItemPathPrefix+"report-1", nil))
	if response.Code != http.StatusUnauthorized || service.getCalls != 0 {
		t.Fatalf("status=%d calls=%d", response.Code, service.getCalls)
	}
}

func TestPathValidationAndServiceErrors(t *testing.T) {
	for _, path := range []string{DiagnosticReportItemPathPrefix, DiagnosticReportItemPathPrefix + "a/b"} {
		response := httptest.NewRecorder()
		h := newTestHandler(t, &fakeReportService{}, HandlerConfig{MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return true }})
		h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		assertError(t, response, http.StatusBadRequest, ErrorCodeInvalidReportID)
	}
	response := httptest.NewRecorder()
	h := newTestHandler(t, &fakeReportService{}, HandlerConfig{MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return true }})
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/outside", nil))
	assertError(t, response, http.StatusNotFound, ErrorCodeNotFound)
	for serviceErr, expected := range map[error]struct {
		status int
		code   string
	}{
		ErrInvalidReport:     {http.StatusUnprocessableEntity, ErrorCodeInvalidDiagnosticReport},
		ErrRejectedReport:    {http.StatusUnprocessableEntity, ErrorCodeDiagnosticReportRejected},
		ErrInvalidReportID:   {http.StatusBadRequest, ErrorCodeInvalidReportID},
		ErrReportNotFound:    {http.StatusNotFound, ErrorCodeDiagnosticReportNotFound},
		ErrUnavailable:       {http.StatusServiceUnavailable, ErrorCodeServiceUnavailable},
		errors.New("secret"): {http.StatusServiceUnavailable, ErrorCodeServiceUnavailable},
	} {
		service := &fakeReportService{getErr: serviceErr}
		h := newTestHandler(t, service, HandlerConfig{MaxRequestBytes: 10, Authorize: func(*http.Request) bool { return true }})
		response := httptest.NewRecorder()
		h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, DiagnosticReportItemPathPrefix+"id", nil))
		if response.Code != expected.status || strings.Contains(response.Body.String(), "secret") {
			t.Fatalf("err=%v status=%d body=%s", serviceErr, response.Code, response.Body)
		}
		var body ErrorResponse
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Code != expected.code {
			t.Fatalf("err=%v code=%q want=%q", serviceErr, body.Code, expected.code)
		}
	}
}

func TestRejectedReportContinuesValidTraceWithoutPayloadLeakOrBehaviorChange(t *testing.T) {
	const submittedTrace = "550e8400-e29b-41d4-a716-446655440020"
	const requestID = "550e8400-e29b-41d4-a716-446655440021"
	ids := []string{"550e8400-e29b-41d4-a716-446655440022", requestID}
	index := 0
	emitter := &recordingEmitter{result: observability.Result{WriteFailed: true}}
	service := &fakeReportService{createErr: ErrRejectedReport}
	h := newTestHandler(t, service, HandlerConfig{
		MaxRequestBytes: 1024,
		Authorize:       func(*http.Request) bool { return true },
		Emitter:         emitter,
		NewUUID: func() string {
			value := ids[index]
			index++
			return value
		},
	})
	body := `{"correlation":{"trace_id":"` + submittedTrace + `"},"user_description":"unsafe-submitted-value"}`
	request := httptest.NewRequest(http.MethodPost, DiagnosticReportsPath, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	assertError(t, response, http.StatusUnprocessableEntity, ErrorCodeDiagnosticReportRejected)
	if len(emitter.requests) != 1 || emitter.requests[0].Event != observability.EventNameDiagnosticReportRejected {
		t.Fatalf("requests=%#v", emitter.requests)
	}
	emitted := emitter.requests[0]
	if emitted.Context.TraceID != submittedTrace || emitted.Context.RequestID != requestID || emitted.Context.Route != DiagnosticReportsPath {
		t.Fatalf("context=%#v", emitted.Context)
	}
	operation, ok := operationcontext.From(service.createContext)
	if !ok || operation.TraceID != submittedTrace || operation.RequestID != requestID || operation.Route != DiagnosticReportsPath {
		t.Fatalf("operation=%#v ok=%v", operation, ok)
	}
	encoded, err := json.Marshal(emitted)
	if err != nil || strings.Contains(string(encoded), "unsafe-submitted-value") {
		t.Fatalf("emitted=%s err=%v", encoded, err)
	}
}
func assertError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status=%d content-type=%q body=%s", response.Code, response.Header().Get("Content-Type"), response.Body)
	}
	var body ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Code != code {
		t.Fatalf("error=%q body=%s", code, response.Body)
	}
}

var _ io.Reader = failingReader{}
