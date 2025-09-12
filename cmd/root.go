package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "3mfanalyzer",
	Short: "A tool for analyzing 3MF files",
	Long:  `3mfanalyzer is a command-line tool for analyzing and extracting information from 3MF files.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Set config file path
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}
	
	// Look for config file in home directory
	viper.AddConfigPath(home)
	viper.SetConfigName(".3mfanalyzer")
	viper.SetConfigType("yaml")
	
	// Read config file if it exists
	viper.ReadInConfig()
	
	// Also read from environment variables
	viper.AutomaticEnv()
}