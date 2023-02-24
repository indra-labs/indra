package main

import (
	"github.com/spf13/cobra"
)

func init() {

	//// Init flags belonging to the seed package
	//seed.InitFlags(seedServeCommand)
	//
	//// Init flags belonging to the rpc package
	//rpc.InitFlags(seedServeCommand)

	seedCommand.AddCommand(seedRPCCmd)
}

var seedRPCCmd = &cobra.Command{
	Use:   "rpc",
	Short: "A list of commands for interacting with a seed",
	Long:  `A list of commands for interacting with a seed.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
