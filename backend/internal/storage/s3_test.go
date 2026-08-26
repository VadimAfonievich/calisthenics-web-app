package storage

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestS3ConfigurationAndStablePublicURL(t *testing.T) {
	_, err := NewS3(S3Config{})
	if err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("incomplete config accepted: %v", err)
	}
	provider, err := NewS3(S3Config{
		Endpoint:  "https://s3.example.com",
		Region:    "auto",
		Bucket:    "demo",
		AccessKey: "access",
		SecretKey: "secret",
		PublicURL: "https://cdn.example.com/media/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := provider.PublicURL("exercises/standard/push-up/demo.mp4"); got != "https://cdn.example.com/media/exercises/standard/push-up/demo.mp4" {
		t.Fatalf("public URL = %q", got)
	}
}

func TestS3UploadAndDeleteUseSignedStableRequests(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		if request.URL.Path != "/demo-bucket/exercises/standard/push-up/demo.mp4" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if !strings.HasPrefix(request.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=access/") {
			t.Errorf("missing SigV4 authorization")
		}
		if request.Method == http.MethodPut && request.Header.Get("Cache-Control") != "public, max-age=31536000, immutable" {
			t.Errorf("cache control = %q", request.Header.Get("Cache-Control"))
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	provider, err := NewS3(S3Config{Endpoint: server.URL, Region: "auto", Bucket: "demo-bucket", AccessKey: "access", SecretKey: "secret", PublicURL: "https://cdn.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	provider.now = func() time.Time { return time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC) }
	object, err := provider.Upload(context.Background(), UploadInput{Key: "exercises/standard/push-up/demo.mp4", ContentType: "video/mp4", Size: 4, Body: bytes.NewReader([]byte("demo"))})
	if err != nil {
		t.Fatal(err)
	}
	if object.URL != "https://cdn.example.com/exercises/standard/push-up/demo.mp4" {
		t.Fatalf("URL = %q", object.URL)
	}
	if err = provider.Delete(context.Background(), object.Key); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d", requests)
	}
}

func TestS3RejectsUnsafeConfiguration(t *testing.T) {
	base := S3Config{Endpoint: "https://s3.example.com/path", Bucket: "demo", AccessKey: "access", SecretKey: "secret", PublicURL: "https://cdn.example.com"}
	if _, err := NewS3(base); err == nil {
		t.Fatal("endpoint path accepted")
	}
	base.Endpoint, base.PublicURL = "https://s3.example.com", "http://cdn.example.com"
	if _, err := NewS3(base); err == nil {
		t.Fatal("insecure public URL accepted")
	}
}
