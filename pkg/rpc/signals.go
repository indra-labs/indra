package rpc

var (
	startupErrors = make(chan error, 128)
	isConfigured  = make(chan bool, 1)
	isReady       = make(chan bool, 1)
)

func WhenStartFailed() chan error {
	return startupErrors
}

func IsConfigured() chan bool {
	return isConfigured
}

func IsReady() chan bool {
	return isReady
}
