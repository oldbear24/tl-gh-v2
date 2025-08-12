package main

import "testing"

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
