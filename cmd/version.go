package main

import "fmt"

var (
	Version    = "dev"
	Commit     = "unknown"
	CommitDate = "unknown"
	TreeState  = "dirty"
)

func versionString() string {
	return fmt.Sprintf(
		"gmc-mqtt version=%s commit=%s commit_date=%s tree_state=%s",
		Version,
		Commit,
		CommitDate,
		TreeState,
	)
}
