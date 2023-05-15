package indra

import (
	"fmt"
)

const (
	// URL is the git URL for the repository.
	URL = ""
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/protocol"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "028d2aa98a187875f4913f3476bc248a5003ae6b"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-05-12T20:20:16+01:00"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.1.11"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/opt/indra-labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 1
	// Patch is the patch version number from the tag.
	Patch = 11
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
