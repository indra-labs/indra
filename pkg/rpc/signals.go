package rpc

var (
	startupErrors = make(chan error, 128)
	isReady       = make(chan bool, 1)
)

func WhenStartFailed() chan error {
	return startupErrors
}

func IsReady() chan bool {
	return isReady
}
