package auth

import (
	"testing"
	"time"
)

func TestIssueAndVerifyToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	token, err := IssueToken("11111111-1111-1111-1111-111111111111", "secret", now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := VerifyToken(token, "secret", now.Add(time.Minute))
	if err != nil || claims.Subject == "" {
		t.Fatalf("claims = %#v, err = %v", claims, err)
	}
}
