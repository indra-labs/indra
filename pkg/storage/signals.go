package storage

var (
	startupErrors  = make(chan error, 128)
	isLockedChan   = make(chan bool, 1)
	isUnlockedChan = make(chan bool, 1)
	isReadyChan    = make(chan bool, 1)
)

func WhenStartupFailed() chan error {
	return startupErrors
}

func WhenIsLocked() chan bool {
	return isLockedChan
}

func WhenIsUnlocked() chan bool {
	return isUnlockedChan
}

func WhenIsReady() chan bool {
	return isReadyChan
}
