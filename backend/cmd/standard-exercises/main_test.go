package main

import (
	"strings"
	"testing"
)

func TestValidationPrimitives(t *testing.T) {
	if !keyPattern.MatchString("push-up-standard") || keyPattern.MatchString("Push Up") {
		t.Fatal("key validation")
	}
	if !duplicates([]string{"push", "push"}) || duplicates([]string{"push", "core"}) {
		t.Fatal("duplicate validation")
	}
	if !validHTTPS("https://example.com/a.mp4") || validHTTPS("http://example.com/a.mp4") {
		t.Fatal("media validation")
	}
}

func validCatalogForTest() catalog {
	x := exercise{Key: "push-up", Name: "Push-up", Description: "Description", Instructions: "Instructions", Difficulty: "beginner", MovementType: "reps", MuscleGroups: []string{"chest"}, Equipment: []string{"floor"}, Tags: []string{"push"}}
	items := make([]exercise, 200)
	for i := range items {
		items[i] = x
		items[i].Key = "push-up-" + strings.Repeat("x", i+1)
	}
	return catalog{Version: 1, Exercises: items}
}

func TestCatalogRejectsDuplicateKey(t *testing.T) {
	c := validCatalogForTest()
	c.Exercises[1].Key = c.Exercises[0].Key
	if errs := validateCatalog(c); len(errs) == 0 {
		t.Fatal("duplicate key accepted")
	}
}

func TestCatalogRejectsInvalidEnums(t *testing.T) {
	c := validCatalogForTest()
	c.Exercises[0].Difficulty = "expert"
	c.Exercises[1].Tags = []string{"Invalid Tag"}
	if errs := validateCatalog(c); len(errs) < 2 {
		t.Fatalf("invalid vocabulary accepted: %v", errs)
	}
}

func TestPlanCreateUpdateIdempotencyAndCoachProtection(t *testing.T) {
	c := catalog{Exercises: []exercise{
		{Key: "new", Name: "New"}, {Key: "changed", Name: "Changed"}, {Key: "same", Name: "Same"}, {Key: "coach", Name: "Coach"},
	}}
	rows := []dbExercise{
		{ID: "2", Key: "changed", Slug: "changed", Name: "Old"},
		{ID: "3", Key: "same", Slug: "same", Name: "Same"},
		{ID: "4", Slug: "coach", Owner: "coach-id", Name: "Coach"},
	}
	// Fill the fields used by equality for the unchanged fixture.
	rows[1].Status = "published"
	p := buildPlan(c, rows)
	if p.Created != 1 || p.Updated != 1 || p.Unchanged != 1 || p.Conflicts != 1 {
		t.Fatalf("unexpected plan: %+v", p)
	}
}
