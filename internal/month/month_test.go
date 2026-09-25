package month

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	// 2026-10-01 00:30 JST は UTC ではまだ 9月
	now := time.Date(2026, 9, 30, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		in   string
		want string
	}{
		{"2026-09", "2026-09"},
		{"", "2026-10"},
		{"current", "2026-10"},
		{"previous", "2026-09"},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in, now)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.in, err)
		}
		if got.String() != tt.want {
			t.Errorf("Parse(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}

	if _, err := Parse("2026/09", now); err == nil {
		t.Error("Parse(\"2026/09\") should fail")
	}
}

func TestAdd(t *testing.T) {
	m := Month{Year: 2026, Month: time.January}
	if got := m.Add(-1).String(); got != "2025-12" {
		t.Errorf("Add(-1) = %s", got)
	}
	if got := m.Add(12).String(); got != "2027-01" {
		t.Errorf("Add(12) = %s", got)
	}
}
