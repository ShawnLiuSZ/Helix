package ui

import "testing"

func TestParseRemember(t *testing.T) {
	cases := []struct {
		name     string
		arg      string
		existing int
		wantKey  string
		wantVal  string
		wantOK   bool
	}{
		{"key colon value", "test cmd: make test", 0, "test cmd", "make test", true},
		{"plain note gets auto key", "always run gofmt", 2, "note-3", "always run gofmt", true},
		{"empty is rejected", "   ", 0, "", "", false},
		{"colon with empty key falls back to note", ": orphan", 0, "note-1", ": orphan", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, val, ok := parseRemember(tc.arg, tc.existing)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if key != tc.wantKey || val != tc.wantVal {
				t.Errorf("got (%q, %q), want (%q, %q)", key, val, tc.wantKey, tc.wantVal)
			}
		})
	}
}
