package indra

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = "github.com/Indra-Labs/indra"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/main"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "c091541adc4919c9c688e72716a6bd75a44213d1"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2022-10-25T09:24:56+02:00"
	// SemVer lists the (latest) git tag on the build.
	SemVer = "v0.0.72"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/loki/src/github.com/Indra-Labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 0
	// Patch is the patch version number from the tag.
	Patch = 72
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
