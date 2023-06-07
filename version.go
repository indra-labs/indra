//go:build !local

// This can be overridden by a developer's version by setting the tag local
// on a modified version. This is useful for the code locations in teh logs.

package indra

import "fmt"

// Put invocations to run all the generators in here (
// check cmd/bumper/ to add them, and they will automatically run with:
//
// $ go generate .
//
// which will run all these generators below and finish with a go install.
//go:generate go install ./...

const (
	// URL is the git URL for the repository.
	URL = "github.com/indra-labs/indra"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/master"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "050e32309721c4423d6e84ea13e0359fc85ecfeb"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-06-07T19:42:31+01:00"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.1.13"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 1
	// Patch is the patch version number from the tag.
	Patch = 13
)

var CI = "false"

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
