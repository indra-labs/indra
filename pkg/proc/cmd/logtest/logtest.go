package main

import (
	"errors"
	"time"

	"github.com/cybriq/proc"
	log2 "github.com/cybriq/proc/pkg/log"
)

var log = log2.GetLogger(proc.PathBase)

func main() {
	log2.App = "logtest"
	log2.SetLogLevel(log2.Trace)
	log2.CodeLoc = false
	log2.SetTimeStampFormat(time.Stamp)
	log.I.C(proc.Version)
	log.T.Ln("testing")
	log.D.Ln("testing")
	log.I.Ln("testing")
	log.W.Ln("testing")
	log.E.Chk(errors.New("testing"))
	log.F.Ln("testing")
	log.I.S(log2.GetAllSubsystems())
	log.I.Ln("setting timestamp format to RFC822Z")
	log2.SetTimeStampFormat(time.RFC822Z)
	log.I.Ln("setting log level to info and printing from all levels")
	log2.SetLogLevel(log2.Info)
	log.T.Ln("testing")
	log.D.Ln("testing")
	log.I.Ln("testing")
	log.W.Ln("testing")
	log.E.Chk(errors.New("testing"))
	log.F.Ln("testing")
	log.T.Ln("testing")

	log2.CodeLoc = true
	log2.SetLogLevel(log2.Trace)
	log.I.C(proc.Version)
	log.T.Ln("testing")
	log.D.Ln("testing")
	log.I.Ln("testing")
	log.W.Ln("testing")
	log.E.Chk(errors.New("testing"))
	log.F.Ln("testing")
	log.I.S(log2.GetAllSubsystems())
	log.I.Ln("setting log level to info and printing from all levels")
	log2.SetLogLevel(log2.Info)
	log.T.Ln("testing")
	log.D.Ln("testing")
	log.I.Ln("testing")
	log.W.Ln("testing")
	log.E.Chk(errors.New("testing"))
	log.F.Ln("testing")
	log.T.Ln("testing")

}
