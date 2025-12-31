package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_GeneratesHeader(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	req, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer req.Body.Close()

	rid := req.Header.Get("X-Request-ID")
	if rid == "" {
		t.Fatalf("expected X-Request-ID header to be set; got empty")
	}
}

func TestRequestIDMiddleware_PassthroughHeader(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/healthz", nil)
	if err != nil {
		t.Fatalf("creating request failed: %v", err)
	}
	req.Header.Set("X-Request-ID", "test-request-id-123")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer res.Body.Close()

	if got := res.Header.Get("X-Request-ID"); got != "test-request-id-123" {
		t.Fatalf("expected X-Request-ID header to be 'test-request-id-123'; got %q", got)
	}
}

func TestRecoverPanicMiddleware_Returns500(t *testing.T) {
	app := newTestApp()

	// create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("BOOM!!!")
	})

	// wrap it with the recover panic middleware
	h := http.Handler(panicHandler)
	h = app.recoverPanicMiddleware(h)
	h = app.requestIDMiddleware(h)

	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500; got %d", res.StatusCode)
	}

	// Request ID header should still be present (outer middleware)
	if rid := res.Header.Get("X-Request-ID"); rid == "" {
		t.Fatalf("expected X-Request-ID header to be set; got empty")
	}
}
