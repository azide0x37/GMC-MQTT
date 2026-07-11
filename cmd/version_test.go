package main

import (
	"strings"
	"testing"
)

func TestVersionStringContainsBuildMetadata(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalCommitDate := CommitDate
	originalTreeState := TreeState
	t.Cleanup(func() {
		Version = originalVersion
		Commit = originalCommit
		CommitDate = originalCommitDate
		TreeState = originalTreeState
	})

	Version = "v1.2.3"
	Commit = "abc123def456"
	CommitDate = "2026-03-25T12:00:00Z"
	TreeState = "clean"

	got := versionString()
	for _, want := range []string{
		"v1.2.3",
		"abc123def456",
		"2026-03-25T12:00:00Z",
		"clean",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("versionString() = %q, want substring %q", got, want)
		}
	}
}
