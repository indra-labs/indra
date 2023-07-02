// Package windows provides some tools for handling launching subprocesses on windows using cmd.exe.
package windows

import (
	"runtime"
)

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
