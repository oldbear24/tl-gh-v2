package main

import (
	"testing"
	"time"
)

func TestParseDateTimeValid(t *testing.T) {
	dateStr := "27-05-2024"
	timeStr := "14:30"

	got, err := parseDateTime(dateStr, timeStr)
	if err != nil {
		t.Fatalf("parseDateTime returned error: %v", err)
	}

	loc, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	want := time.Date(2024, time.May, 27, 14, 30, 0, 0, loc)
	if !got.Equal(want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestParseDateTimeInvalid(t *testing.T) {
	_, err := parseDateTime("2024/05/27", "14:30")
	if err == nil {
    t.Fatalf("expected error for invalid date format, got nil")
  }
}
func TestRollDiceRange(t *testing.T) {
	saw100 := false
	for i := 0; i < 10000; i++ {
		n := rollDice()
		if n < 1 || n > 100 {
			t.Fatalf("rollDice returned %d, want between 1 and 100", n)
		}
		if n == 100 {
			saw100 = true
			break
		}
	}
	if !saw100 {
		t.Fatalf("rollDice never returned 100 in 10000 attempts")
	}
}
