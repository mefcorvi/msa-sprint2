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
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code, rec.Body.String()
}

func ready(context.Context) error    { return nil }
func notReady(context.Context) error { return errors.New("database unavailable") }

func TestPingAnswersPong(t *testing.T) {
	code, body := get(t, NewHandler(notReady, false), "/ping")

	if code != http.StatusOK || body != "pong" {
		t.Fatalf("got %d %q", code, body)
	}
}

func TestReadyReflectsDependencies(t *testing.T) {
	if code, _ := get(t, NewHandler(ready, false), "/ready"); code != http.StatusOK {
		t.Fatalf("ready service answered %d", code)
	}
	if code, _ := get(t, NewHandler(notReady, false), "/ready"); code != http.StatusServiceUnavailable {
		t.Fatalf("not ready service answered %d", code)
	}
}

func TestFeatureEndpointFollowsFlag(t *testing.T) {
	if code, _ := get(t, NewHandler(ready, false), "/feature"); code != http.StatusNotFound {
		t.Fatalf("feature disabled, got %d", code)
	}
	if code, body := get(t, NewHandler(ready, true), "/feature"); code != http.StatusOK || body != "Feature X is enabled!" {
		t.Fatalf("feature enabled, got %d %q", code, body)
	}
}
