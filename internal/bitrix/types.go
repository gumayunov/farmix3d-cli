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

// ProductInfo represents product information extracted from filename
type ProductInfo struct {
	ID       string  // Product ID from Bitrix24
	Quantity float64 // Quantity extracted from filename or default 1.0
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

// StoreDocument represents a warehouse document
type StoreDocument struct {
	ID           string `json:"id"`
	DocType      string `json:"docType"`      // 'S' for receipt
	Currency     string `json:"currency"`
	DateDocument string `json:"dateDocument"`
	Commentary   string `json:"commentary"`
	ResponsibleID string `json:"responsibleId"`
}

// StoreDocumentElement represents an element in a warehouse document
type StoreDocumentElement struct {
	DocID           string  `json:"docId"`
	StoreFrom       int     `json:"storeFrom"`       // 0 for receipt
	StoreTo         string  `json:"storeTo"`         // Target warehouse ID
	ElementID       string  `json:"elementId"`       // Product ID
	Amount          float64 `json:"amount"`          // Quantity
	PurchasingPrice float64 `json:"purchasingPrice"` // Price per unit
}

// StoreDocumentModeResponse represents the response from catalog.document.mode.status
// The API returns result as a direct string value, not as an object
type StoreDocumentModeResponse string

// CreateStoreDocumentResponse represents the response from catalog.document.add
// The API returns result as a direct string value (document ID)
type CreateStoreDocumentResponse string

// AddStoreDocumentElementResponse represents the response from catalog.document.element.add
// The API returns result as a direct string value (element ID)
type AddStoreDocumentElementResponse string

// ConfirmStoreDocumentResponse represents the response from catalog.document.confirm
// The API returns result as a direct boolean value
type ConfirmStoreDocumentResponse bool

// StoreImage represents an image in Bitrix24
type StoreImage struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// Store represents a warehouse in Bitrix24
type Store struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Active      string      `json:"active"`      // "Y" or "N"
	Code        *string     `json:"code"`        // Store code (can be null)
	Sort        int         `json:"sort"`        // Sort order
	Address     string      `json:"address"`     // Store address
	Description *string     `json:"description"` // Store description (can be null)
	ImageID     *StoreImage `json:"imageId"`     // Image info (can be null)
	Phone       *string     `json:"phone"`       // Phone number (can be null)
	Schedule    *string     `json:"schedule"`    // Working schedule (can be null)
	Email       *string     `json:"email"`       // Email (can be null)
	Coords      *string     `json:"coords"`      // Coordinates (can be null)
	SiteID      *string     `json:"siteId"`      // Site ID (can be null)
	IssuingCenter *string   `json:"issuingCenter"` // "Y" or "N" (can be null)
	ShippingCenter *string  `json:"shippingCenter"` // "Y" or "N" (can be null)
}

// GetStoreResponse represents the response from catalog.store.get
type GetStoreResponse struct {
	Result Store `json:"result"`
}

// ListStoresResponse represents the response from catalog.store.list
type ListStoresResponse struct {
	Result struct {
		Stores []Store `json:"stores"`
	} `json:"result"`
	Total int `json:"total"`
}

// DealReportRow represents a deal row in the report with custom fields
type DealReportRow struct {
	ID              string      `json:"ID"`
	Title           string      `json:"TITLE"`
	DateCreate      string      `json:"DATE_CREATE"`
	MachineCost     interface{} `json:"machine_cost"`     // Custom field - can be string or number
	HumanCost       interface{} `json:"human_cost"`       // Custom field - can be string or number
	MaterialCost    interface{} `json:"material_cost"`    // Custom field - can be string or number
	TotalCost       interface{} `json:"total_cost"`       // Custom field - can be string or number
	PaymentReceived interface{} `json:"payment_received"` // Custom field - can be string or number
}

// ListDealsResponse represents the response from crm.deal.list
type ListDealsResponse struct {
	Result []map[string]interface{} `json:"result"` // Array of deals with dynamic fields
	Total  int                      `json:"total"`
}

// ReportCustomFields contains the field codes for custom fields in reports
type ReportCustomFields struct {
	MachineCost     string `json:"machine_cost"`
	HumanCost       string `json:"human_cost"`
	MaterialCost    string `json:"material_cost"`
	TotalCost       string `json:"total_cost"`
	PaymentReceived string `json:"payment_received"`
}