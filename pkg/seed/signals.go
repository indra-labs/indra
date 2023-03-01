package seed

var (
	startupErrors  = make(chan error, 32)
	isReadyChan    = make(chan bool, 1)
	isShutdownChan = make(chan bool, 1)
)

func WhenStartFailed() chan error {
	return startupErrors
}

func WhenReady() chan bool {
	return isReadyChan
}

func WhenShutdown() chan bool {
	return isShutdownChan
}
