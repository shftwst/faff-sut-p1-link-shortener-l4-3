// Package store owns the Postgres connection foundation: it constructs a
// pgxpool and confirms connectivity with a single Ping at startup. It exposes
// no queries yet; the links repository lands in a later epic.
//
// Every error this package surfaces is scrubbed of the DSN. A raw pgx parse or
// connection error frequently embeds the connection string (host, user, and
// password), so callers must log SafeReason(err, dsn), never err.Error()
// verbatim.
package store

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectTimeout bounds the parse+connect+ping sequence at startup.
const connectTimeout = 10 * time.Second

// parseError marks a failure to parse the DSN. The underlying pgx parse error is
// deliberately discarded because it can echo the DSN verbatim.
type parseError struct{}

func (parseError) Error() string { return "DATABASE_URL is malformed and could not be parsed" }

// Open parses the DSN, constructs a pgxpool, and Pings once to confirm the
// database is reachable. On any failure it returns an error that is safe to pass
// to SafeReason for logging.
func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, parseError{}
	}

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// SafeReason returns a stable, secret-free reason string for a store error. It
// never contains the DSN or its password. When the error is not a known parse
// failure it redacts any DSN or password substring from the driver error text
// before returning it, so a wrapped pgx error can never leak the connection
// string into the logs.
func SafeReason(err error, dsn string) string {
	if err == nil {
		return ""
	}
	if _, ok := err.(parseError); ok {
		return err.Error()
	}
	return "could not connect to the database named by DATABASE_URL: " + scrub(err.Error(), dsn)
}

// scrub removes the DSN and its password from a message. It replaces the whole
// DSN string first, then the parsed password (in case the driver reconstructed
// or partially quoted the connection string).
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
