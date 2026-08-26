package exercises

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExerciseDemoMediaDTOIsPublicAndSerializable(t *testing.T) {
	exercise := Exercise{
		ID:   "30000000-0000-0000-0000-000000000001",
		Name: "Push-up",
		DemoMedia: &DemoMedia{
			URL:       "https://cdn.example.com/exercises/push-up/demo.mp4",
			Type:      "video",
			MIMEType:  "video/mp4",
			PosterURL: "https://cdn.example.com/exercises/push-up/poster.webp",
		},
	}
	raw, err := json.Marshal(exercise)
	if err != nil {
		t.Fatal(err)
	}
	value := string(raw)
	for _, field := range []string{`"demo_media"`, `"mime_type":"video/mp4"`, `"poster_url"`} {
		if !strings.Contains(value, field) {
			t.Fatalf("missing public demo field %s in %s", field, value)
		}
	}
	for _, secret := range []string{"storage_key", "storage_provider", "credentials"} {
		if strings.Contains(value, secret) {
			t.Fatalf("internal media field leaked: %s", secret)
		}
	}
}
