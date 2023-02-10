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
	ParentGitCommit = "79dd94cb702c2c4d5dc3510e791b7ae9ef2493ef"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-02-10T15:55:02Z"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.0.7"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/lyo/Seafile/Git/indra-labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 0
	// Patch is the patch version number from the tag.
	Patch = 7
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
