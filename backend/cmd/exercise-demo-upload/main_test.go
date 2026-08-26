package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadFlagsRequireExplicitMode(t *testing.T) {
	if _, err := parse([]string{"--standard-key", "push-up", "--file", "demo.mp4"}); err == nil {
		t.Fatal("implicit write mode accepted")
	}
	if _, err := parse([]string{"--standard-key", "push-up", "--file", "demo.mp4", "--dry-run", "--confirm"}); err == nil {
		t.Fatal("multiple modes accepted")
	}
	if _, err := parse([]string{"--standard-key", "Push Up", "--file", "demo.mp4", "--dry-run"}); err == nil {
		t.Fatal("invalid standard key accepted")
	}
}

func TestStableSystemStorageKey(t *testing.T) {
	if got := storageKey("push-up-standard", "video/mp4"); got != "exercises/standard/push-up-standard/demo.mp4" {
		t.Fatalf("storage key = %q", got)
	}
}

func TestInspectFileValidatesTypeSizeAndDuration(t *testing.T) {
	directory := t.TempDir()
	gif := filepath.Join(directory, "demo.gif")
	if err := os.WriteFile(gif, []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := inspectFile(gif, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = info.File.Close()
	if info.Type != "image" || info.MIMEType != "image/gif" {
		t.Fatalf("unexpected media inspection: %#v", info)
	}
	if _, err = inspectFile(gif, 3); err == nil {
		t.Fatal("image duration accepted")
	}
	bad := filepath.Join(directory, "demo.mp4")
	if err = os.WriteFile(bad, []byte("not a real mp4"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = inspectFile(bad, 3); err == nil {
		t.Fatal("extension-only fake video accepted")
	}
}
