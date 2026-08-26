package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type catalog struct {
	Version   int        `json:"version"`
	Exercises []exercise `json:"exercises"`
}
type exercise struct {
	Key            string   `json:"key"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Instructions   string   `json:"instructions"`
	CommonMistakes string   `json:"common_mistakes"`
	CoachTips      string   `json:"coach_tips"`
	Difficulty     string   `json:"difficulty"`
	MovementType   string   `json:"movement_type"`
	MuscleGroups   []string `json:"muscle_groups"`
	Equipment      []string `json:"equipment"`
	Tags           []string `json:"tags"`
	Media          *struct {
		ImageURL string `json:"image_url"`
		VideoURL string `json:"video_url"`
	} `json:"media"`
}

var keyPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func allowed(v string, values ...string) bool {
	for _, x := range values {
		if v == x {
			return true
		}
	}
	return false
}
func duplicates(values []string) bool {
	seen := map[string]bool{}
	for _, v := range values {
		if seen[v] {
			return true
		}
		seen[v] = true
	}
	return false
}
func https(raw string) bool {
	if raw == "" {
		return true
	}
	u, e := url.Parse(raw)
	return e == nil && u.Scheme == "https" && u.Host != ""
}

func main() {
	file := flag.String("file", "", "catalog JSON file")
	validateOnly := flag.Bool("validate-only", false, "validate without database writes")
	flag.Parse()
	if !*validateOnly {
		fmt.Fprintln(os.Stderr, "refusing to import: only --validate-only is implemented")
		os.Exit(2)
	}
	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file is required")
		os.Exit(2)
	}
	raw, e := os.ReadFile(*file)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	var c catalog
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if e = decoder.Decode(&c); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	errs := []string{}
	if c.Version != 1 {
		errs = append(errs, "version must be 1")
	}
	keys := map[string]bool{}
	for i, x := range c.Exercises {
		p := fmt.Sprintf("exercises[%d]", i)
		if !keyPattern.MatchString(x.Key) {
			errs = append(errs, p+".key is invalid")
		} else if keys[x.Key] {
			errs = append(errs, p+".key is duplicated: "+x.Key)
		}
		keys[x.Key] = true
		if strings.TrimSpace(x.Name) == "" || strings.TrimSpace(x.Description) == "" || strings.TrimSpace(x.Instructions) == "" {
			errs = append(errs, p+" requires name, description and instructions")
		}
		if !allowed(x.Difficulty, "beginner", "intermediate", "advanced") {
			errs = append(errs, p+".difficulty is invalid")
		}
		if !allowed(x.MovementType, "reps", "duration", "distance", "custom") {
			errs = append(errs, p+".movement_type is invalid")
		}
		if len(x.MuscleGroups) == 0 || duplicates(x.MuscleGroups) {
			errs = append(errs, p+".muscle_groups must be non-empty and unique")
		}
		if duplicates(x.Equipment) || len(x.Tags) == 0 || duplicates(x.Tags) {
			errs = append(errs, p+" equipment/tags must be unique and tags non-empty")
		}
		for _, tag := range x.Tags {
			if !keyPattern.MatchString(tag) {
				errs = append(errs, p+".tags contains invalid value: "+tag)
			}
		}
		if x.Media != nil && (!https(x.Media.ImageURL) || !https(x.Media.VideoURL)) {
			errs = append(errs, p+".media URLs must use https")
		}
	}
	if len(errs) > 0 {
		for _, x := range errs {
			fmt.Fprintln(os.Stderr, x)
		}
		os.Exit(1)
	}
	fmt.Printf("VALID: %d standard exercises; no database changes applied\n", len(c.Exercises))
}
