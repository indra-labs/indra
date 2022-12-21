package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureDir checks a file could be written to a path, creates the directories as needed
func EnsureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
}

// FileExists reports whether the named file or directory exists.
func FileExists(filePath string) bool {
	_, e := os.Stat(filePath)
	return e == nil
}

// MinUint32 is a helper function to return the minimum of two uint32s. This avoids a math import and the need to cast
// to floats.
func MinUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// PrependForWindows runs a command with a terminal
func PrependForWindows(args []string) []string {
	if runtime.GOOS == "windows" {
		args = append(
			[]string{
				"cmd.exe",
				"/C",
			},
			args...,
		)
	}
	return args
}

// PrependForWindowsWithStart runs a process independently
func PrependForWindowsWithStart(args []string) []string {
	if runtime.GOOS == "windows" {
		args = append(
			[]string{
				"cmd.exe",
				"/C",
				"start",
			},
			args...,
		)
	}
	return args
}

func Norm(s string) string {
	return strings.ToLower(s)
}
