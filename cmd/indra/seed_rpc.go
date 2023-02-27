package main

import (
	"context"
	"fmt"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"os"
	"syscall"
)

func init() {

	//// Init flags belonging to the seed package
	//seed.InitFlags(seedServeCommand)
	//
	//// Init flags belonging to the rpc package
	//rpc.InitFlags(seedServeCommand)

	seedCommand.AddCommand(seedRPCCmd)

	initUnlock(unlockRPCCmd)

	seedRPCCmd.AddCommand(unlockRPCCmd)
}

var seedRPCCmd = &cobra.Command{
	Use:   "rpc",
	Short: "A list of commands for interacting with a seed",
	Long:  `A list of commands for interacting with a seed.`,
}

var (
	unlockTargetFlag = "target"
)

var (
	unlockTarget string
)

func initUnlock(cmd *cobra.Command) {

	cmd.Flags().StringVarP(&unlockTarget, unlockTargetFlag, "",
		"unix:///tmp/indra.sock",
		"the url of the rpc server",
	)

	viper.BindPFlag(unlockTargetFlag, cmd.PersistentFlags().Lookup(unlockTargetFlag))

}

var unlockRPCCmd = &cobra.Command{
	Use:   "unlock",
	Short: "unlocks the encrypted storage",
	Long:  `unlocks the encrypted storage.`,
	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var conn *grpc.ClientConn

		conn, err = rpc.Dial(viper.GetString("target"))

		if err != nil {
			check(err)
			os.Exit(1)
		}

		var password []byte

		fmt.Print("Enter Encryption Key: ")
		password, err = term.ReadPassword(int(syscall.Stdin))
		fmt.Println()

		u := storage.NewUnlockServiceClient(conn)

		_, err = u.Unlock(context.Background(), &storage.UnlockRequest{
			Key: string(password),
		})

		if err != nil {
			check(err)
			return
		}

	},
}
