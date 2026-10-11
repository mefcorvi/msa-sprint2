// Package httpapi serves the service's HTTP endpoints for Kubernetes probes and the feature flag.
package httpapi

import (
	"context"
	"net/http"
)

// ReadinessCheck reports whether the service can serve traffic (database reachable and migrated).
type ReadinessCheck func(ctx context.Context) error

// NewHandler: /ping and /health answer while the process is alive, /ready also checks dependencies,
// /feature exists only when feature X is enabled. Every response carries X-App-Version, so canary
// routing can be observed from outside.
func NewHandler(ready ReadinessCheck, featureX bool, version string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("pong"))
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if err := ready(r.Context()); err != nil {
			http.Error(w, "not ready: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ready"))
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(version))
	})
	if featureX {
		mux.HandleFunc("GET /feature", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("Feature X is enabled!"))
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-App-Version", version)
		mux.ServeHTTP(w, r)
	})
}
