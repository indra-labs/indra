package indra


// Put invocations to run all the generators in here (
// check cmd/bumper/ to add them, and they will automatically run with:
//
// $ go generate .
//
// which will run all these generators below and finish with a go install.
//go:generate go run ./pkg/relay/gen/main.go
//go:generate go install ./...

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = ""
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/main"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "e7f66e13659ff52dfedfca1dc76a0facc91674b7"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-03-01T09:28:47Z"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.0.10"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/opt/indra-labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 0
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
