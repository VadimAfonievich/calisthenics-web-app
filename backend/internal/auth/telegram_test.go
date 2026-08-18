package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func signedInitData(t *testing.T, botToken string, authDate int64) string {
	t.Helper()
	values := url.Values{"auth_date": {""}, "query_id": {"query"}, "user": {`{"id":42,"first_name":"Анна","username":"anna"}`}}
	values.Set("auth_date", strconv.FormatInt(authDate, 10))
	data := "auth_date=" + values.Get("auth_date") + "\nquery_id=query\nuser=" + values.Get("user")
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secret.Write([]byte(botToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	_, _ = mac.Write([]byte(data))
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return values.Encode()
}

func TestValidateInitData(t *testing.T) {
	now := time.Unix(1_700_000_100, 0)
	user, err := ValidateInitData(signedInitData(t, "bot-token", now.Add(-time.Minute).Unix()), "bot-token", now, 24*time.Hour)
	if err != nil || user.ID != 42 || user.FirstName != "Анна" {
		t.Fatalf("user = %#v, err = %v", user, err)
	}
}

func TestValidateInitDataRejectsExpiredAndInvalidSignature(t *testing.T) {
	now := time.Unix(1_700_000_100, 0)
	if _, err := ValidateInitData(signedInitData(t, "bot-token", now.Add(-25*time.Hour).Unix()), "bot-token", now, 24*time.Hour); err != ErrExpiredInitData {
		t.Fatalf("err = %v", err)
	}
	if _, err := ValidateInitData(signedInitData(t, "bot-token", now.Unix()), "different-token", now, 24*time.Hour); err != ErrInvalidInitData {
		t.Fatalf("err = %v", err)
	}
}
