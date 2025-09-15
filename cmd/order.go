package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"farmix-cli/internal/bitrix"
	"farmix-cli/internal/formatter"
	"farmix-cli/internal/parser"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	orderDealID string
)

var orderCmd = &cobra.Command{
	Use:   "order [file]",
	Short: "Generate order and assignment Excel reports for a 3MF file with Bitrix24 integration",
	Long: `Generate detailed order and assignment reports with Bitrix24 CRM integration.
	
This command creates two Excel files:
- [filename]-order.xlsx    - Detailed order report with materials, costs, and CRM data
- [filename]-assignment.xlsx - Production assignment report for operators

The command integrates with Bitrix24 CRM to include:
- Deal information and responsible person
- Customer company name and contact details
- Direct links to CRM records`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runOrderCommand(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runOrderCommand(filePath string) error {
	// Validate deal ID
	if err := bitrix.ValidateDealID(orderDealID); err != nil {
		return fmt.Errorf("invalid deal ID: %v", err)
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".3mf") {
		return fmt.Errorf("file must have .3mf extension: %s", filePath)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get webhook URL from config
	webhookURL := viper.GetString("bitrix_webhook_url")
	if webhookURL == "" {
		return fmt.Errorf("bitrix_webhook_url not configured. Please set it in ~/.farmix-cli config")
	}

	fmt.Printf("Processing 3MF file: %s\n", filePath)
	fmt.Printf("Deal ID: %s\n", orderDealID)

	// Parse 3MF file
	data, err := parser.Parse3MF(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse 3MF file: %v", err)
	}

	// Create Bitrix24 client
	client := bitrix.NewClient(webhookURL)

	// Get deal information
	fmt.Println("Getting deal information from Bitrix24...")
	deal, err := client.GetDeal(orderDealID)
	if err != nil {
		return fmt.Errorf("failed to get deal information: %v", err)
	}

	// Get customer name
	customerName, err := client.GetCustomerName(deal)
	if err != nil {
		return fmt.Errorf("failed to get customer name: %v", err)
	}

	// Get assigned user name
	assignedUser, err := client.GetUser(deal.AssignedByID)
	if err != nil {
		return fmt.Errorf("failed to get assigned user information: %v", err)
	}

	fmt.Printf("Deal: %s\n", deal.Title)
	fmt.Printf("Customer: %s\n", customerName)
	fmt.Printf("Assigned to: %s\n", assignedUser.FullName)

	// Generate output file names
	baseName := strings.TrimSuffix(filepath.Base(filePath), ".3mf")
	orderPath := baseName + "-order.xlsx"
	assignmentPath := baseName + "-assignment.xlsx"

	// Create order report
	fmt.Printf("Creating order report: %s\n", orderPath)
	if err := formatter.FormatAsOrderExcel(data, deal, assignedUser, customerName, client, orderPath); err != nil {
		return fmt.Errorf("failed to create order report: %v", err)
	}

	// Create assignment report
	fmt.Printf("Creating assignment report: %s\n", assignmentPath)
	if err := formatter.FormatAsAssignmentExcel(data, deal, assignedUser, customerName, client, assignmentPath); err != nil {
		return fmt.Errorf("failed to create assignment report: %v", err)
	}

	fmt.Printf("Reports created successfully:\n")
	fmt.Printf("  - %s\n", orderPath)
	fmt.Printf("  - %s\n", assignmentPath)

	return nil
}

func init() {
	orderCmd.Flags().StringVar(&orderDealID, "deal-id", "", "Bitrix24 deal ID (required)")
	orderCmd.MarkFlagRequired("deal-id")
	rootCmd.AddCommand(orderCmd)
}