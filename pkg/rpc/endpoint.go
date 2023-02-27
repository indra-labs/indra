package rpc

import "strings"

type Endpoint string

func (e Endpoint) Address() string {

	before, _, found := strings.Cut(string(e), ":")

	if !found {
		return ""
	}

	return before
}

func (e Endpoint) Port() string {

	_, after, found := strings.Cut(string(e), ":")

	if !found {
		return ""
	}

	return after
}

func (e Endpoint) String() string {
	return string(e)
}

func EndpointString(endpoint string) (ep Endpoint) {

	_, after, found := strings.Cut(string(endpoint), "//")

	if !found {
		ep = Endpoint(endpoint)

		return
	}

	ep = Endpoint(after)

	return
}
