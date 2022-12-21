package normalize

import (
	"net"
	"strconv"
)

// Address returns addr with the passed default port appended if there is not
// already a port specified.
func Address(addr, defaultPort string, userOnly bool) (a string, e error) {
	var p string
	a, p, e = net.SplitHostPort(addr)
	if log.E.Chk(e) || p == "" {
		return net.JoinHostPort(a, defaultPort), e
	}
	if userOnly {
		p = ClampPortRange(p, defaultPort, 1024, 65535)
	} else {
		p = ClampPortRange(p, defaultPort, 1, 65535)
	}
	return net.JoinHostPort(a, p), e
}

// Addresses returns a new slice with all the passed peer addresses normalized
// with the given default port, and all duplicates removed.
func Addresses(addrs []string, defaultPort string, userOnly bool) (a []string,
	e error) {

	for i := range addrs {
		addrs[i], e = Address(addrs[i], defaultPort, userOnly)
	}
	a = RemoveDuplicateAddresses(addrs)
	return
}

// RemoveDuplicateAddresses returns a new slice with all duplicate entries in
// addrs removed.
func RemoveDuplicateAddresses(addrs []string) (result []string) {
	result = make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

func ClampPortRange(port, defaultPort string, min, max int) string {
	p, err := strconv.Atoi(port)
	if err != nil {
		return defaultPort
	}
	if p < min {
		port = strconv.FormatInt(int64(min), 10)
	} else if p > max {
		port = strconv.FormatInt(int64(max), 10)
	}
	return port
}
