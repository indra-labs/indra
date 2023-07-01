//go:build !local

// Package indra is the root of the repository for the Indra distributed VPN, containing mainly the version information for included executables to use for information and identification on the network.
//
// See [pkg/github.com/indra-labs/indra/cmd/indra] for the main server executable.
//
// Put invocations to run all the generators in here check [pkg/github.com/indra-labs/indra/cmd/bumper] to add them, and they will automatically run with:
//
// $ go generate .
//
// which will run all these generators below and finish with a go install.
//
//go:generate go install ./...
package indra

import "fmt"

const (
	// URL is the git URL for the repository.
	URL = "github.com/indra-labs/indra"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/master"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "88d5987612e40fe889fa0f8ac553a9b3a22c01ff"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-07-01T20:50:47+01:00"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.1.16"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 1
	// Patch is the patch version number from the tag.
	Patch = 16
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
