package version

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func Info() string {
	return fmt.Sprintf("jira version %s (commit: %s, built: %s)", Version, Commit, BuildTime)
}
