package main

import "testing"

func TestManifestValidation(t *testing.T) {
	good := manifest{Version: 1, Mappings: []mapping{{"push-up-standard", "10000000-0000-0000-0000-000000000001"}}}
	if validate(good) != nil {
		t.Fatal("valid rejected")
	}
	good.Mappings = append(good.Mappings, good.Mappings[0])
	if validate(good) == nil {
		t.Fatal("duplicate accepted")
	}
}

func TestManifestValidationRejectsMalformedUUIDAndKey(t *testing.T) {
	for _, item := range []mapping{
		{StandardKey: "Push Up", MediaAssetID: "10000000-0000-0000-0000-000000000001"},
		{StandardKey: "push-up-standard", MediaAssetID: "10000000-0000-0000-0000-00000000000-"},
	} {
		if validate(manifest{Version: 1, Mappings: []mapping{item}}) == nil {
			t.Fatalf("invalid mapping accepted: %#v", item)
		}
	}
}

func TestManifestValidationAllowsPendingProductionAssets(t *testing.T) {
	pending := manifest{Version: 1, Mappings: []mapping{{StandardKey: "push-up-standard"}}}
	if err := validate(pending); err != nil {
		t.Fatalf("pending production manifest rejected: %v", err)
	}
}
