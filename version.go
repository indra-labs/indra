// Package indra is the root level package for Indranet, a low latency, 
// Lightning Network monetised distributed VPN protocol designed for providing
// strong anonymity to valuable internet traffic.
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
	ParentGitCommit = "461c4d70ab24685a7f3973a55c86bf840f6d022a"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2022-12-09T11:29:44+01:00"
	// SemVer lists the (latest) git tag on the build.
	SemVer = "v0.0.162"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/loki/src/github.com/Indra-Labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 0
	// Patch is the patch version number from the tag.
	Patch = 162
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
