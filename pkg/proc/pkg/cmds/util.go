package cmds

import "strings"

func IsBoolString(s string) (is bool) {
	switch strings.TrimSpace(s) {
	case "f", "false", "off", "-", "t", "true", "on", "+":
		return true
	default:
		return
	}
}
