package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMissingDirIsError(t *testing.T) {
	// A missing migrations directory is a real misconfiguration and must error
	// before any database contact.
	_, err := Run("postgres://u:p@localhost:1/none?sslmode=disable", filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected an error for a missing migrations directory")
	}
	if !strings.Contains(err.Error(), "not present") {
		t.Fatalf("error should mention the directory is not present, got %q", err.Error())
	}
}

func TestNoMigrationFiles(t *testing.T) {
	empty := t.TempDir()
	if !noMigrationFiles(empty) {
		t.Fatal("empty dir should report no migration files")
	}

	withSQL := t.TempDir()
	if err := os.WriteFile(filepath.Join(withSQL, "0001_init.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatal(err)
	}
	if noMigrationFiles(withSQL) {
		t.Fatal("dir with a .sql file should report migration files present")
	}
}

func TestSafeReasonScrubsDSN(t *testing.T) {
	dsn := "postgres://user:sup3rSecret@db:5432/app?sslmode=disable"
	reason := SafeReason(errString("failed: "+dsn), dsn)
	if strings.Contains(reason, "sup3rSecret") || strings.Contains(reason, dsn) {
		t.Fatalf("reason leaked the DSN/password: %q", reason)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
