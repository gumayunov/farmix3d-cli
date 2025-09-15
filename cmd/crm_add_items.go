package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"farmix-cli/internal/bitrix"

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
	Short: "Add 3D model files (STL/STEP) as products to Bitrix24 deal",
	Long: `Add 3D model files from a directory as products to a Bitrix24 deal.
	
This command will:
1. Get deal information from Bitrix24
2. Create/find customer folder in catalog
3. Create project subfolder with name "project - deal_id"  
4. Create products for each 3D model file (.stl and .step)
5. Add products to the deal

Product names will have "Деталь " prefix and include the file extension.

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
		return fmt.Errorf("3D files directory cannot be empty")
	}

	// Check if 3D files directory exists
	if _, err := os.Stat(stlDir); os.IsNotExist(err) {
		return fmt.Errorf("3D files directory does not exist: %s", stlDir)
	}

	// Get webhook URL from config
	webhookURL := viper.GetString("bitrix_webhook_url")
	if webhookURL == "" {
		return fmt.Errorf("bitrix_webhook_url not configured. Please set it in ~/.farmix-cli config")
	}

	// Get catalog ID from config
	catalogID := viper.GetString("catalog_id")
	if catalogID == "" {
		return fmt.Errorf("catalog_id not configured. Please set it in ~/.farmix-cli config")
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

	// Find 3D files
	fmt.Printf("Scanning for 3D files (STL/STEP) in %s...\n", stlDir)
	files3D, err := find3DFiles(stlDir)
	if err != nil {
		return fmt.Errorf("failed to find 3D files: %v", err)
	}

	if len(files3D) == 0 {
		return fmt.Errorf("no 3D files (STL/STEP) found in directory: %s", stlDir)
	}

	fmt.Printf("Found %d 3D files\n", len(files3D))

	// Create products for 3D files
	if dryRun {
		fmt.Printf("[DRY RUN] Analyzing products that would be created...\n")
	} else {
		fmt.Println("Creating products in catalog...")
	}
	products, err := client.CreateProductsFrom3DFiles(files3D, projectSectionID, catalogID, dryRun)
	if err != nil {
		return fmt.Errorf("failed to create products: %v", err)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would process %d products\n", len(products))
	} else {
		fmt.Printf("Created %d products\n", len(products))
	}

	// Add products to deal
	if dryRun {
		fmt.Printf("[DRY RUN] Checking what products would be added to deal...\n")
	} else {
		fmt.Println("Adding products to deal...")
	}
	productRows := bitrix.CreateDealProductRows(products)
	err = client.AddProductRowsToDeal(dealID, productRows, dryRun)
	if err != nil {
		return fmt.Errorf("failed to add products to deal: %v", err)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would add %d products to deal %s\n", len(products), dealID)
	} else {
		fmt.Printf("Successfully added %d products to deal %s\n", len(products), dealID)
	}
	if dryRun {
		fmt.Println("[DRY RUN] Products that would be processed:")
	} else {
		fmt.Println("Products created:")
	}
	for i, fileName := range files3D {
		cleanName, quantity := bitrix.ParseFileName(fileName)
		productName := bitrix.FormatProductName(cleanName, quantity)
		if dryRun {
			fmt.Printf("  - %s (ID: %s, Quantity: %.0f)\n", productName, products[i].ID, quantity)
		} else {
			fmt.Printf("  - %s (ID: %s, Quantity: %.0f)\n", productName, products[i].ID, quantity)
		}
	}

	return nil
}

// find3DFiles finds all 3D model files (.stl and .step) in the specified directory
// Returns files sorted alphabetically by clean name (without quantity prefix)
func find3DFiles(dir string) ([]string, error) {
	var files3D []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		lowerName := strings.ToLower(d.Name())
		if strings.HasSuffix(lowerName, ".stl") || strings.HasSuffix(lowerName, ".step") {
			files3D = append(files3D, d.Name())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files alphabetically by clean name (without quantity prefix)
	sort.Slice(files3D, func(i, j int) bool {
		cleanI, _ := bitrix.ParseFileName(files3D[i])
		cleanJ, _ := bitrix.ParseFileName(files3D[j])
		return strings.ToLower(cleanI) < strings.ToLower(cleanJ)
	})

	return files3D, nil
}

func init() {
	crmAddItemsCmd.Flags().StringVar(&dealID, "deal-id", "", "Bitrix24 deal ID (required)")
	crmAddItemsCmd.Flags().StringVar(&projectName, "project-name", "", "Project name for folder creation (required)")
	crmAddItemsCmd.Flags().StringVar(&stlDir, "stl-dir", "", "Directory containing 3D model files (STL/STEP) (required)")
	crmAddItemsCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be created without making changes")

	crmAddItemsCmd.MarkFlagRequired("deal-id")
	crmAddItemsCmd.MarkFlagRequired("project-name")
	crmAddItemsCmd.MarkFlagRequired("stl-dir")

	rootCmd.AddCommand(crmAddItemsCmd)
}