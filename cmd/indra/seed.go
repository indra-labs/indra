package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(seedCommand)
}

var seedCommand = &cobra.Command{
	Use:   "seed",
	Short: "Commands related to seeding",
	Long:  `Commands related to seeding.`,
}
