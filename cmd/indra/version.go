package main

import (
	"fmt"
	"github.com/indra-labs/indra"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version number",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(indra.SemVer)
	},
}
