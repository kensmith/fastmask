package main

import "testing"

func TestGenPrefix(t *testing.T) {
	got1 := GenPrefix()
	got2 := GenPrefix()
	if len(got1) <= 0 || len(got2) <= 0 {
		t.Errorf("generated prefixes were empty")
	}
	if got1 == got2 {
		t.Errorf("generated prefixes should differ but they were the same, %s, %s", got1, got2)
	}
}
