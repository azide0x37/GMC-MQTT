package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildStateCalculations(t *testing.T) {
	state := buildState(100, 3.7, "1.0", "GMC-300", "ABC", 42)

	if state.CPM != 100 {
		t.Fatalf("expected cpm 100, got %v", state.CPM)
	}
	if state.Battery != 3.7 {
		t.Fatalf("expected battery 3.7, got %v", state.Battery)
	}
	if state.USV != 0.57 {
		t.Fatalf("expected usv 0.57, got %v", state.USV)
	}
	if state.MR != 0.057 {
		t.Fatalf("expected mr 0.057, got %v", state.MR)
	}
	if !state.Healthy {
		t.Fatalf("expected healthy state")
	}
	if state.Timestamp == "" {
		t.Fatalf("expected timestamp")
	}
}

func TestWriteStateAtomicallyWritesJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run", "state.json")
	state := buildState(7, 4.1, "fw", "model", "serial", 3)

	if err := writeState(path, state); err != nil {
		t.Fatalf("writeState failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	var decoded StateSnapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("state is not JSON: %v", err)
	}
	if decoded.CPM != 7 || decoded.Serial != "serial" {
		t.Fatalf("unexpected decoded state: %+v", decoded)
	}
}

func TestAppendLedgerWritesJSONLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	err := appendLedger(path, LedgerEvent{Event: "collector_error", Message: "boom"})
	if err != nil {
		t.Fatalf("appendLedger failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(data), `"event":"collector_error"`) {
		t.Fatalf("ledger missing event: %s", data)
	}
}
