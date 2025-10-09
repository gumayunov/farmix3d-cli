package bitrix

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

// CheckStoreDocumentMode checks if warehouse management is enabled in Bitrix24
func (c *Client) CheckStoreDocumentMode() (bool, error) {
	params := map[string]interface{}{}

	resp, err := c.makeJSONRequest("catalog.document.mode.status", params)
	if err != nil {
		return false, fmt.Errorf("failed to check warehouse mode: %v", err)
	}

	// Read raw response for parsing
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response manually
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return false, fmt.Errorf("failed to unmarshal raw response: %v", err)
	}

	// Check if there's an error in response
	if errorObj, exists := rawResponse["error"]; exists {
		errorDesc := ""
		if desc, ok := rawResponse["error_description"].(string); ok {
			errorDesc = fmt.Sprintf(": %s", desc)
		}
		return false, fmt.Errorf("Bitrix24 API error %v%s", errorObj, errorDesc)
	}

	// Extract result
	result, exists := rawResponse["result"]
	if !exists {
		return false, fmt.Errorf("no 'result' field in response")
	}

	// Check if warehouse mode is enabled
	if modeStr, ok := result.(string); ok {
		return modeStr == "Y", nil
	} else {
		return false, fmt.Errorf("unexpected result type for mode: %T", result)
	}
}

// CreateStoreDocument creates a new warehouse receipt document
func (c *Client) CreateStoreDocument(deal *Deal, currency, commentary string) (string, error) {
	// Use current date in Bitrix24 format
	currentDate := time.Now().Format(time.RFC3339)

	// Use deal's assigned user as responsible, fallback to "1" (admin)
	responsibleID := deal.AssignedByID
	if responsibleID == "" || responsibleID == "0" {
		responsibleID = "1" // Default to admin user
	}

	fields := map[string]interface{}{
		"docType":      "S",        // Receipt document type
		"responsibleId": responsibleID,
		"currency":     currency,
		"dateDocument": currentDate,
		"commentary":   commentary,
		"title":        fmt.Sprintf("Оприходование изделий по сделке %s", deal.ID), // Document title (will be updated with ID)
		"UF_CAT_STORE_DOCUMENT_S_1758649547": deal.ID, // Link to deal (custom field)
		// Do not set status field - let API use default status
	}

	params := map[string]interface{}{
		"fields": fields,
	}

	resp, err := c.makeJSONRequest("catalog.document.add", params)
	if err != nil {
		return "", fmt.Errorf("failed to create warehouse document: %v", err)
	}

	// Read raw response for parsing
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response manually
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal raw response: %v", err)
	}

	// Check if there's an error in response
	if errorObj, exists := rawResponse["error"]; exists {
		errorDesc := ""
		if desc, ok := rawResponse["error_description"].(string); ok {
			errorDesc = fmt.Sprintf(": %s", desc)
		}
		return "", fmt.Errorf("Bitrix24 API error %v%s", errorObj, errorDesc)
	}

	// Extract result
	result, exists := rawResponse["result"]
	if !exists {
		return "", fmt.Errorf("no 'result' field in response")
	}

	// Convert result to string
	documentID := ""
	switch v := result.(type) {
	case string:
		documentID = v
	case float64:
		documentID = fmt.Sprintf("%.0f", v)
	case map[string]interface{}:
		// Result is an object, try to extract document ID
		if id, exists := v["document"]; exists {
			switch idVal := id.(type) {
			case string:
				documentID = idVal
			case float64:
				documentID = fmt.Sprintf("%.0f", idVal)
			case map[string]interface{}:
				// Document field is also an object, try to extract ID from it
				if docId, docExists := idVal["id"]; docExists {
					switch docIdVal := docId.(type) {
					case string:
						documentID = docIdVal
					case float64:
						documentID = fmt.Sprintf("%.0f", docIdVal)
					default:
						return "", fmt.Errorf("unexpected nested document ID type: %T", docIdVal)
					}
				} else {
					// Debug: show what's in the document object
					fmt.Printf("DEBUG: Document object contents: %+v\n", idVal)
					return "", fmt.Errorf("no 'id' field found in document object")
				}
			default:
				return "", fmt.Errorf("unexpected document ID type in result object: %T", idVal)
			}
		} else if id, exists := v["id"]; exists {
			switch idVal := id.(type) {
			case string:
				documentID = idVal
			case float64:
				documentID = fmt.Sprintf("%.0f", idVal)
			default:
				return "", fmt.Errorf("unexpected ID type in result object: %T", idVal)
			}
		} else {
			// Debug: show what's in the result object
			fmt.Printf("DEBUG: Result object contents: %+v\n", v)
			return "", fmt.Errorf("no 'document' or 'id' field found in result object")
		}
	default:
		return "", fmt.Errorf("unexpected result type: %T", result)
	}

	return documentID, nil
}


// AddElementsToStoreDocument adds product elements to a warehouse document
func (c *Client) AddElementsToStoreDocument(documentID string, products []DealProductRow, storeID string) error {
	for _, product := range products {
		err := c.addSingleElementToDocument(documentID, product, storeID)
		if err != nil {
			return fmt.Errorf("failed to add product %s to document: %v", product.ProductID.String(), err)
		}
	}
	return nil
}

// addSingleElementToDocument adds a single product element to warehouse document
func (c *Client) addSingleElementToDocument(documentID string, product DealProductRow, storeID string) error {
	// Convert documentID to integer if it's a numeric string
	var docID interface{} = documentID
	if id, err := strconv.Atoi(documentID); err == nil {
		docID = id
	}

	// Convert storeID to integer if it's a numeric string
	var storeToID interface{} = storeID
	if id, err := strconv.Atoi(storeID); err == nil {
		storeToID = id
	}

	fields := map[string]interface{}{
		"docId":          docID,                      // Document ID as number
		"storeFrom":      0,                          // 0 for receipt documents
		"storeTo":        storeToID,                  // Target warehouse ID as number
		"elementId":      product.ProductID.String(), // Product ID
		"amount":         product.Quantity,           // Quantity
		"purchasingPrice": 0,                         // Purchasing price = 0
		"sellingPrice":   product.Price,              // Selling price from deal
	}

	fmt.Printf("DEBUG: Добавляем товар %s (количество: %.2f, цена продажи: %.2f) в документ %v на склад %v\n",
		product.ProductID.String(), product.Quantity, product.Price, docID, storeToID)

	params := map[string]interface{}{
		"fields": fields,
	}

	resp, err := c.makeRequest("catalog.document.element.add", params)
	if err != nil {
		return fmt.Errorf("failed to add element to document: %v", err)
	}

	// Read raw response for parsing
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response manually since structure changed
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return fmt.Errorf("failed to unmarshal raw response: %v", err)
	}

	// Check if there's an error in response
	if errorObj, exists := rawResponse["error"]; exists {
		errorDesc := ""
		if desc, ok := rawResponse["error_description"].(string); ok {
			errorDesc = fmt.Sprintf(": %s", desc)
		}
		return fmt.Errorf("Bitrix24 API error %v%s", errorObj, errorDesc)
	}

	// result contains the element ID, we don't need to use it here
	return nil
}

// ConfirmStoreDocument confirms (проводит) the warehouse document to update inventory
func (c *Client) ConfirmStoreDocument(documentID string) error {
	// Convert documentID to integer if it's a numeric string
	var docID interface{} = documentID
	if id, err := strconv.Atoi(documentID); err == nil {
		docID = id
	}

	params := map[string]interface{}{
		"id": docID,
	}

	resp, err := c.makeJSONRequest("catalog.document.confirm", params)
	if err != nil {
		return fmt.Errorf("failed to confirm warehouse document: %v", err)
	}

	// Read raw response for parsing
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response manually
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return fmt.Errorf("failed to unmarshal raw response: %v", err)
	}

	// Check if there's an error in response
	if errorObj, exists := rawResponse["error"]; exists {
		errorDesc := ""
		if desc, ok := rawResponse["error_description"].(string); ok {
			errorDesc = fmt.Sprintf(": %s", desc)
		}
		return fmt.Errorf("Bitrix24 API error %v%s", errorObj, errorDesc)
	}

	// Extract result
	result, exists := rawResponse["result"]
	if !exists {
		return fmt.Errorf("no 'result' field in response")
	}

	// Check if confirmation was successful
	if confirmed, ok := result.(bool); ok {
		if !confirmed {
			return fmt.Errorf("failed to confirm document: API returned false")
		}
	} else {
		return fmt.Errorf("unexpected result type for confirm: %T", result)
	}

	return nil
}

// GetStore retrieves warehouse information by ID
func (c *Client) GetStore(storeID string) (*Store, error) {
	params := map[string]interface{}{
		"id": storeID,
	}

	resp, err := c.makeJSONRequest("catalog.store.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %v", err)
	}

	// Read raw response for parsing
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response manually
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw response: %v", err)
	}

	// Check if there's an error in response
	if errorObj, exists := rawResponse["error"]; exists {
		return nil, fmt.Errorf("Bitrix24 API error: %v", errorObj)
	}

	// Extract result
	result, exists := rawResponse["result"]
	if !exists {
		return nil, fmt.Errorf("no 'result' field in response")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("result is not a map: %T", result)
	}

	storeData, exists := resultMap["store"]
	if !exists {
		return nil, fmt.Errorf("no 'store' field in result")
	}

	// Convert to JSON and back to parse into Store struct
	storeJSON, err := json.Marshal(storeData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal store data: %v", err)
	}

	var store Store
	if err := json.Unmarshal(storeJSON, &store); err != nil {
		return nil, fmt.Errorf("failed to unmarshal store: %v", err)
	}

	return &store, nil
}

// ListStores retrieves list of all warehouses
func (c *Client) ListStores() ([]Store, error) {
	params := map[string]interface{}{
		"select": []string{"id", "title", "active", "code", "address", "description", "sort"},
	}

	resp, err := c.makeJSONRequest("catalog.store.list", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list stores: %v", err)
	}

	// Read raw response body for parsing
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response manually
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw response: %v", err)
	}

	// Check if there's an error in response
	if errorObj, exists := rawResponse["error"]; exists {
		return nil, fmt.Errorf("Bitrix24 API error: %v", errorObj)
	}

	// Extract result
	result, exists := rawResponse["result"]
	if !exists {
		return nil, fmt.Errorf("no 'result' field in response")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("result is not a map: %T", result)
	}

	storesData, exists := resultMap["stores"]
	if !exists {
		return nil, fmt.Errorf("no 'stores' field in result")
	}

	// Convert to JSON and back to parse into Store structs
	storesJSON, err := json.Marshal(storesData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stores data: %v", err)
	}

	var stores []Store
	if err := json.Unmarshal(storesJSON, &stores); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stores: %v", err)
	}

	return stores, nil
}