package indra

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = ""
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/main"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "907e50a6e03155b5ec546685d2b4df35326ba09c"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-01-29T11:44:40Z"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.0.5"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/opt/indra-labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 0
	// Patch is the patch version number from the tag.
	Patch = 5
)

// Version returns a pretty printed version information string.
func Version() string {
	return fmt.Sprint(
		"\nRepository Information\n",
		"\tGit repository: "+URL+"\n",
		"\tBranch: "+GitRef+"\n",
		"\tParentGitCommit: "+ParentGitCommit+"\n",
		"\tBuilt: "+BuildTime+"\n",
		"\tSemVer: "+SemVer+"\n",
	)
}
