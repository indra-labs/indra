package seed

import "github.com/spf13/cobra"

var rpcCmd = &cobra.Command{
	Use:   "rpc",
	Short: "A list of commands for interacting with a seed",
	Long:  `A list of commands for interacting with a seed.`,
}
