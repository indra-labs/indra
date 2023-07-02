// Package norm is a string comparison library that makes everything lowercase before comparison for case insensitive equality testing.
package norm

import (
	"strings"
)

func Eq(a, b string) bool {
	an, bn := Norm(a), Norm(b)
	return an == bn
}

func Norm(s string) string { return strings.ToLower(s) }
