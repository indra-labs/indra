package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var max int
	var longest string
	filepath.Walk(os.Args[1], func(path string, info fs.FileInfo, err error) error {
		if strings.HasPrefix(path, ".") ||
			!strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") { // doesn't matter as much if test logs rel path grow initially
			return nil
		}
		if len(path) > max {
			max = len(path)
			longest = path
		}
		return nil
	})
	if e := os.WriteFile("pkg/proc/log/length.go", []byte(fmt.Sprintf("package log\n\nconst maxLen = %d\n\n", max)), 0600); e != nil {
		fmt.Println(e)
	}
	fmt.Println(longest)
}
