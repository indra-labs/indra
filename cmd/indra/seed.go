package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(seedCommand)
}

var seedCommand = &cobra.Command{
	Use:   "seed",
	Short: "run and manage your seed node",
	Long:  `run and manage your seed node`,
}
