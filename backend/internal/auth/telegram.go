package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidInitData = errors.New("invalid Telegram initData")
	ErrExpiredInitData = errors.New("expired Telegram initData")
)

type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhotoURL  string `json:"photo_url"`
}

// ValidateInitData verifies Telegram's Mini App signature and returns the
// identity embedded in the signed user field.
func ValidateInitData(raw, botToken string, now time.Time, maxAge time.Duration) (TelegramUser, error) {
	if raw == "" || botToken == "" {
		return TelegramUser{}, ErrInvalidInitData
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	hash := values.Get("hash")
	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if hash == "" || err != nil || authDate <= 0 {
		return TelegramUser{}, ErrInvalidInitData
	}
	issuedAt := time.Unix(authDate, 0)
	if issuedAt.After(now.Add(5*time.Minute)) || now.Sub(issuedAt) > maxAge {
		return TelegramUser{}, ErrExpiredInitData
	}

	pairs := make([]string, 0, len(values)-1)
	for key, entries := range values {
		if key == "hash" {
			continue
		}
		for _, value := range entries {
			pairs = append(pairs, key+"="+value)
		}
	}
	sort.Strings(pairs)
	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secretMAC.Write([]byte(botToken))
	checkMAC := hmac.New(sha256.New, secretMAC.Sum(nil))
	_, _ = checkMAC.Write([]byte(strings.Join(pairs, "\n")))
	expected, err := hex.DecodeString(hash)
	if err != nil || !hmac.Equal(expected, checkMAC.Sum(nil)) {
		return TelegramUser{}, ErrInvalidInitData
	}

	var user TelegramUser
	if err := json.Unmarshal([]byte(values.Get("user")), &user); err != nil || user.ID <= 0 || strings.TrimSpace(user.FirstName) == "" {
		return TelegramUser{}, fmt.Errorf("%w: user", ErrInvalidInitData)
	}
	return user, nil
}
