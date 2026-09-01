// Package config reads and validates the service configuration from the
// environment at startup. It never returns the value of a secret (such as the
// database password embedded in DATABASE_URL) in an error: a validation failure
// names the offending variable, it does not echo its value.
package config

import (
	"errors"
	"os"
)

// DefaultHTTPAddr is the listen address used when HTTP_ADDR is unset.
const DefaultHTTPAddr = ":8080"

// Config is the immutable, validated configuration read once at startup.
type Config struct {
	// DatabaseURL is the Postgres DSN. Required.
	DatabaseURL string
	// HTTPAddr is the TCP listen address. Defaults to DefaultHTTPAddr.
	HTTPAddr string
}

// ErrMissingDatabaseURL is returned when DATABASE_URL is absent or blank. Its
// text names the variable only and never contains a value.
var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required but was empty or unset")

// Load reads the configuration from the process environment and validates it.
// A missing or blank DATABASE_URL is a fatal configuration error. The returned
// error is safe to log: it never contains the value of any environment variable.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultHTTPAddr
	}
	if cfg.DatabaseURL == "" {
		return Config{}, ErrMissingDatabaseURL
	}
	return cfg, nil
}
