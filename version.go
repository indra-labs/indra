package indra

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = "github.com/indra-labs/indra"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/protocol"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "0a37f65155ad746edf54a4f653d65bfc889e078a"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-01-08T13:56:39Z"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.1.5"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/loki/src/github.com/indra-labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 1
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
