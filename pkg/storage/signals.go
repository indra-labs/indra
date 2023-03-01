package storage

var (
	startupErrors  = make(chan error, 128)
	isLockedChan   = make(chan bool, 1)
	isUnlockedChan = make(chan bool, 1)
	isReadyChan    = make(chan bool, 1)
)

var (
	isReady bool
)

func WhenStartFailed() chan error {
	return startupErrors
}

func WhenIsLocked() chan bool {
	return isLockedChan
}

func WhenIsUnlocked() chan bool {
	return isUnlockedChan
}

func WhenReady() chan bool {
	return isReadyChan
}
