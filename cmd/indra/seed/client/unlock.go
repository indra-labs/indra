package client

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"os"
)

var (
	unlockTargetFlag   = "target"
	unlockPassFileFlag = "keyfile"
)

var (
	unlockTarget       string
	unlockPassFilePath string
)

func initUnlock(cmd *cobra.Command) {

	cmd.Flags().StringVarP(&unlockTarget, unlockTargetFlag, "",
		"unix:///tmp/indra.sock",
		"the url of the rpc server",
	)

	viper.BindPFlag(unlockTargetFlag, cmd.Flags().Lookup(unlockTargetFlag))

	cmd.Flags().StringVarP(&unlockPassFilePath, unlockPassFileFlag, "",
		"",
		"",
	)

	viper.BindPFlag(unlockPassFileFlag, cmd.Flags().Lookup(unlockPassFileFlag))
}

var unlockRPCCmd = &cobra.Command{
	Use:   "unlock",
	Short: "unlocks the encrypted storage",
	Long:  `unlocks the encrypted storage.`,
	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var conn *grpc.ClientConn

		conn, err = rpc.Dial(viper.GetString(unlockTargetFlag))

		if err != nil {
			check(err)
			os.Exit(1)
		}

		var keyFileBytes []byte

		if keyFileBytes, err = os.ReadFile(viper.GetString(unlockPassFileFlag)); check(err) {
			os.Exit(0)
		}

		//var password []byte
		//
		//fmt.Print("Enter Encryption Key: ")
		//password, err = term.ReadPassword(int(syscall.Stdin))
		//fmt.Println()

		u := storage.NewUnlockServiceClient(conn)

		_, err = u.Unlock(context.Background(), &storage.UnlockRequest{
			Key: string(string(keyFileBytes)),
		})

		if err != nil {
			check(err)
			return
		}

		log.I.Ln("successfully unlocked")
	},
}
