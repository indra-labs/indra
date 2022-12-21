package qu

import (
	logg "log"
	"os"
)

var log = logg.New(os.Stderr, "basedd ", logg.Llongfile)
