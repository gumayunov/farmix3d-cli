package bitrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Bitrix24 API client
type Client struct {
	webhookURL string
	httpClient *http.Client
}

// NewClient creates a new Bitrix24 client
func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetWebhookURL returns the webhook URL (for internal use)
func (c *Client) GetWebhookURL() string {
	return c.webhookURL
}

// MakeRequest makes an HTTP request to Bitrix24 API (public for testing)
func (c *Client) MakeRequest(method string, params map[string]interface{}) (*http.Response, error) {
	return c.makeRequest(method, params)
}

// makeRequest makes an HTTP request to Bitrix24 API
func (c *Client) makeRequest(method string, params map[string]interface{}) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s/%s", c.webhookURL, method)
	
	// Prepare form data
	formData := url.Values{}
	for key, value := range params {
		switch v := value.(type) {
		case string:
			formData.Set(key, v)
		case []string:
			// For string arrays, use Bitrix24 format: key[]=value1&key[]=value2
			for _, item := range v {
				formData.Add(key+"[]", item)
			}
		case map[string]interface{}:
			// Special handling for filter parameter
			if key == "filter" {
				// For filter, use Bitrix24 format: filter[FIELD]=value
				for filterKey, filterValue := range v {
					if str, ok := filterValue.(string); ok {
						formData.Set(fmt.Sprintf("filter[%s]", filterKey), str)
					} else {
						// Convert other types to string
						jsonValue, err := json.Marshal(filterValue)
						if err != nil {
							return nil, fmt.Errorf("failed to marshal filter value %s: %v", filterKey, err)
						}
						formData.Set(fmt.Sprintf("filter[%s]", filterKey), string(jsonValue))
					}
				}
			} else if key == "fields" {
				// For fields, use Bitrix24 format: fields[FIELD]=value
				for fieldKey, fieldValue := range v {
					if str, ok := fieldValue.(string); ok {
						formData.Set(fmt.Sprintf("fields[%s]", fieldKey), str)
					} else {
						// Convert other types to string
						jsonValue, err := json.Marshal(fieldValue)
						if err != nil {
							return nil, fmt.Errorf("failed to marshal field value %s: %v", fieldKey, err)
						}
						formData.Set(fmt.Sprintf("fields[%s]", fieldKey), string(jsonValue))
					}
				}
			} else {
				// For other nested objects, encode as JSON
				jsonValue, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal %s: %v", key, err)
				}
				formData.Set(key, string(jsonValue))
			}
		case []interface{}:
			// Special handling for 'rows' parameter in deal.productrows.set
			if key == "rows" {
				for i, item := range v {
					if rowMap, ok := item.(map[string]interface{}); ok {
						for rowKey, rowValue := range rowMap {
							if str, ok := rowValue.(string); ok {
								formData.Set(fmt.Sprintf("rows[%d][%s]", i, rowKey), str)
							} else {
								jsonValue, err := json.Marshal(rowValue)
								if err != nil {
									return nil, fmt.Errorf("failed to marshal row value %s: %v", rowKey, err)
								}
								formData.Set(fmt.Sprintf("rows[%d][%s]", i, rowKey), string(jsonValue))
							}
						}
					} else {
						jsonValue, err := json.Marshal(item)
						if err != nil {
							return nil, fmt.Errorf("failed to marshal row item: %v", err)
						}
						formData.Add("rows[]", string(jsonValue))
					}
				}
			} else {
				// For generic arrays, try to convert to strings
				for _, item := range v {
					if str, ok := item.(string); ok {
						formData.Add(key+"[]", str)
					} else {
						jsonValue, err := json.Marshal(item)
						if err != nil {
							return nil, fmt.Errorf("failed to marshal array item in %s: %v", key, err)
						}
						formData.Add(key+"[]", string(jsonValue))
					}
				}
			}
		default:
			// For other types, convert to string
			jsonValue, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal %s: %v", key, err)
			}
			formData.Set(key, string(jsonValue))
		}
	}

	// Log the request for debugging (remove in production)
	// fmt.Printf("DEBUG: %s request to %s\n", "POST", requestURL)
	// fmt.Printf("DEBUG: Form data: %s\n", formData.Encode())
	
	req, err := http.NewRequest("POST", requestURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	// Log response status (remove in production)
	// fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	
	return resp, nil
}

// ParseResponse parses HTTP response into a generic BitrixResponse (public for testing)
func (c *Client) ParseResponse(resp *http.Response, target interface{}) error {
	return c.parseResponse(resp, target)
}

// parseResponse parses HTTP response into a generic BitrixResponse
func (c *Client) parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var bitrixResp BitrixResponse
	if err := json.Unmarshal(body, &bitrixResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if bitrixResp.Error != nil {
		return fmt.Errorf("Bitrix24 API error %s: %s", bitrixResp.Error.ErrorCode, bitrixResp.Error.ErrorDescription)
	}

	// Marshal the result back to JSON and unmarshal into target
	resultJSON, err := json.Marshal(bitrixResp.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	if err := json.Unmarshal(resultJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal result into target: %v", err)
	}

	return nil
}