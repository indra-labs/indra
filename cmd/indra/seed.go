package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "run and manage your seed node",
	Long:  `run and manage your seed node`,
}
