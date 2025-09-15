package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "farmix-cli",
	Short: "Farmix CLI - инструмент для 3D печати и анализа файлов",
	Long:  `farmix-cli - консольная утилита для анализа 3MF файлов, слайсинга STL моделей, расчета объемов и интеграции с Bitrix24 CRM.`,
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
	viper.SetConfigName(".farmix-cli")
	viper.SetConfigType("yaml")
	
	// Read config file if it exists
	viper.ReadInConfig()
	
	// Also read from environment variables
	viper.AutomaticEnv()
}