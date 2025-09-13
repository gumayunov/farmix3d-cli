package cmd

import (
	"fmt"
	"os"

	"3mfanalyzer/internal/bitrix"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	spreadDealID string
	spreadMethod string
	spreadDryRun bool
)

var crmSpreadPriceCmd = &cobra.Command{
	Use:   "crm-spread-price",
	Short: "Distribute deal amount among products proportionally",
	Long: `Distribute the total deal amount among products proportionally based on the selected method.

This command will:
1. Get deal information and total amount from Bitrix24
2. Get all products in the deal
3. Calculate proportional prices based on the method
4. Update product unit prices with proper rounding

Available methods:
  count  - Distribute based on product quantities (default)
  volume - Distribute based on product volumes (future)
  bbox   - Distribute based on bounding box volumes (future)

Use --dry-run flag to preview the price distribution without making changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCRMSpreadPrice(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runCRMSpreadPrice() error {
	// Validate parameters
	if err := bitrix.ValidateDealID(spreadDealID); err != nil {
		return fmt.Errorf("invalid deal ID: %v", err)
	}

	// Currently only count method is supported
	if spreadMethod != "count" {
		return fmt.Errorf("method '%s' is not supported yet. Only 'count' is currently available", spreadMethod)
	}

	// Get webhook URL from config
	webhookURL := viper.GetString("bitrix_webhook_url")
	if webhookURL == "" {
		return fmt.Errorf("bitrix_webhook_url not configured. Please set it in ~/.3mfanalyzer config")
	}

	if spreadDryRun {
		fmt.Printf("[DRY RUN] Processing deal %s with method '%s'...\n", spreadDealID, spreadMethod)
	} else {
		fmt.Printf("Processing deal %s with method '%s'...\n", spreadDealID, spreadMethod)
	}

	// Create Bitrix24 client
	client := bitrix.NewClient(webhookURL)

	// Get deal information with amount
	if spreadDryRun {
		fmt.Printf("[DRY RUN] Getting deal information...\n")
	} else {
		fmt.Println("Getting deal information...")
	}
	
	deal, err := client.GetDealWithAmount(spreadDealID)
	if err != nil {
		return fmt.Errorf("failed to get deal information: %v", err)
	}

	if deal.Opportunity <= 0 {
		return fmt.Errorf("deal amount is zero or negative: %.2f", deal.Opportunity)
	}

	if spreadDryRun {
		fmt.Printf("[DRY RUN] Deal amount: %.2f %s\n", deal.Opportunity, deal.CurrencyID)
	} else {
		fmt.Printf("Deal amount: %.2f %s\n", deal.Opportunity, deal.CurrencyID)
	}

	// Get existing products in deal
	if spreadDryRun {
		fmt.Printf("[DRY RUN] Getting products in deal...\n")
	} else {
		fmt.Println("Getting products in deal...")
	}
	
	products, err := client.GetExistingProductRows(spreadDealID)
	if err != nil {
		return fmt.Errorf("failed to get deal products: %v", err)
	}

	if len(products) == 0 {
		return fmt.Errorf("no products found in deal %s", spreadDealID)
	}

	if spreadDryRun {
		fmt.Printf("[DRY RUN] Found %d products in deal\n", len(products))
	} else {
		fmt.Printf("Found %d products in deal\n", len(products))
	}

	// Spread prices based on method
	switch spreadMethod {
	case "count":
		err = client.SpreadPriceByCount(spreadDealID, deal.Opportunity, deal.CurrencyID, spreadDryRun)
		if err != nil {
			return fmt.Errorf("failed to spread prices by count: %v", err)
		}
	default:
		return fmt.Errorf("unsupported method: %s", spreadMethod)
	}

	if spreadDryRun {
		fmt.Printf("[DRY RUN] Price distribution completed\n")
	} else {
		fmt.Printf("Successfully updated product prices\n")
	}

	return nil
}

func init() {
	crmSpreadPriceCmd.Flags().StringVar(&spreadDealID, "deal-id", "", "Bitrix24 deal ID (required)")
	crmSpreadPriceCmd.Flags().StringVar(&spreadMethod, "method", "count", "Distribution method: count, volume, bbox")
	crmSpreadPriceCmd.Flags().BoolVar(&spreadDryRun, "dry-run", false, "Preview price distribution without making changes")

	crmSpreadPriceCmd.MarkFlagRequired("deal-id")

	rootCmd.AddCommand(crmSpreadPriceCmd)
}