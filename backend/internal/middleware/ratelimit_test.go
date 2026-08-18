package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLimitBodyRejectsOversizedPayload(t *testing.T) {
	next := LimitBody(3, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, e := io.ReadAll(r.Body)
		if e == nil {
			t.Fatal("expected body limit error")
		}
		w.WriteHeader(204)
	}))
	w := httptest.NewRecorder()
	next.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("1234")))
	if w.Code != 204 {
		t.Fatalf("got %d", w.Code)
	}
}
