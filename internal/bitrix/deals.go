package bitrix

import (
	"fmt"
	"strconv"
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
			"PRODUCT_ID": product.ProductID,
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
func (c *Client) AddProductRowsToDeal(dealID string, newProducts []DealProductRow) error {
	// Get existing products
	existingProducts, err := c.GetExistingProductRows(dealID)
	if err != nil {
		// If getting existing products fails, just add new ones
		existingProducts = []DealProductRow{}
	}

	// Combine existing and new products
	allProducts := append(existingProducts, newProducts...)

	return c.AddProductsToDeal(dealID, allProducts)
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