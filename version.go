//go:build !local

// Package indra is the root of the repository for the Indra distributed VPN, containing mainly the version information for included executables to use for information and identification on the network.
//
// See [pkg/git.indra-labs.org/dev/ind/cmd/indra] for the main server executable.
//
// Put invocations to run all the generators in here check [cmd/bumper] to add them, and they will automatically run with:
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
	URL = "git.indra-labs.org/dev/ind"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/15"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "742a44445cc6862042a76fdb02e64236fd4340bb"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2023-08-09T21:04:58+01:00"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "v0.1.20"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 1
	// Patch is the patch version number from the tag.
	Patch = 20
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
