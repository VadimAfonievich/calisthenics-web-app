package main

import "testing"

func TestValidationPrimitives(t *testing.T) {
	if !keyPattern.MatchString("push-up-standard") || keyPattern.MatchString("Push Up") {
		t.Fatal("key validation")
	}
	if !duplicates([]string{"push", "push"}) || duplicates([]string{"push", "core"}) {
		t.Fatal("duplicate validation")
	}
	if !https("https://example.com/a.mp4") || https("http://example.com/a.mp4") {
		t.Fatal("media validation")
	}
}
