package bitrix

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ListDealsWithCustomFields retrieves deals with custom fields, excluding specified statuses
// categoryIDs - optional list of category IDs to filter by (empty = all categories)
func (c *Client) ListDealsWithCustomFields(customFields ReportCustomFields, excludedStatuses []string, categoryIDs []string) ([]DealReportRow, error) {
	// Build select fields list - standard fields + custom fields
	selectFields := []string{
		"ID",
		"TITLE",
		"DATE_CREATE",
		"CATEGORY_ID",
		"OPPORTUNITY",
	}

	// Add custom fields to select if they are configured
	if customFields.MachineCost != "" {
		selectFields = append(selectFields, customFields.MachineCost)
	}
	if customFields.HumanCost != "" {
		selectFields = append(selectFields, customFields.HumanCost)
	}
	if customFields.MaterialCost != "" {
		selectFields = append(selectFields, customFields.MaterialCost)
	}
	if customFields.TotalCost != "" {
		selectFields = append(selectFields, customFields.TotalCost)
	}
	if customFields.PaymentReceived != "" {
		selectFields = append(selectFields, customFields.PaymentReceived)
	}

	// Build filter to exclude final statuses
	filter := make(map[string]interface{})

	// Exclude specified statuses (e.g., WON, LOST)
	// In Bitrix24, to exclude multiple values, we use "!@STAGE_ID" with array
	if len(excludedStatuses) > 0 {
		// Convert to interface slice for the filter
		statusesInterface := make([]interface{}, len(excludedStatuses))
		for i, status := range excludedStatuses {
			statusesInterface[i] = status
		}
		filter["!@STAGE_ID"] = statusesInterface
	}

	// Filter by category IDs if specified
	// In Bitrix24, to filter by multiple values, we use "@CATEGORY_ID" with array
	if len(categoryIDs) > 0 {
		// Convert to interface slice for the filter
		categoriesInterface := make([]interface{}, len(categoryIDs))
		for i, categoryID := range categoryIDs {
			categoriesInterface[i] = categoryID
		}
		filter["@CATEGORY_ID"] = categoriesInterface
	}

	params := map[string]interface{}{
		"select": selectFields,
		"filter": filter,
		"order":  map[string]interface{}{"ID": "ASC"}, // Sort by ID ascending
	}

	resp, err := c.makeRequest("crm.deal.list", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list deals: %v", err)
	}

	// Parse response as array of maps
	var result []map[string]interface{}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse deals response: %v", err)
	}

	// Convert to DealReportRow structs
	deals := make([]DealReportRow, 0, len(result))
	for _, dealMap := range result {
		deal := DealReportRow{
			ID:         getStringValue(dealMap, "ID"),
			Title:      getStringValue(dealMap, "TITLE"),
			DateCreate: getStringValue(dealMap, "DATE_CREATE"),
			CategoryID: getStringValue(dealMap, "CATEGORY_ID"),
		}

		// Map custom fields
		if customFields.MachineCost != "" {
			deal.MachineCost = dealMap[customFields.MachineCost]
		}
		if customFields.HumanCost != "" {
			deal.HumanCost = dealMap[customFields.HumanCost]
		}
		if customFields.MaterialCost != "" {
			deal.MaterialCost = dealMap[customFields.MaterialCost]
		}
		if customFields.TotalCost != "" {
			deal.TotalCost = dealMap[customFields.TotalCost]
		}

		// Map standard deal fields
		deal.Opportunity = dealMap["OPPORTUNITY"]

		if customFields.PaymentReceived != "" {
			deal.PaymentReceived = dealMap[customFields.PaymentReceived]
		}

		deals = append(deals, deal)
	}

	// Sort by ID (converting string to int for proper numeric sorting)
	sort.Slice(deals, func(i, j int) bool {
		idI, errI := strconv.Atoi(deals[i].ID)
		idJ, errJ := strconv.Atoi(deals[j].ID)
		if errI != nil || errJ != nil {
			// Fallback to string comparison if conversion fails
			return deals[i].ID < deals[j].ID
		}
		return idI < idJ
	})

	return deals, nil
}

// ParseCustomFieldValue converts a custom field value to a standardized format
// Returns string representation of the value, handling numbers, strings, and booleans
func ParseCustomFieldValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		// Remove currency suffix from monetary fields (e.g., "1000|RUB" -> "1000")
		if strings.Contains(v, "|") {
			parts := strings.Split(v, "|")
			return parts[0]
		}
		return v
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.2f", v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case bool:
		if v {
			return "Да"
		}
		return "Нет"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// getStringValue safely extracts a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// ListDealCategories retrieves deal categories (funnels) from Bitrix24
// Returns a map of category ID to category name for quick lookups
func (c *Client) ListDealCategories() (map[string]string, error) {
	params := map[string]interface{}{
		"entityTypeId": 2, // 2 = Deals (CRM_DEAL)
	}

	resp, err := c.makeRequest("crm.category.list", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %v", err)
	}

	// Parse response
	var result struct {
		Result []DealCategory `json:"result"`
	}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse categories response: %v", err)
	}

	// Convert to map: ID -> Name
	categoryMap := make(map[string]string)
	for _, category := range result.Result {
		categoryMap[fmt.Sprintf("%d", category.ID)] = category.Name
	}

	return categoryMap, nil
}
