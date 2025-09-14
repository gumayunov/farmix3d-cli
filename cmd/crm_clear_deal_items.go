package cmd

import (
	"fmt"
	"os"

	"3mfanalyzer/internal/bitrix"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clearDealID string
	clearDryRun bool
)

var crmClearDealItemsCmd = &cobra.Command{
	Use:   "crm-clear-deal-items",
	Short: "Clear all products and services from a Bitrix24 deal",
	Long: `Clear all products and services from a Bitrix24 deal.

This command will:
1. Get deal information from Bitrix24
2. Show all existing products/services in the deal
3. Remove all products/services from the deal

Use --dry-run flag to preview what products would be cleared without making changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCRMClearDealItems(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runCRMClearDealItems() error {
	// Validate parameters
	if err := bitrix.ValidateDealID(clearDealID); err != nil {
		return fmt.Errorf("invalid deal ID: %v", err)
	}

	// Get webhook URL from config
	webhookURL := viper.GetString("bitrix_webhook_url")
	if webhookURL == "" {
		return fmt.Errorf("bitrix_webhook_url not configured. Please set it in ~/.3mfanalyzer config")
	}

	if clearDryRun {
		fmt.Printf("[DRY RUN] Processing deal %s...\n", clearDealID)
	} else {
		fmt.Printf("Processing deal %s...\n", clearDealID)
	}

	// Create Bitrix24 client
	client := bitrix.NewClient(webhookURL)

	// Clear deal product rows
	err := client.ClearDealProductRows(clearDealID, clearDryRun)
	if err != nil {
		return fmt.Errorf("failed to clear deal items: %v", err)
	}

	if clearDryRun {
		fmt.Printf("[DRY RUN] Clear deal items operation completed\n")
	} else {
		fmt.Printf("Clear deal items operation completed successfully\n")
	}

	return nil
}

func init() {
	crmClearDealItemsCmd.Flags().StringVar(&clearDealID, "deal-id", "", "Bitrix24 deal ID (required)")
	crmClearDealItemsCmd.Flags().BoolVar(&clearDryRun, "dry-run", false, "Preview what would be cleared without making changes")

	crmClearDealItemsCmd.MarkFlagRequired("deal-id")

	rootCmd.AddCommand(crmClearDealItemsCmd)
}