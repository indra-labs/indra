package main

import (
	"fmt"
	"git.indra-labs.org/dev/ind"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version number",
	Long:  `All software has versions. This is mine. Semver formatted.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(indra.SemVer)
	},
}
