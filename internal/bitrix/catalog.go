package bitrix

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

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

// EnsureCustomerSection ensures customer section exists, creates if not
func (c *Client) EnsureCustomerSection(customerName string, catalogID string, dryRun bool) (string, error) {
	sections, err := c.ListSections(catalogID)
	if err != nil {
		return "", fmt.Errorf("failed to list sections: %v", err)
	}

	// Look for customer section in root (parentID = "")
	if section := c.FindSectionByName(sections, customerName, ""); section != nil {
		if dryRun {
			fmt.Printf("[DRY RUN] Customer section '%s' exists (ID: %d)\n", customerName, section.ID)
		}
		return fmt.Sprintf("%d", section.ID), nil
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Customer section '%s' does not exist - would create new section\n", customerName)
		// Return a placeholder ID for dry run
		return "dry-run-customer-section-id", nil
	}

	// Create customer section in root
	sectionID, err := c.CreateSection(customerName, "", catalogID)
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

// CreateProductsFromSTLFiles creates products for STL files in the specified section
func (c *Client) CreateProductsFromSTLFiles(stlFiles []string, sectionID string, catalogID string, dryRun bool) ([]string, error) {
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
	
	var productIDs []string
	var createdCount int
	var skippedCount int
	
	for _, fileName := range stlFiles {
		// Remove .stl extension from product name
		productName := strings.TrimSuffix(fileName, ".stl")
		
		// Check if product already exists
		if existingProduct := c.FindProductByName(existingProducts, productName); existingProduct != nil {
			if dryRun {
				fmt.Printf("[DRY RUN] Product '%s' already exists (ID: %d) - would skip creation\n", productName, existingProduct.ID)
			} else {
				fmt.Printf("Product '%s' already exists (ID: %d), skipping creation\n", productName, existingProduct.ID)
			}
			productIDs = append(productIDs, fmt.Sprintf("%d", existingProduct.ID))
			skippedCount++
			continue
		}
		
		if dryRun {
			fmt.Printf("[DRY RUN] Product '%s' does not exist - would create new product\n", productName)
			// Use placeholder ID for dry run
			productIDs = append(productIDs, fmt.Sprintf("dry-run-product-%d", createdCount+1))
			createdCount++
		} else {
			// Create new product
			fmt.Printf("Creating product '%s'...\n", productName)
			productID, err := c.CreateProduct(productName, sectionID, catalogID)
			if err != nil {
				return nil, fmt.Errorf("failed to create product '%s': %v", productName, err)
			}
			
			productIDs = append(productIDs, productID)
			createdCount++
		}
	}
	
	if dryRun {
		fmt.Printf("[DRY RUN] Products analysis: %d would be created, %d already exist\n", createdCount, skippedCount)
	} else {
		fmt.Printf("Products processed: %d created, %d skipped (already existed)\n", createdCount, skippedCount)
	}
	return productIDs, nil
}

// CreateDealProductRows converts product IDs to deal product rows
func CreateDealProductRows(productIDs []string) []DealProductRow {
	var rows []DealProductRow
	
	for _, productID := range productIDs {
		row := DealProductRow{
			ProductID: productID,
			Quantity:  1.0,
			Price:     0.0, // Default price, can be set later in Bitrix24
		}
		rows = append(rows, row)
	}
	
	return rows
}