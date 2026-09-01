package store

import (
	"strings"
	"testing"
)

// A recognisable password token used to prove it is never surfaced in a reason.
const secretToken = "sup3rSecret"

func TestSafeReasonScrubsPasswordFromDriverError(t *testing.T) {
	dsn := "postgres://user:" + secretToken + "@db:5432/app?sslmode=disable"
	// Simulate a wrapped driver error that embedded the whole DSN.
	drvErr := errString("dial error connecting with " + dsn)
	reason := SafeReason(drvErr, dsn)
	if strings.Contains(reason, secretToken) {
		t.Fatalf("reason leaked the password token: %q", reason)
	}
	if strings.Contains(reason, dsn) {
		t.Fatalf("reason leaked the raw DSN: %q", reason)
	}
}

func TestSafeReasonParseErrorHasNoValue(t *testing.T) {
	dsn := "not://a valid dsn:" + secretToken + "@"
	reason := SafeReason(parseError{}, dsn)
	if strings.Contains(reason, secretToken) {
		t.Fatalf("parse-error reason leaked the token: %q", reason)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
