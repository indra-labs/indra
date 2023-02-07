package relay

// HandleTimeout is called automatically after an expected amount of time.
func (pr *PendingResponse) HandleTimeout() {
	log.D.Ln("response timeout")
}
