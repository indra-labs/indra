package norm

import (
	"strings"
)

func Eq(a, b string) bool {
	an, bn := Norm(a), Norm(b)
	return an == bn
}

func Norm(s string) string { return strings.ToLower(s) }
