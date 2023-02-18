package main

import (
	"errors"
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var indraTxt = `indra (` + indra.SemVer + `) - Network Freedom.

indra is a lightning powered, distributed virtual private network for anonymising traffic on decentralised protocol networks.
`

var (
	cfgFile   string
	cfgSave   bool
	logsDir   string
	logsLevel string
	dataDir   string
	network   string

	rootCmd = &cobra.Command{
		Use:   "indra",
		Short: "Network Freedom.",
		Long:  indraTxt,
	}
)

func init() {

	viper.SetEnvPrefix("INDRA")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogging)
	cobra.OnInitialize(initData)

	cobra.OnFinalize(persistConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config-file", "C", "", "config file (default is $HOME/.indra.toml)")
	rootCmd.PersistentFlags().BoolVarP(&cfgSave, "config-save", "", false, "saves the config file with any eligible envs/flags passed")
	rootCmd.PersistentFlags().StringVarP(&logsDir, "logs-dir", "L", "", "logging directory (default is $HOME/.indra/logs)")
	rootCmd.PersistentFlags().StringVarP(&logsLevel, "logs-level", "", "info", "set logging level  off|fatal|error|warn|info|check|debug|trace")
	rootCmd.PersistentFlags().StringVarP(&dataDir, "data-dir", "D", "", "data directory (default is $HOME/.indra/data)")
	rootCmd.PersistentFlags().StringVarP(&network, "network", "N", "mainnet", "selects the network  mainnet|testnet|simnet")

	viper.BindPFlag("logs-dir", rootCmd.PersistentFlags().Lookup("logs-dir"))
	viper.BindPFlag("logs-level", rootCmd.PersistentFlags().Lookup("logs-level"))
	viper.BindPFlag("data-dir", rootCmd.PersistentFlags().Lookup("data-dir"))
	viper.BindPFlag("network", rootCmd.PersistentFlags().Lookup("network"))
}

func initData() {

	if dataDir == "" {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		dataDir = home + "/indra/data"
	}

}

func initLogging() {

	if logsDir == "" {

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		logsDir = home + "/indra/logs"
	}

	log2.SetLogLevel(log2.GetLevelByString(viper.GetString("logs-level"), log2.Debug))
	log2.CodeLocations(false)

	if log2.GetLogLevel() == log2.Debug {
		log2.CodeLocations(true)
	}
}

func initConfig() {

	if cfgFile == "" {

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cfgFile = home + "/.indra.toml"
	}

	viper.SetConfigFile(cfgFile)

	if _, err := os.Stat(cfgFile); errors.Is(err, os.ErrNotExist) {
		return
	}

	if err := viper.ReadInConfig(); err != nil {
		log.E.Ln("failed to read config file:", err)
		os.Exit(1)
	}

}

func persistConfig() {

	if !cfgSave {
		return
	}

	if err := viper.WriteConfig(); err != nil {
		log.E.Ln("failed to save config file:", err)
		os.Exit(1)
	}
}
