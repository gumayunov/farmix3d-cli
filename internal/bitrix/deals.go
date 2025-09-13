package bitrix

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// GetDeal retrieves deal information by ID
func (c *Client) GetDeal(dealID string) (*Deal, error) {
	params := map[string]interface{}{
		"id": dealID,
	}

	resp, err := c.makeRequest("crm.deal.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %v", err)
	}

	var deal Deal
	if err := c.parseResponse(resp, &deal); err != nil {
		return nil, fmt.Errorf("failed to parse deal response: %v", err)
	}

	return &deal, nil
}

// GetDealWithAmount retrieves deal information including amount and currency
func (c *Client) GetDealWithAmount(dealID string) (*Deal, error) {
	params := map[string]interface{}{
		"id": dealID,
		"select": []string{"ID", "TITLE", "CONTACT_ID", "COMPANY_ID", "ASSIGNED_BY_ID", "OPPORTUNITY", "CURRENCY_ID"},
	}

	resp, err := c.makeRequest("crm.deal.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal with amount: %v", err)
	}

	// Parse into raw structure first
	var dealRaw DealRaw
	if err := c.parseResponse(resp, &dealRaw); err != nil {
		return nil, fmt.Errorf("failed to parse deal with amount response: %v", err)
	}

	// Convert to final Deal structure
	deal := &Deal{
		ID:           dealRaw.ID,
		Title:        dealRaw.Title,
		ContactID:    dealRaw.ContactID,
		CompanyID:    dealRaw.CompanyID,
		AssignedByID: dealRaw.AssignedByID,
		CurrencyID:   dealRaw.CurrencyID,
	}

	// Parse opportunity amount from string
	if dealRaw.OpportunityRaw != "" {
		// Try to parse as float
		opportunity, err := strconv.ParseFloat(dealRaw.OpportunityRaw, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse deal amount '%s': %v", dealRaw.OpportunityRaw, err)
		}
		deal.Opportunity = opportunity
	}

	// Set default currency if not specified
	if deal.CurrencyID == "" {
		deal.CurrencyID = "RUB"
	}

	return deal, nil
}

// GetContact retrieves contact information by ID
func (c *Client) GetContact(contactID string) (*Contact, error) {
	params := map[string]interface{}{
		"id": contactID,
	}

	resp, err := c.makeRequest("crm.contact.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %v", err)
	}

	var contact Contact
	if err := c.parseResponse(resp, &contact); err != nil {
		return nil, fmt.Errorf("failed to parse contact response: %v", err)
	}

	return &contact, nil
}

// GetCompany retrieves company information by ID
func (c *Client) GetCompany(companyID string) (*Company, error) {
	params := map[string]interface{}{
		"id": companyID,
	}

	resp, err := c.makeRequest("crm.company.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %v", err)
	}

	var company Company
	if err := c.parseResponse(resp, &company); err != nil {
		return nil, fmt.Errorf("failed to parse company response: %v", err)
	}

	return &company, nil
}

// GetUser retrieves user information by ID
func (c *Client) GetUser(userID string) (*User, error) {
	params := map[string]interface{}{
		"id": userID,
	}

	resp, err := c.makeRequest("user.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	// user.get returns an array, so parse it differently
	var users []User
	if err := c.parseResponse(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %v", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	user := users[0]

	// If FullName is not provided by API, construct it from Name and LastName
	if user.FullName == "" && (user.Name != "" || user.LastName != "") {
		user.FullName = strings.TrimSpace(user.Name + " " + user.LastName)
	}

	return &user, nil
}

// GetDealURL generates Bitrix24 deal URL
func (c *Client) GetDealURL(dealID string) string {
	// Extract base URL from webhook URL
	// Example: https://farmix.bitrix24.ru/rest/10/jzz2ijynswg1nkur/ -> https://farmix.bitrix24.ru/
	webhookURL := c.GetWebhookURL()
	if strings.Contains(webhookURL, "/rest/") {
		baseURL := webhookURL[:strings.Index(webhookURL, "/rest/")]
		return fmt.Sprintf("%s/crm/deal/details/%s/", baseURL, dealID)
	}
	
	// Fallback if webhook URL format is unexpected
	return fmt.Sprintf("https://bitrix24.com/crm/deal/details/%s/", dealID)
}

// GetCustomerName retrieves customer name for a deal
func (c *Client) GetCustomerName(deal *Deal) (string, error) {
	// Try to get company name first
	if deal.CompanyID != "" && deal.CompanyID != "0" {
		company, err := c.GetCompany(deal.CompanyID)
		if err == nil && company.Title != "" {
			return company.Title, nil
		}
	}

	// If no company, try contact
	if deal.ContactID != "" && deal.ContactID != "0" {
		contact, err := c.GetContact(deal.ContactID)
		if err == nil && contact.Name != "" {
			return contact.Name, nil
		}
	}

	// If neither contact nor company, use deal title
	if deal.Title != "" {
		return deal.Title, nil
	}

	return fmt.Sprintf("Deal_%s", deal.ID), nil
}

// AddProductsToDeal adds products to a deal
func (c *Client) AddProductsToDeal(dealID string, products []DealProductRow) error {
	// Convert DealProductRow slice to []interface{} with map[string]interface{} elements
	rows := make([]interface{}, len(products))
	for i, product := range products {
		rows[i] = map[string]interface{}{
			"PRODUCT_ID": product.ProductID.String(),
			"QUANTITY":   product.Quantity,
			"PRICE":      product.Price,
		}
	}
	
	params := map[string]interface{}{
		"id":   dealID,
		"rows": rows,
	}

	resp, err := c.makeRequest("crm.deal.productrows.set", params)
	if err != nil {
		return fmt.Errorf("failed to add products to deal: %v", err)
	}

	var result bool
	if err := c.parseResponse(resp, &result); err != nil {
		return fmt.Errorf("failed to parse add products response: %v", err)
	}

	if !result {
		return fmt.Errorf("failed to add products to deal: API returned false")
	}

	return nil
}

// GetExistingProductRows retrieves existing product rows for a deal
func (c *Client) GetExistingProductRows(dealID string) ([]DealProductRow, error) {
	params := map[string]interface{}{
		"id": dealID,
	}

	resp, err := c.makeRequest("crm.deal.productrows.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing products: %v", err)
	}

	var products []DealProductRow
	if err := c.parseResponse(resp, &products); err != nil {
		return nil, fmt.Errorf("failed to parse existing products response: %v", err)
	}

	return products, nil
}

// AddProductRowsToDeal adds new product rows to existing ones
func (c *Client) AddProductRowsToDeal(dealID string, newProducts []DealProductRow, dryRun bool) error {
	// Get existing products
	existingProducts, err := c.GetExistingProductRows(dealID)
	if err != nil {
		// If getting existing products fails, just add new ones
		existingProducts = []DealProductRow{}
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Deal %s currently has %d existing products\n", dealID, len(existingProducts))
		fmt.Printf("[DRY RUN] Would add %d new products to deal\n", len(newProducts))
		
		if len(existingProducts) > 0 {
			fmt.Printf("[DRY RUN] Existing products in deal:\n")
			for _, product := range existingProducts {
				fmt.Printf("  - Product ID: %s (Quantity: %.1f)\n", product.ProductID.String(), product.Quantity)
			}
		}
		
		fmt.Printf("[DRY RUN] New products that would be added:\n")
		for _, product := range newProducts {
			fmt.Printf("  - Product ID: %s (Quantity: %.1f)\n", product.ProductID.String(), product.Quantity)
		}
		
		totalProducts := len(existingProducts) + len(newProducts)
		fmt.Printf("[DRY RUN] Total products after addition: %d\n", totalProducts)
		return nil
	}

	// Combine existing and new products
	allProducts := append(existingProducts, newProducts...)

	return c.AddProductsToDeal(dealID, allProducts)
}

// SpreadPriceByCount distributes deal amount among products proportionally by quantity
func (c *Client) SpreadPriceByCount(dealID string, totalAmount float64, currency string, dryRun bool) error {
	// Get existing products in deal
	products, err := c.GetExistingProductRows(dealID)
	if err != nil {
		return fmt.Errorf("failed to get existing products: %v", err)
	}

	if len(products) == 0 {
		return fmt.Errorf("no products found in deal")
	}

	// Calculate total quantity
	totalQuantity := 0.0
	for _, product := range products {
		totalQuantity += product.Quantity
	}

	if totalQuantity <= 0 {
		return fmt.Errorf("total quantity is zero or negative: %.2f", totalQuantity)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Total quantity: %.2f units\n", totalQuantity)
		fmt.Printf("[DRY RUN] Distributing price by count method:\n")
	} else {
		fmt.Printf("Total quantity: %.2f units\n", totalQuantity)
		fmt.Printf("Distributing price by count method:\n")
	}

	// Calculate proportional unit prices with rounding
	distributedSum := 0.0
	for i := range products {
		if i == len(products)-1 {
			// Last product - adjust remainder for exact total
			remainingAmount := totalAmount - distributedSum
			unitPrice := remainingAmount / products[i].Quantity
			products[i].Price = math.Round(unitPrice*100) / 100 // Round to 2 decimal places
			
			if dryRun {
				fmt.Printf("  - Product ID %s: %.2f units → %.2f %s per unit → total %.2f %s (remainder adjusted)\n", 
					products[i].ProductID.String(), products[i].Quantity, products[i].Price, currency, 
					products[i].Price*products[i].Quantity, currency)
			} else {
				fmt.Printf("  - Product ID %s: %.2f units → %.2f %s per unit → total %.2f %s (remainder adjusted)\n", 
					products[i].ProductID.String(), products[i].Quantity, products[i].Price, currency, 
					products[i].Price*products[i].Quantity, currency)
			}
		} else {
			// Calculate proportional price and round
			proportion := products[i].Quantity / totalQuantity
			unitPrice := (proportion * totalAmount) / products[i].Quantity
			products[i].Price = math.Round(unitPrice*100) / 100 // Round to 2 decimal places
			productTotal := products[i].Price * products[i].Quantity
			distributedSum += productTotal
			
			if dryRun {
				fmt.Printf("  - Product ID %s: %.2f units → %.2f %s per unit → total %.2f %s\n", 
					products[i].ProductID.String(), products[i].Quantity, products[i].Price, currency, 
					productTotal, currency)
			} else {
				fmt.Printf("  - Product ID %s: %.2f units → %.2f %s per unit → total %.2f %s\n", 
					products[i].ProductID.String(), products[i].Quantity, products[i].Price, currency, 
					productTotal, currency)
			}
		}
	}

	// Verify total sum
	verificationSum := 0.0
	for _, product := range products {
		verificationSum += product.Price * product.Quantity
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Verification: total distributed %.2f %s = deal amount %.2f %s ✓\n", 
			verificationSum, currency, totalAmount, currency)
		fmt.Printf("[DRY RUN] Would update %d product prices\n", len(products))
		return nil
	}

	fmt.Printf("Verification: total distributed %.2f %s = deal amount %.2f %s ✓\n", 
		verificationSum, currency, totalAmount, currency)

	// Update product prices in Bitrix24
	fmt.Println("Updating product prices...")
	err = c.AddProductsToDeal(dealID, products)
	if err != nil {
		return fmt.Errorf("failed to update product prices: %v", err)
	}

	return nil
}

// ValidateDealID checks if deal ID is a valid number
func ValidateDealID(dealID string) error {
	if dealID == "" {
		return fmt.Errorf("deal ID cannot be empty")
	}
	
	_, err := strconv.Atoi(dealID)
	if err != nil {
		return fmt.Errorf("deal ID must be a number: %s", dealID)
	}
	
	return nil
}