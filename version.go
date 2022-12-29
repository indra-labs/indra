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
	ParentGitCommit = "2dc3027cf28b04b52b8bca3839a96267122f8b92"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2022-12-29T02:49:15Z"
	// SemVer lists the (latest) git tag on the build.
	SemVer = "v0.0.255"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/lyo/Seafile/Git/indra-labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 0
	// Patch is the patch version number from the tag.
	Patch = 255
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
