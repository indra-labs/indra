package normalize

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/cybriq/proc/pkg/appdata"
)

func ResolvePath(input, appName string, abs bool) (cleaned string, e error) {
	if strings.HasPrefix(input, "~") {
		homeDir := getHomeDir()
		input = strings.Replace(input, "~", homeDir, 1)
		cleaned = filepath.Clean(input)
	} else {

		if abs {
			if cleaned, e = filepath.Abs(cleaned); log.E.Chk(e) {
				return
			}
			// if the path is relative, either ./ or not starting with a / then
			// we assume the path is relative to the app data directory
		} else if !strings.HasPrefix(string(os.PathSeparator), cleaned) ||
			strings.HasPrefix("."+string(os.PathSeparator), cleaned) {

			cleaned = filepath.Join(appdata.Dir(appName, false), cleaned)
		}
	}
	return
}

func getHomeDir() (homeDir string) {
	var usr *user.User
	var e error
	if usr, e = user.Current(); !log.E.Chk(e) {
		homeDir = usr.HomeDir
	}
	// Fall back to standard HOME environment variable that
	// works for most POSIX OSes if the directory from the
	// Go standard lib failed.
	if e != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
