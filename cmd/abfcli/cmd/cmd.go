/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/pkg/logger"
	"github.com/spf13/cobra"
)

var ErrNotEnoughArguments = errors.New("not enough arguments to call command")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "abfcli",
	Short: "CLI client for anti-bruteforce service",
	Long: `Anti-bruteforce service is created to 
	allow or decine requests for given combinations of 
	login, password or ip address. 

	Standart path to config: configs/config_cli.toml
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("help") || cmd.Flags().Changed("h") {
			return
		}
	},
}

var configFile string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	cfg := *config.NewConfig(configFile)
	logg := logger.New(cfg)

	if err != nil {
		logg.Error(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tests.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "configs/config_cli.toml", "Path to configuration file")

	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(blackCmd)
	rootCmd.AddCommand(whiteCmd)
}
