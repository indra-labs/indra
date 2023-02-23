package rpc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
)

var (
	isUnixSockEnabled bool = false
	unixSock          net.Listener
)

var (
	unixPathFlag  = "rpc-unix-listen"
	unixPathUsage = "binds to a unix socket with path"
	unixPath      = "/tmp/indra.sock"
)

func defineUnixSocket(cmd *cobra.Command) {

	cmd.PersistentFlags().StringVarP(
		&unixPath,
		unixPathFlag,
		"",
		unixPath,
		unixPathUsage,
	)

	viper.BindPFlag(
		unixPathFlag,
		cmd.PersistentFlags().Lookup(unixPathFlag),
	)
}

func configureUnixSocket() {

	if viper.GetString(unixPathFlag) == "" {
		return
	}

	log.I.Ln("enabling unix listener:", viper.GetString(unixPath))

	isUnixSockEnabled = true
}

func startUnixSocket() (err error) {

	if !isUnixSockEnabled {
		return
	}

	if unixSock, err = net.Listen("unix", unixPath); check(err) {
		return
	}

	go server.Serve(unixSock)

	return
}

func stopUnixSocket() (err error) {

	if !isUnixSockEnabled {
		return
	}

	if err = unixSock.Close(); check(err) {
		// continue
	}

	//if err = os.Remove(unixPath); check(err) {
	//	// continue
	//}

	return
}
