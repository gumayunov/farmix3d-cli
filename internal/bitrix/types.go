package bitrix

// Deal represents a Bitrix24 deal
type Deal struct {
	ID            string `json:"ID"`
	Title         string `json:"TITLE"`
	ContactID     string `json:"CONTACT_ID"`
	CompanyID     string `json:"COMPANY_ID"`
	AssignedByID  string `json:"ASSIGNED_BY_ID"`
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
	ProductID string  `json:"PRODUCT_ID"`
	Quantity  float64 `json:"QUANTITY"`
	Price     float64 `json:"PRICE"`
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