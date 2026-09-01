// Command api is the link-shortener service entrypoint. Startup is a strict,
// fail-fast sequence: load config, connect to Postgres and Ping, run migrations,
// then bind and serve. The HTTP port is never bound unless every prior step
// succeeds, so a healthy api implies a reachable database.
package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"linkshortener/internal/config"
	"linkshortener/internal/httpapi"
	"linkshortener/internal/migrate"
	"linkshortener/internal/store"
)

// migrationsDir is the golang-migrate source directory, resolved relative to the
// working directory (the image WORKDIR is /, so this is /migrations).
const migrationsDir = "migrations"

// shutdownTimeout bounds the graceful drain on SIGINT/SIGTERM.
const shutdownTimeout = 10 * time.Second

func main() {
	os.Exit(run())
}

// run performs the startup sequence and blocks serving until a shutdown signal.
// It returns the process exit code so deferred cleanup still runs (os.Exit in
// main would skip defers).
func run() int {
	// 1. Configuration. A missing/blank DATABASE_URL is fatal before any port is
	// bound. The error names the variable and never echoes its value.
	cfg, err := config.Load()
	if err != nil {
		logFatal("config_invalid", err.Error())
		return 1
	}

	ctx := context.Background()

	// 2. Database connection + Ping. On failure the port is never bound. The
	// reason is scrubbed of the DSN so the password cannot reach the logs.
	pool, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logFatal("db_connect_failed", store.SafeReason(err, cfg.DatabaseURL))
		return 1
	}
	defer pool.Close()

	// 3. Migrations. ErrNoChange (E1's empty migrations dir) is success. Any
	// other error is fatal before binding; the reason is DSN-scrubbed.
	status, err := migrate.Run(cfg.DatabaseURL, migrationsDir)
	if err != nil {
		logFatal("migrate_failed", migrate.SafeReason(err, cfg.DatabaseURL))
		return 1
	}
	logInfo("migrate_done", "migration runner ran: "+status)

	// 4/5. Bind and serve. The listener is opened only now — after config, the
	// DB connection, and the migration runner all succeeded — so the port is
	// never bound in a half-up state. The "listening" line is logged strictly
	// after the migration line above.
	mux := httpapi.NewRouter(pool)
	srv := &http.Server{Handler: mux}

	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		logFatal("listen_failed", "could not bind "+cfg.HTTPAddr)
		return 1
	}
	logInfo("listening", "http server listening on "+cfg.HTTPAddr)

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		logFatal("serve_failed", "http server stopped serving: "+err.Error())
		return 1
	case <-stop:
		// Graceful shutdown: stop accepting new connections and drain in-flight
		// requests within the bounded timeout, then exit 0.
		logInfo("shutdown", "signal received, draining connections")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logInfo("shutdown_forced", "drain did not complete within timeout")
		}
		return 0
	}
}

// logFatal writes a single-line JSON fatal record to stderr. reason must never
// contain a secret; callers pass DSN-scrubbed strings.
func logFatal(event, reason string) {
	writeLine(os.Stderr, map[string]string{"level": "fatal", "event": event, "reason": reason})
}

// logInfo writes a single-line JSON info record to stdout.
func logInfo(event, msg string) {
	writeLine(os.Stdout, map[string]string{"level": "info", "event": event, "msg": msg})
}

func writeLine(w *os.File, rec map[string]string) {
	b, err := json.Marshal(rec)
	if err != nil {
		return
	}
	_, _ = w.Write(append(b, '\n'))
}
