package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"gopkg.in/src-d/go-git.v4"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func main() {
	remote := os.Args[1]
	tag := os.Args[2]
	dir := os.Args[3]
	e := os.RemoveAll(dir)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	repo, e := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      remote,
		Progress: os.Stdout,
	})
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	branch, e := repo.Tag(tag)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	w, e := repo.Worktree()
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	e = w.Checkout(&git.CheckoutOptions{
		Branch: branch.Name(),
	})
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	e = os.RemoveAll(filepath.Join(dir, ".git"))
	check(e)
	e = os.RemoveAll(filepath.Join(dir, ".github"))
	check(e)
	e = os.RemoveAll(filepath.Join(dir, ".vscode"))
	check(e)
	// e = os.RemoveAll(filepath.Join(dir, "cmd"))
	// check(e)
	e = os.Remove(filepath.Join(dir, "go.mod"))
	check(e)
	e = os.Remove(filepath.Join(dir, "go.sum"))
	check(e)
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		file, e := os.ReadFile(path)
		if e != nil {
			return nil
		}
		_, filename := filepath.Split(path)
		if filename == "go.mod" ||
			filename == "go.sum" ||
			strings.HasSuffix(filename, "_test.go") ||
			strings.HasSuffix(filename, ".md") {
			e = os.Remove(path)
			check(e)
		} else {
			e = os.WriteFile(path,
				[]byte(strings.ReplaceAll(string(file),
					"\"github.com/lightningnetwork/lnd",
					"\"github.com/indra-labs/indra/pkg/lnd")), 0755)
			check(e)
		}
		return nil
	})
	runCmdWithOutput("go", "mod", "tidy")
}

func runCmdWithOutput(cmd ...string) (out string, err error) {
	c := exec.Command(cmd[0], cmd[1:]...)
	var output []byte
	output, err = c.CombinedOutput()
	if err == nil && string(output) != "" {
		errPrintln(string(output))
	}
	out = string(output)
	return
}

func errPrintln(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}
