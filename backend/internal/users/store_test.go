package users

import (
	"reflect"
	"testing"
)

func TestAvailableModes(t *testing.T) {
	tests := []struct {
		role string
		want []string
	}{
		{"user", []string{"student"}},
		{"", []string{"student"}},
		{"coach", []string{"student", "coach"}},
		{"admin", []string{"student", "coach"}},
		{"super_admin", []string{"student", "coach", "admin"}},
	}
	for _, tt := range tests {
		if got := AvailableModes(tt.role); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("AvailableModes(%q)=%v, want %v", tt.role, got, tt.want)
		}
	}
}
