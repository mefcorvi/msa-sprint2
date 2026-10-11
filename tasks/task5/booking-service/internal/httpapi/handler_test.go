package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func get(t *testing.T, h http.Handler, path string) (int, string) {
	t.Helper()
	rec := serve(h, path)
	return rec.Code, rec.Body.String()
}

func serve(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func ready(context.Context) error    { return nil }
func notReady(context.Context) error { return errors.New("database unavailable") }

func TestPingAnswersPong(t *testing.T) {
	code, body := get(t, NewHandler(notReady, false, "v1"), "/ping")

	if code != http.StatusOK || body != "pong" {
		t.Fatalf("got %d %q", code, body)
	}
}

func TestReadyReflectsDependencies(t *testing.T) {
	if code, _ := get(t, NewHandler(ready, false, "v1"), "/ready"); code != http.StatusOK {
		t.Fatalf("ready service answered %d", code)
	}
	if code, _ := get(t, NewHandler(notReady, false, "v1"), "/ready"); code != http.StatusServiceUnavailable {
		t.Fatalf("not ready service answered %d", code)
	}
}

func TestFeatureEndpointFollowsFlag(t *testing.T) {
	if code, _ := get(t, NewHandler(ready, false, "v1"), "/feature"); code != http.StatusNotFound {
		t.Fatalf("feature disabled, got %d", code)
	}
	if code, body := get(t, NewHandler(ready, true, "v1"), "/feature"); code != http.StatusOK || body != "Feature X is enabled!" {
		t.Fatalf("feature enabled, got %d %q", code, body)
	}
}

func TestVersionIsReportedInBodyAndHeader(t *testing.T) {
	h := NewHandler(ready, false, "v2")

	if code, body := get(t, h, "/version"); code != http.StatusOK || body != "v2" {
		t.Fatalf("got %d %q", code, body)
	}
	if got := serve(h, "/ping").Header().Get("X-App-Version"); got != "v2" {
		t.Fatalf("X-App-Version = %q", got)
	}
}
