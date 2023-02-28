package storage

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	storeKeyFlag     = "store-key"
	storeKeyFileFlag = "store-keyfile"
	//storeKeyRPCFlag   = "store-key-rpc"
	storeFilePathFlag = "store-path"
	//storeAskPassFlag  = "store-ask-pass"
)

var (
	storeEncryptionKey     string
	storeEncryptionKeyFile string
	//storeEncryptionKeyRPC  bool
	storeFilePath string
	//storeAskPass  bool
)

func InitFlags(cmd *cobra.Command) {

	cmd.Flags().StringVarP(&storeEncryptionKey, storeKeyFlag, "",
		"",
		"the key required to unlock storage",
	)

	viper.BindPFlag(storeKeyFlag, cmd.Flags().Lookup(storeKeyFlag))

	cmd.PersistentFlags().StringVarP(&storeEncryptionKeyFile, storeKeyFileFlag, "",
		"",
		"the path of the keyfile required to unlock storage",
	)

	viper.BindPFlag(storeKeyFileFlag, cmd.PersistentFlags().Lookup(storeKeyFileFlag))

	cmd.PersistentFlags().StringVarP(&storeFilePath, storeFilePathFlag, "",
		"",
		"the path of the database  (defaults to --data-dir/indra.db)",
	)

	viper.BindPFlag(storeFilePathFlag, cmd.PersistentFlags().Lookup(storeFilePathFlag))

	//cmd.PersistentFlags().BoolVarP(&storeEncryptionKeyRPC, storeKeyRPCFlag, "",
	//	false,
	//	"looks for the encryption key via RPC",
	//)
	//
	//viper.BindPFlag(storeKeyRPCFlag, cmd.PersistentFlags().Lookup(storeKeyRPCFlag))

	//cmd.PersistentFlags().BoolVarP(&storeAskPass, storeAskPassFlag, "",
	//	false,
	//	"prompts the user for a password to unlock storage",
	//)
	//
	//viper.BindPFlag(storeAskPassFlag, cmd.PersistentFlags().Lookup(storeAskPassFlag))
}
