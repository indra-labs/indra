// Package main is a subset of bumper, all it does mainly is refresh the
// PathBase, so that log prints correctly parse the embedded debug information
// used to show the source code location.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
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
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func errPrintln(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func main() {
	log2.App = "buidl"
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
			rss = strings.Split(rs, "https://")
			if len(rss) > 1 {
				rsss := strings.Split(rss[1], ".git")
				URL = rsss[0]
				break
			}

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
	// Update SemVer
	SemVer = fmt.Sprintf("v%d.%d.%d", Major, Minor, Patch)
	PathBase = tr.Filesystem.Root() + "/"
	versionFile := `// Package indra is the root level package for Indranet, a low latency, 
// Lightning Network monetised distributed VPN protocol designed for providing
// strong anonymity to valuable internet traffic.
package indra

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = "%s"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "%s"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "%s"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "%s"
	// SemVer lists the (latest) git tag on the build.
	SemVer = "%s"
	// PathBase is the path base returned from runtime caller.
	PathBase = "%s"
	// Major is the major number from the tag.
	Major = %d
	// Minor is the minor number from the tag.
	Minor = %d
	// Patch is the patch version number from the tag.
	Patch = %d
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
`
	versionFileOut := fmt.Sprintf(
		versionFile,
		URL,
		GitRef,
		ParentGitCommit,
		BuildTime,
		SemVer,
		PathBase,
		Major,
		Minor,
		Patch,
	)
	path := filepath.Join(PathBase, "version.go")
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
	// Lastly, we need to regenerate the version of bumper if it changed.
	// Rather than check, we will just run the compilation command anyway.
	if e = runCmd("go", "install", "./cmd/buidl/."); check(e) {
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
