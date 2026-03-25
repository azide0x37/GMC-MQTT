package main

import "testing"

func TestBuildStateCalculations(t *testing.T) {
	state := buildState(100, 3.7, "1.0", "GMC-300", "ABC", 42)

	if state["cpm"] != uint16(100) {
		t.Fatalf("expected cpm 100, got %v", state["cpm"])
	}
	if state["battery"] != 3.7 {
		t.Fatalf("expected battery 3.7, got %v", state["battery"])
	}
	if state["usv"] != 0.57 {
		t.Fatalf("expected usv 0.57, got %v", state["usv"])
	}
	if state["mr"] != 0.057 {
		t.Fatalf("expected mr 0.057, got %v", state["mr"])
	}
}
