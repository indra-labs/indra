// Bumper is a tool for creating version information to be placed at the repository root of a project.
//
// It provides basic release information, references the parent Git commit hash, automatically increments the minor version, tags the commit with the version so that it is easy for importing projects to use a Semantic Versioning version code instead of depending on automatic generated codes from Go's module system.
//
// If 'minor' or 'major' are the first space-separated command line parameter, it bumps the respective number instead of the patch version.
//
// To indicate a line break, to make 72 column commit comments with additional lines, use '--' and it and the surrounding spaces are converted to a double line break/carriage return, inserting an empty line, effectively.
//
// todo: maybe even better, just have it automatically break the thing at the last whole word before 72?
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

var (
	URL                 string
	GitRef              string
	ParentGitCommit     string
	BuildTime           string
	SemVer              string
	Major, Minor, Patch int
	PathBase            string
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

func errPrintln(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func main() {
	log2.App.Store("bumper")
	if len(os.Args) < 2 {
		log.E.Ln("at least one argument is required in order to bump and push this repo")
		os.Exit(1)
	}
	var minor, major bool
	switch {
	case os.Args[1] == "minor":
		minor = true
		os.Args = append(os.Args[0:1], os.Args[2:]...)
	case os.Args[1] == "major":
		major = true
		os.Args = append(os.Args[0:1], os.Args[2:]...)
	}
	log2.App.Store("bumper")
	BuildTime = time.Now().Format(time.RFC3339)
	var cwd string
	var e error
	if cwd, e = os.Getwd(); check(e) {
		os.Exit(1)
	}
	var repo *git.Repository
	if repo, e = git.PlainOpen(cwd); check(e) {
		os.Exit(1)
	}
	var rr []*git.Remote
	if rr, e = repo.Remotes(); check(e) {
		os.Exit(1)
	}
	for i := range rr {
		rs := rr[i].String()
		if strings.HasPrefix(rs, "origin") {
			rss := strings.Split(rs, "git@")
			if len(rss) > 1 {
				rsss := strings.Split(rss[1], ".git")
				URL = strings.ReplaceAll(rsss[0], ":", "/")
				break
			}
			// This command must be used with ssh addresses only.
			// rss = strings.Split(rs, "https://")
			// if len(rss) > 1 {
			// 	rsss := strings.Split(rss[1], ".git")
			// 	URL = rsss[0]
			// 	break
			// }
		}
	}
	var tr *git.Worktree
	if tr, e = repo.Worktree(); check(e) {
	}
	var rh *plumbing.Reference
	if rh, e = repo.Head(); check(e) {
		os.Exit(1)
	}
	rhs := rh.Strings()
	GitRef = rhs[0]
	ParentGitCommit = rhs[1]
	var rt storer.ReferenceIter
	if rt, e = repo.Tags(); check(e) {
		os.Exit(1)
	}
	var maxVersion int
	if e = rt.ForEach(
		func(pr *plumbing.Reference) (e error) {
			s := strings.Split(pr.String(), "/")
			prs := s[2]
			if strings.HasPrefix(prs, "v") {
				var va [3]int
				_, _ = fmt.Sscanf(
					prs,
					"v%d.%d.%d",
					&va[0],
					&va[1],
					&va[2],
				)
				vn := va[0]*1000000 + va[1]*1000 + va[2]
				if maxVersion < vn {
					maxVersion = vn
					Major = va[0]
					Minor = va[1]
					Patch = va[2]
				}
			}
			return
		},
	); check(e) {
		return
	}
	br := strings.Split(GitRef, "/")
	branch := br[len(br)-1]
	startArgs := 1
	branchParam := os.Args[1]
	if major || minor {
		branchParam = os.Args[2]
	}
	var out string
	if out, e = runCmdWithOutput("git", "branch"); check(e) {
		os.Exit(1)
	}
	splitted := strings.Split(out, "\n")
	var isBranch bool
	for i := range splitted {
		if len(splitted[i]) < 2 {
			continue
		}
		if splitted[i][2:] == branchParam {
			isBranch = true
			break
		}
	}
	if isBranch {
		branch = branchParam
	}
	if isBranch {
		startArgs++
	}
	tag := true
	if branch != "main" && branch != "master" {
		tag = false
	} else {
		switch {
		case minor:
			Minor++
			Patch = 0
		case major:
			Major++
			Minor = 0
			Patch = 0
		default:
			Patch++
		}
	}
	SemVer = fmt.Sprintf("v%d.%d.%d", Major, Minor, Patch)
	PathBase = tr.Filesystem.Root() + "/"
	var dir string
	if dir, e = os.Getwd(); check(e) {
	}
	name := filepath.Base(dir)
	versionFile := `//go:build !local

// Package indra is the root of the repository for the Indra distributed VPN, containing mainly the version information for included executables to use for information and identification on the network.
//
// See [pkg/git.indra-labs.org/dev/ind/cmd/indra] for the main server executable.
//
// Put invocations to run all the generators in here check [pkg/git.indra-labs.org/dev/ind/cmd/bumper] to add them, and they will automatically run with:
//
// $ go generate .
//
// which will run all these generators below and finish with a go install.
//
//` + `go:generate go install ./...
package ` + name + `

import "fmt"

const (
	// URL is the git URL for the repository.
	URL = "%s"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "%s"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "%s"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "%s"
	// SemVer lists the (latest) git tag on the release.
	SemVer = "%s"
	// Major is the major number from the tag.
	Major = %d
	// Minor is the minor number from the tag.
	Minor = %d
	// Patch is the patch version number from the tag.
	Patch = %d
)

var CI = "%s"

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
`
	versionFileOut := fmt.Sprintf(
		versionFile,
		URL,
		GitRef,
		ParentGitCommit,
		BuildTime,
		SemVer,
		Major,
		Minor,
		Patch,
		"false",
	)
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "version.go")
	if e = os.WriteFile(path, []byte(versionFileOut), 0666); check(e) {
		os.Exit(1)
	}
	log.I.Ln(
		"\nRepository Information\n",
		"\tGit repository: "+URL+"\n",
		"\tBranch: "+GitRef+"\n",
		"\tParentGitCommit: "+ParentGitCommit+"\n",
		"\tBuilt: "+BuildTime+"\n",
		"\tSemVer: "+SemVer+"\n",
		"\tMajor:", Major, "\n",
		"\tMinor:", Minor, "\n",
		"\tPatch:", Patch, "\n",
	)
	if e = runCmd("git", "add", "."); check(e) {
		os.Exit(1)
	}
	commitString := strings.Join(os.Args[startArgs:], " ")
	commitString = strings.ReplaceAll(commitString, " -- ", "\n\n")
	if e = runCmd("git", "commit", "-m"+commitString); check(e) {
		os.Exit(1)
	}
	if tag {
		if e = runCmd("git", "tag", SemVer); check(e) {
			os.Exit(1)
		}
	}
	if e = runCmd("git", "push", "origin", branch); check(e) {
		os.Exit(1)
	}
	if e = runCmd("git", "push", "origin", SemVer); check(e) {
		os.Exit(1)
	}
	// Lastly, we need to regenerate the version of bumper if it changed.
	// Rather than check, we will just run the compilation command anyway.
	if e = runCmd("go", "install", "./cmd/bumper/."); e != nil {
		os.Exit(1)
	}
	return
}

func runCmd(cmd ...string) (err error) {
	c := exec.Command(cmd[0], cmd[1:]...)
	var output []byte
	output, err = c.CombinedOutput()
	if err == nil && string(output) != "" {
		errPrintln(string(output))
	}
	return
}

func runCmdWithOutput(cmd ...string) (out string, err error) {
	c := exec.Command(cmd[0], cmd[1:]...)
	var output []byte
	output, err = c.CombinedOutput()
	// if err == nil && string(output) != "" {
	// 	errPrintln(string(output))
	// }
	out = string(output)
	return
}
