package relay

import (
	"github.com/spf13/cobra"
)

func Init(c *cobra.Command) {
	c.AddCommand(relayCommand)
}

var relayCommand = &cobra.Command{
	Use:   "relay",
	Short: "run a relay",
	Long:  "Runs indra as a full relay, with optional client",
		Run: func(cmd *cobra.Command, args []string) {
		},
}
