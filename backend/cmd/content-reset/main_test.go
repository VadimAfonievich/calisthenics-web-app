package main

import (
	"strings"
	"testing"
)

func TestResetStatementsAreOwnerScopedAndNonDeleting(t *testing.T) {
	for _, q := range resetStatements {
		upper := strings.ToUpper(q)
		if !strings.Contains(q, "owner_user_id IS NULL") {
			t.Fatalf("not system scoped: %s", q)
		}
		if strings.Contains(upper, "DELETE") || strings.Contains(upper, "TRUNCATE") {
			t.Fatalf("destructive statement: %s", q)
		}
	}
}
