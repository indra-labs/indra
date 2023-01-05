package indra

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = "github.com/Indra-Labs/indra"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/ind-bootstrap"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "de0d9d356f80f91633a7173f5f2fabb5737b48b2"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2022-12-31T15:18:29Z"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.1.2"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/lyo/Seafile/Git/indra-labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 1
	// Patch is the patch version number from the tag.
	Patch = 2
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
