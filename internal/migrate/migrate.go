// Package migrate wraps the golang-migrate runner the api invokes at startup.
// It applies all pending migrations to the latest version before the server
// binds. E1 ships an empty migrations directory, so the runner returns
// ErrNoChange, which is treated as success.
//
// Errors are scrubbed of the DSN: golang-migrate wraps the underlying driver
// error, which can embed the connection string, so callers must log
// SafeReason(err, dsn) rather than err.Error() verbatim.
package migrate

import (
	"errors"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	// Postgres database driver and file source, registered via blank import.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Run applies all pending migrations found under dir against the database named
// by dsn. It returns a short human-readable status ("no change" or "applied")
// and an error. ErrNoChange is not an error: an empty migrations directory
// yields "no change".
func Run(dsn, dir string) (status string, err error) {
	m, err := migrate.New("file://"+dir, dsn)
	if err != nil {
		return "", err
	}
	defer m.Close()

	err = m.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		return "no change", nil
	case err != nil:
		return "", err
	default:
		return "applied", nil
	}
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
			out = strings.ReplaceAll(out, pw, "[redacted]")
		}
	}
	return out
}
