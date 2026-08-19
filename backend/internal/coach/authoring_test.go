package coach

import "testing"

func TestSlugBaseHumanTitles(t *testing.T) {
	tests := map[string]string{"Как научиться стоять на руках": "kak-nauchitsya-stoyat-na-rukah", "Handstand Basics": "handstand-basics", "***": "content"}
	for in, want := range tests {
		if got := slugBase(in); got != want {
			t.Errorf("slugBase(%q)=%q want %q", in, got, want)
		}
	}
}

func TestCoachOwnershipScope(t *testing.T) {
	where, args := scope("coach", "10000000-0000-0000-0000-000000000001")
	if where == "" || len(args) != 1 {
		t.Fatal("coach content must stay owner-scoped")
	}
	where, args = scope("admin", "x")
	if where != "" || len(args) != 0 {
		t.Fatal("admin should retain cross-owner scope")
	}
}

func TestCategoryOptionsQueryHasNoUnexpectedBindArguments(t *testing.T) {
	if got := optionArgs(0, "coach", "user-id"); len(got) != 0 {
		t.Fatalf("category query args=%v", got)
	}
	if got := optionArgs(1, "coach", "user-id"); len(got) != 2 {
		t.Fatalf("owner-scoped option args=%v", got)
	}
}
