package proc

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = "github.com/cybriq/proc"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/master"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "85b3415a92a7b1cd3ce86d040105408aa5159e0c"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2022-12-30T16:05:24Z"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.20.10"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/loki/src/github.com/cybriq/proc/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 20
	// Patch is the patch version number from the tag.
	Patch = 10
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
