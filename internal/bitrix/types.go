package bitrix

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ProductIDString is a custom type that can unmarshal both string and number IDs from JSON
type ProductIDString string

// UnmarshalJSON implements custom JSON unmarshaling for ProductIDString
// It converts both string and numeric JSON values to string
func (p *ProductIDString) UnmarshalJSON(data []byte) error {
	// Remove quotes if present (string case)
	s := strings.Trim(string(data), `"`)
	
	// If it's a raw number (no quotes), convert to string
	if s == string(data) {
		// It's a number, convert to string
		var num float64
		if err := json.Unmarshal(data, &num); err != nil {
			return fmt.Errorf("failed to unmarshal ProductID as number: %v", err)
		}
		*p = ProductIDString(fmt.Sprintf("%.0f", num))
	} else {
		// It's already a string
		*p = ProductIDString(s)
	}
	
	return nil
}

// String returns the string representation
func (p ProductIDString) String() string {
	return string(p)
}

// Deal represents a Bitrix24 deal
type Deal struct {
	ID            string  `json:"ID"`
	Title         string  `json:"TITLE"`
	ContactID     string  `json:"CONTACT_ID"`
	CompanyID     string  `json:"COMPANY_ID"`
	AssignedByID  string  `json:"ASSIGNED_BY_ID"`
	Opportunity   float64 `json:"-"`             // Deal amount (parsed separately)
	CurrencyID    string  `json:"CURRENCY_ID"`   // Deal currency
}

// DealRaw represents a Bitrix24 deal with raw string fields for parsing
type DealRaw struct {
	ID            string `json:"ID"`
	Title         string `json:"TITLE"`
	ContactID     string `json:"CONTACT_ID"`
	CompanyID     string `json:"COMPANY_ID"`
	AssignedByID  string `json:"ASSIGNED_BY_ID"`
	OpportunityRaw string `json:"OPPORTUNITY"`   // Deal amount as string
	CurrencyID    string `json:"CURRENCY_ID"`   // Deal currency
}

// Contact represents a Bitrix24 contact
type Contact struct {
	ID   string `json:"ID"`
	Name string `json:"NAME"`
}

// Company represents a Bitrix24 company
type Company struct {
	ID    string `json:"ID"`
	Title string `json:"TITLE"`
}

// User represents a Bitrix24 user
type User struct {
	ID       string `json:"ID"`
	Name     string `json:"NAME"`
	LastName string `json:"LAST_NAME"`
	FullName string `json:"FULL_NAME"`
}

// ProductSection represents a catalog section (folder)
type ProductSection struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID *int   `json:"iblockSectionId"` // Can be null for root sections
}

// Product represents a catalog product
type Product struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	IblockSectionId  *int   `json:"iblockSectionId"` // Can be null
}

// DealProductRow represents a product row in a deal
type DealProductRow struct {
	ProductID ProductIDString `json:"PRODUCT_ID"` // Product ID with custom unmarshaling
	Quantity  float64         `json:"QUANTITY"`
	Price     float64         `json:"PRICE"`
}


// BitrixResponse represents a generic API response
type BitrixResponse struct {
	Result interface{} `json:"result"`
	Error  *BitrixError `json:"error"`
}

// BitrixError represents an API error
type BitrixError struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// GetDealResponse represents the response from crm.deal.get
type GetDealResponse struct {
	Result Deal `json:"result"`
}

// GetContactResponse represents the response from crm.contact.get
type GetContactResponse struct {
	Result Contact `json:"result"`
}

// GetCompanyResponse represents the response from crm.company.get
type GetCompanyResponse struct {
	Result Company `json:"result"`
}

// CreateSectionResponse represents the response from catalog.section.add
type CreateSectionResponse struct {
	Result string `json:"result"`
}

// CreateProductResponse represents the response from catalog.product.add
type CreateProductResponse struct {
	Result string `json:"result"`
}

// AddProductRowResponse represents the response from crm.deal.productrows.set
type AddProductRowResponse struct {
	Result bool `json:"result"`
}

// ListSectionsResponse represents the response from catalog.section.list
type ListSectionsResponse struct {
	Result struct {
		Sections []ProductSection `json:"sections"`
	} `json:"result"`
	Total int `json:"total"`
}