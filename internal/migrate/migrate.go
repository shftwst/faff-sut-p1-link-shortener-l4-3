// Package migrate wraps the golang-migrate runner the api invokes at startup.
// It applies all pending migrations to the latest version before the server
// binds. E1 ships an empty migrations directory, so the runner finds nothing to
// apply, which is treated as success.
//
// Errors are scrubbed of the DSN: golang-migrate wraps the underlying driver
// error, which can embed the connection string, so callers must log
// SafeReason(err, dsn) rather than err.Error() verbatim.
package migrate

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	// Postgres database driver and file source, registered via blank import.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Run applies all pending migrations found under dir against the database named
// by dsn. It returns a short human-readable status ("applied" or a "no change"
// variant) and an error.
//
// Two non-error cases both mean "nothing to apply": golang-migrate's ErrNoChange
// (migration files exist but none are pending) and a present-but-empty
// migrations directory (E1 ships no migration files, for which the file source
// reports "file does not exist"). A genuinely missing directory, or a real
// migration failure when migration files are present, is a hard error.
func Run(dsn, dir string) (status string, err error) {
	// A missing migrations directory is a real misconfiguration; only a present
	// but empty directory is E1's expected no-op.
	if fi, statErr := os.Stat(dir); statErr != nil || !fi.IsDir() {
		return "", fmt.Errorf("migrations directory %q is not present", dir)
	}

	m, err := migrate.New("file://"+dir, dsn)
	if err != nil {
		// A genuine runner-construction failure (bad DSN, driver init, DB
		// unreachable) must surface — never mask it just because the dir is
		// empty.
		return "", err
	}
	defer m.Close()

	// Empty source is a deterministic no-op: the runner is constructed and has
	// nothing to apply. Short-circuit here rather than calling Up() and trying to
	// tell its "no migration files" error apart from a real failure once
	// golang-migrate has wrapped it — that ambiguity is exactly what would let a
	// real error be swallowed on E1's always-empty dir.
	if noMigrationFiles(dir) {
		return "no change (empty migrations dir)", nil
	}

	// A present, non-empty dir: apply. Only ErrNoChange is a no-op; every other
	// error is a genuine migration failure and surfaces.
	switch upErr := m.Up(); {
	case upErr == nil:
		return "applied", nil
	case errors.Is(upErr, migrate.ErrNoChange):
		return "no change", nil
	default:
		return "", upErr
	}
}

// noMigrationFiles reports whether dir contains no *.sql migration files.
func noMigrationFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			return false
		}
	}
	return true
}

// SafeReason returns a secret-free reason string for a migration error, with any
// DSN or password substring redacted from the wrapped driver error text.
func SafeReason(err error, dsn string) string {
	if err == nil {
		return ""
	}
	return "migration runner failed: " + scrub(err.Error(), dsn)
}

func scrub(msg, dsn string) string {
	if dsn == "" {
		return msg
	}
	out := strings.ReplaceAll(msg, dsn, "[redacted-dsn]")
	if u, err := url.Parse(dsn); err == nil && u.User != nil {
		if pw, ok := u.User.Password(); ok && pw != "" {
			// Redact the password and its common encoded forms, so a driver
			// error that reproduced it percent-encoded still cannot leak it.
			for _, form := range []string{pw, url.QueryEscape(pw), url.PathEscape(pw)} {
				out = strings.ReplaceAll(out, form, "[redacted]")
			}
		}
	}
	return out
}
