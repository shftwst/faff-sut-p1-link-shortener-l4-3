package config

import (
	"strings"
	"testing"
)

func TestLoadMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_ADDR", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected an error when DATABASE_URL is unset, got nil")
	}
	// The error must name the variable but never echo a value.
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("error should name DATABASE_URL, got %q", err.Error())
	}
}

func TestLoadDefaultsHTTPAddr(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@db:5432/x?sslmode=disable")
	t.Setenv("HTTP_ADDR", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPAddr != DefaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, DefaultHTTPAddr)
	}
}

func TestLoadHonoursHTTPAddr(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@db:5432/x?sslmode=disable")
	t.Setenv("HTTP_ADDR", ":9090")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL should be populated")
	}
}
