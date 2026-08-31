package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

func signedTenantInitData(token, start string, telegramID int64) string {
	values := url.Values{
		"auth_date": {strconv.FormatInt(time.Now().Unix(), 10)},
		"query_id":  {"phase19a"},
		"user":      {`{"id":` + strconv.FormatInt(telegramID, 10) + `,"first_name":"Tenant E2E"}`},
	}
	if start != "" {
		values.Set("start_param", start)
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values.Get(key))
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secret.Write([]byte(token))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	_, _ = mac.Write([]byte(strings.Join(parts, "\n")))
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return values.Encode()
}

func TestSignedTelegramTenantBootstrapPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	const ownerA = "95000000-0000-0000-0000-000000000001"
	const ownerB = "95000000-0000-0000-0000-000000000002"
	const tenantA = "95000000-0000-0000-0000-000000000011"
	const tenantB = "95000000-0000-0000-0000-000000000012"
	const tenantOff = "95000000-0000-0000-0000-000000000013"
	const telegramID int64 = 950000099
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id IN ($1,$2,$3)`, tenantA, tenantB, tenantOff)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id IN ($1,$2,$3)`, tenantA, tenantB, tenantOff)
		_, _ = pool.Exec(ctx, `DELETE FROM user_progress WHERE user_id IN (SELECT id FROM users WHERE telegram_id=$1)`, telegramID)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN (SELECT id FROM users WHERE telegram_id=$1) OR user_id IN ($2,$3)`, telegramID, ownerA, ownerB)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE telegram_id=$1 OR id IN ($2,$3)`, telegramID, ownerA, ownerB)
	}()
	for index, id := range []string{ownerA, ownerB} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,'Owner')`, id, 950000001+index); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Owner')`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id,status) VALUES($1,'tenant-a','Tenant A',$4,'active'),($2,'tenant-b','Tenant B',$5,'active'),($3,'tenant-off','Tenant Off',$4,'suspended')`, tenantA, tenantB, tenantOff, ownerA, ownerB); err != nil {
		t.Fatal(err)
	}

	const botToken = "phase19a-test-token"
	handler := telegramAuthHandler(AuthDependencies{BotToken: botToken, JWTSecret: "phase19a-jwt-secret-long-enough", Users: users.NewStore(pool)})
	request := func(initData, suffix string) (int, map[string]any) {
		body, _ := json.Marshal(map[string]string{"init_data": initData})
		req := httptest.NewRequest(http.MethodPost, "/auth/telegram"+suffix, bytes.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		var response map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		return rec.Code, response
	}

	for _, slug := range []string{"tenant-a", "tenant-a", "tenant-b"} {
		code, response := request(signedTenantInitData(botToken, slug, telegramID), "")
		if code != http.StatusOK {
			t.Fatalf("signed %s: code=%d body=%v", slug, code, response)
		}
		user := response["user"].(map[string]any)
		current := user["current_tenant"].(map[string]any)
		if current["slug"] != slug {
			t.Fatalf("current tenant=%v want=%s", current, slug)
		}
	}
	var memberships int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM tenant_memberships m JOIN users u ON u.id=m.user_id WHERE u.telegram_id=$1`, telegramID).Scan(&memberships); err != nil || memberships != 2 {
		t.Fatalf("memberships=%d err=%v", memberships, err)
	}

	code, _ := request(signedTenantInitData(botToken, "", telegramID), "?startapp=tenant-a")
	if code != http.StatusOK {
		t.Fatalf("unsigned frontend query must not break auth: %d", code)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id JOIN users u ON u.id=m.user_id WHERE u.telegram_id=$1 AND t.slug='tenant-a'`, telegramID).Scan(&memberships); err != nil || memberships != 1 {
		t.Fatalf("query bootstrap changed membership: %d %v", memberships, err)
	}
	for _, slug := range []string{"missing-tenant", "tenant-off"} {
		code, _ = request(signedTenantInitData(botToken, slug, telegramID), "")
		if code != http.StatusNotFound {
			t.Fatalf("%s code=%d want=404", slug, code)
		}
	}
	code, _ = request(signedTenantInitData("wrong-token", "tenant-a", telegramID), "")
	if code != http.StatusUnauthorized {
		t.Fatalf("invalid signature code=%d want=401", code)
	}
}
