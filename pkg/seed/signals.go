package seed

var (
	startupErrors  = make(chan error, 32)
	isReadyChan    = make(chan bool, 1)
	isShutdownChan = make(chan bool, 1)
)

func WhenStartFailed() chan error {
	return startupErrors
}

func IsReady() chan bool {
	return isReadyChan
}

func IsShutdown() chan bool {
	return isShutdownChan
}
