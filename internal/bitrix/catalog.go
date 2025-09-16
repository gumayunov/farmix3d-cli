package bitrix

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// FileInfo represents a 3D file with its directory path information
type FileInfo struct {
	FileName string // name of the file (e.g., "2x_gear.stl")
	DirPath  string // relative directory path from base directory (e.g., "arms/mechanisms")
}

// PRODUCT_NAME_PREFIX is the prefix added to all product names created from 3D files
const PRODUCT_NAME_PREFIX = "Изделие "

// COMPANIES_FOLDER_NAME is the name of the folder where all customer companies are stored
const COMPANIES_FOLDER_NAME = "Компании"

// ParseFileName extracts quantity and clean name from 3D model filename
// Supports formats: "2x_part.stl", "3х_gear.step", "1x SMA Hear.stl", "simple.stl"
// Returns clean name without extension and quantity (default 1.0)
func ParseFileName(fileName string) (cleanName string, quantity float64) {
	// Default quantity
	quantity = 1.0
	
	// Remove file extension
	nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	
	// Regex to match quantity prefixes: 2x_, 3х_, 1x SMA, 10x_, etc.
	// Supports both 'x' and 'х' (cyrillic) with underscore or space separator
	quantityRegex := regexp.MustCompile(`^(\d+)[xх][_\s](.+)$`)
	matches := quantityRegex.FindStringSubmatch(nameWithoutExt)
	
	if len(matches) == 3 {
		// Found quantity prefix
		if parsedQuantity, err := strconv.ParseFloat(matches[1], 64); err == nil && parsedQuantity > 0 {
			quantity = parsedQuantity
			cleanName = matches[2]
		} else {
			// Invalid quantity (0 or negative), treat as regular filename
			cleanName = nameWithoutExt
		}
	} else {
		// No quantity prefix, use whole name without extension
		cleanName = nameWithoutExt
	}
	
	return cleanName, quantity
}

// formatDirPrefix converts directory path to prefix for product name
// Example: "" -> ""
// Example: "parts" -> "parts "
// Example: "arms/mechanisms" -> "arms.mechanisms "
func formatDirPrefix(dirPath string) string {
	if dirPath == "" {
		return ""
	}
	return strings.ReplaceAll(dirPath, "/", ".") + " "
}

// FormatProductName formats clean name into product name with quotes and quantity suffix
// Example: "bracket" -> "Изделие \"bracket\"" (quantity 1.0)
// Example: "bracket" -> "Изделие \"bracket Q4\"" (quantity 4.0)
func FormatProductName(cleanName string, quantity float64) string {
	if quantity > 1.0 {
		return PRODUCT_NAME_PREFIX + "\"" + cleanName + " Q" + fmt.Sprintf("%.0f", quantity) + "\""
	}
	return PRODUCT_NAME_PREFIX + "\"" + cleanName + "\""
}

// FormatProductNameWithDir formats clean name with directory prefix into product name
// Example: ("bracket", "", 1.0) -> "Изделие \"bracket\""
// Example: ("bracket", "parts", 1.0) -> "Изделие \"parts bracket\""
// Example: ("gear", "arms/mechanisms", 2.0) -> "Изделие \"arms.mechanisms gear Q2\""
func FormatProductNameWithDir(cleanName string, dirPath string, quantity float64) string {
	dirPrefix := formatDirPrefix(dirPath)
	nameWithDir := dirPrefix + cleanName

	if quantity > 1.0 {
		return PRODUCT_NAME_PREFIX + "\"" + nameWithDir + " Q" + fmt.Sprintf("%.0f", quantity) + "\""
	}
	return PRODUCT_NAME_PREFIX + "\"" + nameWithDir + "\""
}

// ListSections retrieves catalog sections
func (c *Client) ListSections(catalogID string) ([]ProductSection, error) {
	params := map[string]interface{}{
		"select": []string{"ID", "NAME", "SECTION_ID"},
		"filter": map[string]interface{}{
			"iblockId": catalogID,
		},
	}

	resp, err := c.makeRequest("catalog.section.list", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list sections: %v", err)
	}

	// First, let's see what the raw response looks like
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse the generic response first
	var bitrixResp BitrixResponse
	if err := json.Unmarshal(body, &bitrixResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if bitrixResp.Error != nil {
		return nil, fmt.Errorf("Bitrix24 API error %s: %s", bitrixResp.Error.ErrorCode, bitrixResp.Error.ErrorDescription)
	}
	
	// Log the raw result for debugging (remove this line in production)
	// resultJSON, _ := json.Marshal(bitrixResp.Result)
	// fmt.Printf("DEBUG: catalog.section.list result: %s\n", string(resultJSON))
	
	// Parse the result object which contains 'sections' field
	type ListResult struct {
		Sections []ProductSection `json:"sections"`
	}
	
	var listResult ListResult
	resultBytes, err := json.Marshal(bitrixResp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %v", err)
	}
	
	if err := json.Unmarshal(resultBytes, &listResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result into sections: %v", err)
	}
	
	return listResult.Sections, nil
}

// FindSectionByName finds a section by name (case-insensitive)
func (c *Client) FindSectionByName(sections []ProductSection, name string, parentID string) *ProductSection {
	for _, section := range sections {
		// Check if section name matches
		if !strings.EqualFold(section.Name, name) {
			continue
		}
		
		// Check parent ID
		if parentID == "" {
			// Looking for root section (parentID should be null)
			if section.ParentID == nil {
				return &section
			}
		} else {
			// Looking for section with specific parent
			if section.ParentID != nil && fmt.Sprintf("%d", *section.ParentID) == parentID {
				return &section
			}
		}
	}
	return nil
}

// CreateSection creates a new catalog section
func (c *Client) CreateSection(name string, parentID string, catalogID string) (string, error) {
	fields := map[string]interface{}{
		"iblockId": catalogID, // Keep as string for now
		"name":     name,
	}
	
	if parentID != "" {
		fields["iblockSectionId"] = parentID
	}

	params := map[string]interface{}{
		"fields": fields,
	}

	resp, err := c.makeRequest("catalog.section.add", params)
	if err != nil {
		return "", fmt.Errorf("failed to create section: %v", err)
	}

	// Debug: let's see what the API actually returns
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	// fmt.Printf("DEBUG: catalog.section.add response: %s\n", string(body))
	
	// Parse the generic response first
	var bitrixResp BitrixResponse
	if err := json.Unmarshal(body, &bitrixResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if bitrixResp.Error != nil {
		return "", fmt.Errorf("Bitrix24 API error %s: %s", bitrixResp.Error.ErrorCode, bitrixResp.Error.ErrorDescription)
	}
	
	// Parse the result which contains a 'section' object
	type CreateSectionResult struct {
		Section struct {
			ID int `json:"id"`
		} `json:"section"`
	}
	
	var createResult CreateSectionResult
	resultBytes, err := json.Marshal(bitrixResp.Result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}
	
	if err := json.Unmarshal(resultBytes, &createResult); err != nil {
		return "", fmt.Errorf("failed to unmarshal result: %v", err)
	}
	
	return fmt.Sprintf("%d", createResult.Section.ID), nil
}

// CreateProduct creates a new catalog product
func (c *Client) CreateProduct(name string, sectionID string, catalogID string) (string, error) {
	fields := map[string]interface{}{
		"name":       name,
		"iblockId":   catalogID,
		"iblockSectionId": sectionID,
	}

	params := map[string]interface{}{
		"fields": fields,
	}

	resp, err := c.makeRequest("catalog.product.add", params)
	if err != nil {
		return "", fmt.Errorf("failed to create product: %v", err)
	}

	// Debug: let's see what the API actually returns
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	// fmt.Printf("DEBUG: catalog.product.add response: %s\n", string(body))
	
	// Parse the generic response first
	var bitrixResp BitrixResponse
	if err := json.Unmarshal(body, &bitrixResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if bitrixResp.Error != nil {
		return "", fmt.Errorf("Bitrix24 API error %s: %s", bitrixResp.Error.ErrorCode, bitrixResp.Error.ErrorDescription)
	}
	
	// Convert result to string (it should be the product ID)
	if productID, ok := bitrixResp.Result.(string); ok {
		return productID, nil
	} else if productIDFloat, ok := bitrixResp.Result.(float64); ok {
		return fmt.Sprintf("%.0f", productIDFloat), nil
	} else {
		// Maybe it's a complex object - API returns 'element' not 'product'
		type CreateProductResult struct {
			Element struct {
				ID int `json:"id"`
			} `json:"element"`
		}
		
		var createResult CreateProductResult
		resultBytes, err := json.Marshal(bitrixResp.Result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %v", err)
		}
		
		if err := json.Unmarshal(resultBytes, &createResult); err != nil {
			return "", fmt.Errorf("unexpected result format: %T", bitrixResp.Result)
		}
		
		return fmt.Sprintf("%d", createResult.Element.ID), nil
	}
}

// ListProducts retrieves catalog products in a section
func (c *Client) ListProducts(catalogID string, sectionID string) ([]Product, error) {
	params := map[string]interface{}{
		"select": []string{"id", "name", "iblockSectionId", "iblockId"},
		"filter": map[string]interface{}{
			"iblockId": catalogID,
		},
	}
	
	// Add section filter if specified
	if sectionID != "" {
		params["filter"].(map[string]interface{})["iblockSectionId"] = sectionID
	}

	resp, err := c.makeRequest("catalog.product.list", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %v", err)
	}

	// Parse response like we do for sections
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	var bitrixResp BitrixResponse
	if err := json.Unmarshal(body, &bitrixResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if bitrixResp.Error != nil {
		return nil, fmt.Errorf("Bitrix24 API error %s: %s", bitrixResp.Error.ErrorCode, bitrixResp.Error.ErrorDescription)
	}
	
	// Parse the result object which contains 'products' field
	type ListProductResult struct {
		Products []Product `json:"products"`
	}
	
	var listResult ListProductResult
	resultBytes, err := json.Marshal(bitrixResp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %v", err)
	}
	
	if err := json.Unmarshal(resultBytes, &listResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result into products: %v", err)
	}
	
	return listResult.Products, nil
}

// FindProductByName finds a product by name in the given products list
func (c *Client) FindProductByName(products []Product, name string) *Product {
	for _, product := range products {
		if strings.EqualFold(product.Name, name) {
			return &product
		}
	}
	return nil
}

// EnsureCompaniesFolder ensures "Компании" folder exists in catalog root, creates if not
func (c *Client) EnsureCompaniesFolder(catalogID string, dryRun bool) (string, error) {
	sections, err := c.ListSections(catalogID)
	if err != nil {
		return "", fmt.Errorf("failed to list sections: %v", err)
	}

	// Look for companies folder in root (parentID = "")
	if section := c.FindSectionByName(sections, COMPANIES_FOLDER_NAME, ""); section != nil {
		if dryRun {
			fmt.Printf("[DRY RUN] Companies folder '%s' exists (ID: %d)\n", COMPANIES_FOLDER_NAME, section.ID)
		}
		return fmt.Sprintf("%d", section.ID), nil
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Companies folder '%s' does not exist - would create new folder\n", COMPANIES_FOLDER_NAME)
		// Return a placeholder ID for dry run
		return "dry-run-companies-folder-id", nil
	}

	// Create companies folder in root
	sectionID, err := c.CreateSection(COMPANIES_FOLDER_NAME, "", catalogID)
	if err != nil {
		return "", fmt.Errorf("failed to create companies folder: %v", err)
	}

	return sectionID, nil
}

// EnsureCustomerSection ensures customer section exists, creates if not
func (c *Client) EnsureCustomerSection(customerName string, catalogID string, dryRun bool) (string, error) {
	// First, ensure companies folder exists
	companiesFolderID, err := c.EnsureCompaniesFolder(catalogID, dryRun)
	if err != nil {
		return "", fmt.Errorf("failed to ensure companies folder: %v", err)
	}

	sections, err := c.ListSections(catalogID)
	if err != nil {
		return "", fmt.Errorf("failed to list sections: %v", err)
	}

	// Look for customer section in companies folder
	if section := c.FindSectionByName(sections, customerName, companiesFolderID); section != nil {
		if dryRun {
			fmt.Printf("[DRY RUN] Customer section '%s' exists in companies folder (ID: %d)\n", customerName, section.ID)
		}
		return fmt.Sprintf("%d", section.ID), nil
	}

	// Also check in root for backward compatibility
	if section := c.FindSectionByName(sections, customerName, ""); section != nil {
		if dryRun {
			fmt.Printf("[DRY RUN] Customer section '%s' exists in root (ID: %d) - should migrate to companies folder\n", customerName, section.ID)
		}
		return fmt.Sprintf("%d", section.ID), nil
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Customer section '%s' does not exist - would create in companies folder\n", customerName)
		// Return a placeholder ID for dry run
		return "dry-run-customer-section-id", nil
	}

	// Create customer section in companies folder
	sectionID, err := c.CreateSection(customerName, companiesFolderID, catalogID)
	if err != nil {
		return "", fmt.Errorf("failed to create customer section: %v", err)
	}

	return sectionID, nil
}

// EnsureProjectSection ensures project section exists under customer, creates if not
func (c *Client) EnsureProjectSection(projectName, dealID, customerSectionID, catalogID string, dryRun bool) (string, error) {
	sectionName := fmt.Sprintf("%s - %s", projectName, dealID)
	
	sections, err := c.ListSections(catalogID)
	if err != nil {
		return "", fmt.Errorf("failed to list sections: %v", err)
	}

	// Look for project section under customer
	if section := c.FindSectionByName(sections, sectionName, customerSectionID); section != nil {
		if dryRun {
			fmt.Printf("[DRY RUN] Project section '%s' exists (ID: %d)\n", sectionName, section.ID)
		}
		return fmt.Sprintf("%d", section.ID), nil
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Project section '%s' does not exist - would create under customer section ID %s\n", sectionName, customerSectionID)
		// Return a placeholder ID for dry run
		return "dry-run-project-section-id", nil
	}

	// Create project section under customer
	sectionID, err := c.CreateSection(sectionName, customerSectionID, catalogID)
	if err != nil {
		return "", fmt.Errorf("failed to create project section: %v", err)
	}

	return sectionID, nil
}

// CreateProductsFrom3DFiles creates products for 3D model files (.stl and .step) in the specified section
func (c *Client) CreateProductsFrom3DFiles(files3D []FileInfo, sectionID string, catalogID string, dryRun bool) ([]ProductInfo, error) {
	// First, get existing products in the section
	if dryRun {
		fmt.Printf("[DRY RUN] Checking for existing products in section %s...\n", sectionID)
	} else {
		fmt.Println("Checking for existing products in section...")
	}
	
	existingProducts, err := c.ListProducts(catalogID, sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing products: %v", err)
	}
	
	if dryRun {
		fmt.Printf("[DRY RUN] Found %d existing products in section\n", len(existingProducts))
	} else {
		fmt.Printf("Found %d existing products in section\n", len(existingProducts))
	}
	
	var products []ProductInfo
	var createdCount int
	var skippedCount int
	
	for _, fileInfo := range files3D {
		// Parse filename to extract quantity and clean name
		cleanName, quantity := ParseFileName(fileInfo.FileName)
		productName := FormatProductNameWithDir(cleanName, fileInfo.DirPath, quantity)
		
		// Check if product already exists
		if existingProduct := c.FindProductByName(existingProducts, productName); existingProduct != nil {
			if dryRun {
				fmt.Printf("[DRY RUN] Product '%s' already exists (ID: %d) - would skip creation (quantity: %.0f)\n", productName, existingProduct.ID, quantity)
			} else {
				fmt.Printf("Product '%s' already exists (ID: %d), skipping creation (quantity: %.0f)\n", productName, existingProduct.ID, quantity)
			}
			products = append(products, ProductInfo{
				ID:       fmt.Sprintf("%d", existingProduct.ID),
				Quantity: quantity,
			})
			skippedCount++
			continue
		}
		
		if dryRun {
			fmt.Printf("[DRY RUN] Product '%s' does not exist - would create new product (quantity: %.0f)\n", productName, quantity)
			// Use placeholder ID for dry run
			products = append(products, ProductInfo{
				ID:       fmt.Sprintf("dry-run-product-%d", createdCount+1),
				Quantity: quantity,
			})
			createdCount++
		} else {
			// Create new product
			fmt.Printf("Creating product '%s' (quantity: %.0f)...\n", productName, quantity)
			productID, err := c.CreateProduct(productName, sectionID, catalogID)
			if err != nil {
				return nil, fmt.Errorf("failed to create product '%s': %v", productName, err)
			}
			
			products = append(products, ProductInfo{
				ID:       productID,
				Quantity: quantity,
			})
			createdCount++
		}
	}
	
	if dryRun {
		fmt.Printf("[DRY RUN] Products analysis: %d would be created, %d already exist\n", createdCount, skippedCount)
	} else {
		fmt.Printf("Products processed: %d created, %d skipped (already existed)\n", createdCount, skippedCount)
	}
	return products, nil
}

// CreateDealProductRows converts ProductInfo to deal product rows
func CreateDealProductRows(products []ProductInfo) []DealProductRow {
	var rows []DealProductRow
	
	for _, product := range products {
		row := DealProductRow{
			ProductID: ProductIDString(product.ID),
			Quantity:  product.Quantity,
			Price:     0.0, // Default price, can be set later in Bitrix24
		}
		rows = append(rows, row)
	}
	
	return rows
}