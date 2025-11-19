package model

import "testing"

func TestIsValidSection(t *testing.T) {
	tests := []struct {
		section string
		want    bool
	}{
		{"A", true},
		{"B", true},
		{"a", false},
		{"C", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsValidSection(tt.section)
		if got != tt.want {
			t.Errorf("IsValidSection(%q) = %v, want %v", tt.section, got, tt.want)
		}
	}
}

func TestIsValidSeatNumber(t *testing.T) {
	tests := []struct {
		seatNumber int32
		want       bool
	}{
		{1, true},
		{5, true},
		{10, true},
		{0, false},
		{11, false},
		{-1, false},
	}

	for _, tt := range tests {
		got := IsValidSeatNumber(tt.seatNumber)
		if got != tt.want {
			t.Errorf("IsValidSeatNumber(%d) = %v, want %v", tt.seatNumber, got, tt.want)
		}
	}
}

