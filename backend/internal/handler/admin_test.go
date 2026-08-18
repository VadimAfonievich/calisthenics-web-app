package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type adminAuthorizerStub struct {
	allowed bool
	err     error
}

func (s adminAuthorizerStub) IsAdmin(context.Context, string) (bool, error) { return s.allowed, s.err }
func TestRequireAdminRejectsUnauthenticated(t *testing.T) {
	next := requireAdmin(adminAuthorizerStub{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	w := httptest.NewRecorder()
	next.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/lessons", nil))
	if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), "UNAUTHORIZED") {
		t.Fatalf("got %d: %s", w.Code, w.Body.String())
	}
}
