package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"3mfanalyzer/internal/bitrix"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dealID      string
	projectName string
	stlDir      string
	dryRun      bool
)

var crmAddItemsCmd = &cobra.Command{
	Use:   "crm-add-items",
	Short: "Add STL files as products to Bitrix24 deal",
	Long: `Add STL files from a directory as products to a Bitrix24 deal.
	
This command will:
1. Get deal information from Bitrix24
2. Create/find customer folder in catalog
3. Create project subfolder with name "project - deal_id"
4. Create products for each STL file
5. Add products to the deal

Use --dry-run flag to preview what would be created without making changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCRMAddItems(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runCRMAddItems() error {
	// Validate parameters
	if err := bitrix.ValidateDealID(dealID); err != nil {
		return fmt.Errorf("invalid deal ID: %v", err)
	}

	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if stlDir == "" {
		return fmt.Errorf("STL directory cannot be empty")
	}

	// Check if STL directory exists
	if _, err := os.Stat(stlDir); os.IsNotExist(err) {
		return fmt.Errorf("STL directory does not exist: %s", stlDir)
	}

	// Get webhook URL from config
	webhookURL := viper.GetString("bitrix_webhook_url")
	if webhookURL == "" {
		return fmt.Errorf("bitrix_webhook_url not configured. Please set it in ~/.3mfanalyzer config")
	}

	// Get catalog ID from config
	catalogID := viper.GetString("catalog_id")
	if catalogID == "" {
		return fmt.Errorf("catalog_id not configured. Please set it in ~/.3mfanalyzer config")
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Processing deal %s with project '%s'...\n", dealID, projectName)
	} else {
		fmt.Printf("Processing deal %s with project '%s'...\n", dealID, projectName)
	}

	// Create Bitrix24 client
	client := bitrix.NewClient(webhookURL)

	// Get deal information
	fmt.Println("Getting deal information...")
	deal, err := client.GetDeal(dealID)
	if err != nil {
		return fmt.Errorf("failed to get deal information: %v", err)
	}

	// Get customer name
	fmt.Println("Getting customer information...")
	customerName, err := client.GetCustomerName(deal)
	if err != nil {
		return fmt.Errorf("failed to get customer name: %v", err)
	}
	fmt.Printf("Customer: %s\n", customerName)

	// Ensure customer section exists
	if dryRun {
		fmt.Printf("[DRY RUN] Checking customer folder '%s'...\n", customerName)
	} else {
		fmt.Printf("Ensuring customer folder '%s' exists...\n", customerName)
	}
	customerSectionID, err := client.EnsureCustomerSection(customerName, catalogID, dryRun)
	if err != nil {
		return fmt.Errorf("failed to ensure customer section: %v", err)
	}

	// Ensure project section exists
	if dryRun {
		fmt.Printf("[DRY RUN] Checking project folder '%s - %s'...\n", projectName, dealID)
	} else {
		fmt.Printf("Ensuring project folder '%s - %s' exists...\n", projectName, dealID)
	}
	projectSectionID, err := client.EnsureProjectSection(projectName, dealID, customerSectionID, catalogID, dryRun)
	if err != nil {
		return fmt.Errorf("failed to ensure project section: %v", err)
	}

	// Find STL files
	fmt.Printf("Scanning for STL files in %s...\n", stlDir)
	stlFiles, err := findSTLFiles(stlDir)
	if err != nil {
		return fmt.Errorf("failed to find STL files: %v", err)
	}

	if len(stlFiles) == 0 {
		return fmt.Errorf("no STL files found in directory: %s", stlDir)
	}

	fmt.Printf("Found %d STL files\n", len(stlFiles))

	// Create products for STL files
	if dryRun {
		fmt.Printf("[DRY RUN] Analyzing products that would be created...\n")
	} else {
		fmt.Println("Creating products in catalog...")
	}
	productIDs, err := client.CreateProductsFromSTLFiles(stlFiles, projectSectionID, catalogID, dryRun)
	if err != nil {
		return fmt.Errorf("failed to create products: %v", err)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would process %d products\n", len(productIDs))
	} else {
		fmt.Printf("Created %d products\n", len(productIDs))
	}

	// Add products to deal
	if dryRun {
		fmt.Printf("[DRY RUN] Checking what products would be added to deal...\n")
	} else {
		fmt.Println("Adding products to deal...")
	}
	productRows := bitrix.CreateDealProductRows(productIDs)
	err = client.AddProductRowsToDeal(dealID, productRows, dryRun)
	if err != nil {
		return fmt.Errorf("failed to add products to deal: %v", err)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would add %d products to deal %s\n", len(productIDs), dealID)
	} else {
		fmt.Printf("Successfully added %d products to deal %s\n", len(productIDs), dealID)
	}
	if dryRun {
		fmt.Println("[DRY RUN] Products that would be processed:")
	} else {
		fmt.Println("Products created:")
	}
	for i, fileName := range stlFiles {
		productName := strings.TrimSuffix(fileName, ".stl")
		if dryRun {
			fmt.Printf("  - %s (ID: %s)\n", productName, productIDs[i])
		} else {
			fmt.Printf("  - %s (ID: %s)\n", productName, productIDs[i])
		}
	}

	return nil
}

// findSTLFiles finds all STL files in the specified directory
func findSTLFiles(dir string) ([]string, error) {
	var stlFiles []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(strings.ToLower(d.Name()), ".stl") {
			stlFiles = append(stlFiles, d.Name())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return stlFiles, nil
}

func init() {
	crmAddItemsCmd.Flags().StringVar(&dealID, "deal-id", "", "Bitrix24 deal ID (required)")
	crmAddItemsCmd.Flags().StringVar(&projectName, "project-name", "", "Project name for folder creation (required)")
	crmAddItemsCmd.Flags().StringVar(&stlDir, "stl-dir", "", "Directory containing STL files (required)")
	crmAddItemsCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be created without making changes")

	crmAddItemsCmd.MarkFlagRequired("deal-id")
	crmAddItemsCmd.MarkFlagRequired("project-name")
	crmAddItemsCmd.MarkFlagRequired("stl-dir")

	rootCmd.AddCommand(crmAddItemsCmd)
}