// Package httpapi wires the HTTP surface for the service: the /healthz liveness
// handler and a structured JSON error envelope used by the 404 and 405
// fallbacks. Later epics register the product routes on this same router.
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrorResponse is the structured error envelope every non-2xx response the
// service originates uses. The message is human-readable and safe to surface: it
// carries no internal detail or secrets.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the inner payload of ErrorResponse.
type ErrorBody struct {
	// Code is a stable, machine-readable token, e.g. "not_found".
	Code string `json:"code"`
	// Message is human-readable and safe to surface.
	Message string `json:"message"`
}

// NewRouter builds the service router. The pool is held for the health surface
// and for the repositories later epics attach; E1's handlers do not query it.
func NewRouter(pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	// /healthz is defined for GET only. Registering the bare path (rather than a
	// method-scoped pattern) lets us return the structured 405 envelope for a
	// wrong method instead of the ServeMux default plain-text body.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		healthz(w, r)
	})

	// Catch-all: any path not matched above is a structured 404.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	})

	return mux
}

// healthz is the shallow liveness handler. Reaching it already implies config,
// the database connection, and migrations succeeded at boot.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
}

// writeError emits the structured JSON error envelope with the given status.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}
