package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// APIError represents an error from the JSMon API
type APIError struct {
	URL     string
	Message string
	Status  int
}

func (e *APIError) Error() string {
	return e.Message
}

// IsRateLimitError checks if the error is a rate limit error
func (e *APIError) IsRateLimitError() bool {
	// Check for 429 status code (Too Many Requests)
	if e.Status == 429 {
		return true
	}
	// Check for rate limit messages in the error
	message := strings.ToLower(e.Message)
	return strings.Contains(message, "rate limit") ||
		strings.Contains(message, "too many requests") ||
		strings.Contains(message, "quota exceeded") ||
		strings.Contains(message, "limit exceeded")
}

// IsInsufficientLimitsError checks if the error is about insufficient scan limits
func (e *APIError) IsInsufficientLimitsError() bool {
	message := strings.ToLower(e.Message)
	return strings.Contains(message, "insufficient") &&
		(strings.Contains(message, "scan limit") ||
			strings.Contains(message, "jsscan limit") ||
			strings.Contains(message, "limit") && strings.Contains(message, "exhausted"))
}

const APIBaseURL = "https://api.jsmon.sh/api/v2"

// Client handles API communication
type Client struct {
	APIKey     string
	HTTPClient *http.Client
	Headers    map[string]string
}

// NewClient creates a new API client
func NewClient(apiKey string, headers map[string]string) *Client {
	// Ensure headers map is never nil
	if headers == nil {
		headers = make(map[string]string)
	}
	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
		Headers:    headers,
	}
}

// CreateWorkspace creates a new workspace
func (c *Client) CreateWorkspace(workspaceName string) (string, error) {
	endpoint := APIBaseURL + "/createWorkspace"

	payload := map[string]string{
		"name": workspaceName,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("createWorkspace failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response to extract workspaceId
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	workspaceID, ok := result["workspaceId"].(string)
	if !ok {
		return "", fmt.Errorf("workspaceId not found in response")
	}

	return workspaceID, nil
}

// UploadURL uploads a URL for scanning
func (c *Client) UploadURL(jsURL, workspaceID string) error {
	endpoint := APIBaseURL + "/uploadUrl?wkspId=" + url.QueryEscape(workspaceID) + "&source=" + url.QueryEscape("cliScan")

	payload := map[string]interface{}{
		"url": jsURL,
	}

	// Include custom headers in the payload if any are provided
	if len(c.Headers) > 0 {
		payload["headers"] = c.Headers
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Note: Custom headers are included in the payload, not in HTTP headers
	// This allows the API to use them when making requests to the target URL

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to parse error message from JSON response
		var errorResp map[string]interface{}
		message := string(body)
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
				message = msg
			}
		}
		return &APIError{
			URL:     jsURL,
			Message: message,
			Status:  resp.StatusCode,
		}
	}

	return nil
}

// ScanDomain scans a domain
func (c *Client) ScanDomain(domain, workspaceID string) error {
	endpoint := APIBaseURL + "/automateScanDomain?wkspId=" + url.QueryEscape(workspaceID) + "&source=" + url.QueryEscape("cliScan")

	payload := map[string]interface{}{
		"domain": domain,
	}

	// Include custom headers in the payload if any are provided
	if len(c.Headers) > 0 {
		payload["headers"] = c.Headers
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Note: Custom headers are included in the payload, not in HTTP headers
	// This allows the API to use them when making requests to the target domain

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	bodyStr := strings.TrimSpace(string(body))

	// Always check for errors in response body, regardless of status code
	// The API might return 200 with error message in body
	hasError := false
	errorMessage := ""

	// Check if response contains error indicators
	bodyLower := strings.ToLower(bodyStr)
	if strings.Contains(bodyLower, "error") ||
		strings.Contains(bodyLower, "failed") ||
		strings.Contains(bodyLower, "error in scanning") {
		hasError = true
		errorMessage = bodyStr

		// Try to parse as JSON first
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			// Check for common error message fields
			if msg, ok := errorResp["message"].(string); ok && msg != "" {
				errorMessage = msg
			} else if msg, ok := errorResp["error"].(string); ok && msg != "" {
				errorMessage = msg
			} else if msg, ok := errorResp["errorMessage"].(string); ok && msg != "" {
				errorMessage = msg
			} else if msg, ok := errorResp["msg"].(string); ok && msg != "" {
				errorMessage = msg
			}
		} else {
			// Not JSON, try to extract meaningful error from text
			// Look for patterns like "[INF] domain, error in scanning"
			lines := strings.Split(bodyStr, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				lineLower := strings.ToLower(line)

				// Look for error lines with [INF] prefix
				// Format: [INF] domain, error in scanning
				if strings.Contains(lineLower, "[inf]") {
					// Extract the part after [INF]
					infIdx := strings.Index(lineLower, "[inf]")
					afterInf := strings.TrimSpace(line[infIdx+5:]) // Skip "[INF]"

					// Check if it contains error
					if strings.Contains(strings.ToLower(afterInf), "error") {
						// Split by comma to get domain and error
						parts := strings.Split(afterInf, ",")
						if len(parts) >= 2 {
							// Get the error part after the comma
							errorPart := strings.TrimSpace(strings.Join(parts[1:], ","))
							if errorPart != "" {
								errorMessage = errorPart
								break
							}
						} else {
							// No comma, extract after "error"
							if idx := strings.Index(strings.ToLower(afterInf), "error"); idx >= 0 {
								errorMessage = strings.TrimSpace(afterInf[idx:])
								break
							}
						}
					}
				} else if strings.Contains(lineLower, "error") && !strings.Contains(lineLower, "[inf]") {
					// Error without [INF] prefix
					if idx := strings.Index(strings.ToLower(line), "error"); idx >= 0 {
						errorMessage = strings.TrimSpace(line[idx:])
						break
					}
				}
			}

			// If we couldn't extract, try to find any line with error
			if errorMessage == "" {
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(strings.ToLower(line), "error") {
						errorMessage = line
						break
					}
				}
			}

			// If still no error message, use a truncated version of the body
			if errorMessage == "" && len(bodyStr) > 0 {
				// Take first 200 chars
				if len(bodyStr) > 200 {
					errorMessage = bodyStr[:200] + "..."
				} else {
					errorMessage = bodyStr
				}
			}
		}
	}

	// Also check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		hasError = true
		if errorMessage == "" {
			// Try to parse as JSON
			var errorResp map[string]interface{}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				if msg, ok := errorResp["message"].(string); ok && msg != "" {
					errorMessage = msg
				} else if msg, ok := errorResp["error"].(string); ok && msg != "" {
					errorMessage = msg
				}
			}
			if errorMessage == "" {
				errorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
		}
	}

	if hasError {
		return &APIError{
			URL:     domain,
			Message: errorMessage,
			Status:  resp.StatusCode,
		}
	}

	return nil
}

// UploadFile uploads a file for scanning
func (c *Client) UploadFile(filePath, workspaceID string) error {
	endpoint := APIBaseURL + "/uploadFile?wkspId=" + url.QueryEscape(workspaceID) + "&source=" + url.QueryEscape("cliScan")

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers (before setting standard headers to allow override)
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("uploadFile failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// TotalCountAnalysis represents the count analysis response
type TotalCountAnalysis struct {
	TotalUrls                           int `json:"totalUrls"`
	TotalDomains                        int `json:"totalDomains"`
	TotalS3Domains                      int `json:"totalS3Domains"`
	TotalEmails                         int `json:"totalEmails"`
	TotalIpv4Addresses                  int `json:"totalIpv4Addresses"`
	TotalApiPaths                       int `json:"totalApiPaths"`
	TotalJwtTokens                      int `json:"totalJwtTokens"`
	TotalGqlQueries                     int `json:"totalGqlQueries"`
	TotalGqlMutations                   int `json:"totalGqlMutations"`
	TotalGqlFragments                   int `json:"totalGqlFragments"`
	TotalNodeModules                    int `json:"totalNodeModules"`
	TotalGuids                          int `json:"totalGuids"`
	TotalValidNodeModules               int `json:"totalValidNodeModules"`
	TotalQueryParamsUrls                int `json:"totalQueryParamsUrls"`
	TotalS3DomainsInvalid               int `json:"totalS3DomainsInvalid"`
	TotalExtractedDomainsStatus         int `json:"totalExtractedDomainsStatus"`
	TotalSocialMediaUrls                int `json:"totalSocialMediaUrls"`
	TotalLocalhostUrls                  int `json:"totalLocalhostUrls"`
	TotalFilteredPortUrls               int `json:"totalFilteredPortUrls"`
	TotalFileExtensionUrls              int `json:"totalFileExtensionUrls"`
	TotalRdsInstances                   int `json:"totalRdsInstances"`
	TotalEc2Instances                   int `json:"totalEc2Instances"`
	TotalS3Buckets                      int `json:"totalS3Buckets"`
	TotalCloudFrontDomains              int `json:"totalCloudFrontDomains"`
	TotalApiGatewayEndpoints            int `json:"totalApiGatewayEndpoints"`
	TotalLambdaFunctions                int `json:"totalLambdaFunctions"`
	TotalCloudWatchEndpoints            int `json:"totalCloudWatchEndpoints"`
	TotalElbEndpoints                   int `json:"totalElbEndpoints"`
	TotalAppSyncEndpoints               int `json:"totalAppSyncEndpoints"`
	TotalCognitoEndpoints               int `json:"totalCognitoEndpoints"`
	TotalCognitoUserPoolIDs             int `json:"totalCognitoUserPoolIDs"`
	TotalCognitoAuthDomains             int `json:"totalCognitoAuthDomains"`
	TotalAmplifyDomains                 int `json:"totalAmplifyDomains"`
	TotalOpenSearchDomains              int `json:"totalOpenSearchDomains"`
	TotalTransferEndpoints              int `json:"totalTransferEndpoints"`
	TotalWorkServices                   int `json:"totalWorkServices"`
	TotalContainerEndpoints             int `json:"totalContainerEndpoints"`
	TotalIotEndpoints                   int `json:"totalIotEndpoints"`
	TotalKinesisEndpoints               int `json:"totalKinesisEndpoints"`
	TotalStsEndpoints                   int `json:"totalStsEndpoints"`
	TotalCloudFormationEndpoints        int `json:"totalCloudFormationEndpoints"`
	TotalCognitoIdentityPoolIDs         int `json:"totalCognitoIdentityPoolIDs"`
	TotalOtherAWSEndpoints              int `json:"totalOtherAWSEndpoints"`
	TotalExecData                       int `json:"totalExecData"`
	TotalSetTimeoutCalls                int `json:"totalSetTimeoutCalls"`
	TotalSetIntervalCalls               int `json:"totalSetIntervalCalls"`
	TotalDomXssPotentialVulnerabilities int `json:"totalDomXssPotentialVulnerabilities"`
	TotalDomBasedDOS                    int `json:"totalDomBasedDOS"`
	TotalClientSideSQLi                 int `json:"totalClientSideSQLi"`
	TotalDomBasedOpenRedirection        int `json:"totalDomBasedOpenRedirection"`
	TotalDomBasedLinkManipulation       int `json:"totalDomBasedLinkManipulation"`
	TotalDomBasedCookieManipulation     int `json:"totalDomBasedCookieManipulation"`
	TotalDomBasedJavaScriptInjection    int `json:"totalDomBasedJavaScriptInjection"`
	TotalDomBasedFilePathManipulation   int `json:"totalDomBasedFilePathManipulation"`
	TotalDomBasedAjaxHeaderManipulation int `json:"totalDomBasedAjaxHeaderManipulation"`
	TotalIpAddresses                    int `json:"totalIpAddresses"`
	TotalAwsAssets                      int `json:"totalAwsAssets"`
	TotalVulnerabilities                int `json:"totalVulnerabilities"`
	TotalExtractedParameters            int `json:"totalExtractedParameters"`
	TotalJsUrls                         int `json:"totalJsUrls"`
	TotalDocuments                      int `json:"totalDocuments"`
}

// GetTotalCountAnalysis retrieves the total count analysis for a workspace
func (c *Client) GetTotalCountAnalysis(workspaceID, runID string) (*TotalCountAnalysis, error) {
	endpoint := APIBaseURL + "/totalCountAnalysis?&wkspId=" + url.QueryEscape(workspaceID)
	if runID != "" {
		endpoint += "&runId=" + url.QueryEscape(runID)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getTotalCountAnalysis failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var countAnalysis TotalCountAnalysis
	if err := json.Unmarshal(body, &countAnalysis); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &countAnalysis, nil
}

// Workspace represents a workspace from the API
type Workspace struct {
	WkspID   string `json:"wkspId"`
	UserID   string `json:"userId"`
	OrgID    string `json:"orgId"`
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	IsShared bool   `json:"isShared"`
	Owner    string `json:"owner"`
}

// GetWorkspacesResponse represents the response from getWorkspaces
type GetWorkspacesResponse struct {
	Message    string      `json:"message"`
	Workspaces []Workspace `json:"workspaces"`
}

// GetWorkspaces retrieves all workspaces for the user
func (c *Client) GetWorkspaces() (*GetWorkspacesResponse, error) {
	endpoint := APIBaseURL + "/getWorkspaces"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getWorkspaces failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var response GetWorkspacesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// FileScan represents a file scan from the fetchScans API
type FileScan struct {
	ID            string      `json:"_id"`
	RunID         string      `json:"runId"`
	Category      string      `json:"category"`
	Asset         string      `json:"asset"` // This is the file name
	LastScannedOn string      `json:"lastScannedOn"`
	ThreatScore   int         `json:"threatScore"`
	Status        string      `json:"status"`
	FileID        string      `json:"fileId"`
	ScanID        string      `json:"scanId"`
	WkspID        string      `json:"wkspId"`
	UserID        string      `json:"userId"`
	Info          string      `json:"info"`
	Monitoring    bool        `json:"monitoring"`
	CronData      interface{} `json:"cronData"`
	CreatedAt     string      `json:"createdAt"`
	UpdatedAt     string      `json:"updatedAt"`
}

// FileScansData represents the data object in the fetchScans API response for file scans
type FileScansData struct {
	Scans      []FileScan `json:"scans"`
	Pagination struct {
		CurrentPage int  `json:"currentPage"`
		TotalPages  int  `json:"totalPages"`
		TotalItems  int  `json:"totalItems"`
		Limit       int  `json:"limit"`
		HasNext     bool `json:"hasNext"`
		HasPrev     bool `json:"hasPrev"`
	} `json:"pagination"`
	Filters struct {
		Category   string `json:"category"`
		Status     string `json:"status"`
		Monitoring string `json:"monitoring"`
		ScoreMin   string `json:"scoreMin"`
		ScoreMax   string `json:"scoreMax"`
		DateFrom   string `json:"dateFrom"`
		DateTo     string `json:"dateTo"`
		Search     string `json:"search"`
	} `json:"filters"`
	Stats struct {
		ID              interface{} `json:"_id"`
		TotalScans      int         `json:"totalScans"`
		URLScans        int         `json:"urlScans"`
		FileScans       int         `json:"fileScans"`
		DomainScans     int         `json:"domainScans"`
		SuccessCount    int         `json:"successCount"`
		FailedCount     int         `json:"failedCount"`
		InProgressCount int         `json:"inProgressCount"`
		AvgThreatScore  float64     `json:"avgThreatScore"`
	} `json:"stats"`
}

// GetFileScansResponse represents the response from fetchScans API for file scans
type GetFileScansResponse struct {
	Success bool          `json:"success"`
	Data    FileScansData `json:"data"`
}

// GetFileScans retrieves file scans for a workspace
func (c *Client) GetFileScans(workspaceID string, page int, status, search, scoreMin, scoreMax, dateFrom, dateTo, limit, monitoring string) (*GetFileScansResponse, error) {
	endpoint := APIBaseURL + "/fetchScans?wkspId=" + url.QueryEscape(workspaceID) + "&page=" + fmt.Sprintf("%d", page) + "&category=fileScan"

	if status != "" {
		endpoint += "&status=" + url.QueryEscape(status)
	}
	if search != "" {
		endpoint += "&search=" + url.QueryEscape(search)
	}
	if scoreMin != "" {
		endpoint += "&scoreMin=" + url.QueryEscape(scoreMin)
	}
	if scoreMax != "" {
		endpoint += "&scoreMax=" + url.QueryEscape(scoreMax)
	}
	if dateFrom != "" {
		endpoint += "&dateFrom=" + url.QueryEscape(dateFrom)
	}
	if dateTo != "" {
		endpoint += "&dateTo=" + url.QueryEscape(dateTo)
	}
	if limit != "" {
		endpoint += "&limit=" + url.QueryEscape(limit)
	}
	if monitoring != "" {
		endpoint += "&monitoring=" + url.QueryEscape(monitoring)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getFileScans failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response GetFileScansResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Pagination represents pagination info from the API
type Pagination struct {
	CurrentPage     int  `json:"currentPage"`
	TotalItems      int  `json:"totalItems"`
	TotalPages      int  `json:"totalPages"`
	ItemsPerPage    int  `json:"itemsPerPage"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

// JSURL represents a JS URL from the intelligence API
type JSURL struct {
	Value       string `json:"value"`
	Occurrences int    `json:"occurrences,omitempty"`
}

// GetJSURLsResponse represents the response from intelligence API
type GetJSURLsResponse struct {
	Data       []JSURL    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// GetJSURLs retrieves scanned URLs for a workspace using intelligence endpoint
func (c *Client) GetJSURLs(workspaceID string, page int, runID, search, status string) (*GetJSURLsResponse, error) {
	endpoint := APIBaseURL + "/intelligence?wkspId=" + url.QueryEscape(workspaceID) + "&options=jsurls&page=" + fmt.Sprintf("%d", page)

	if runID != "" {
		endpoint += "&runId=" + url.QueryEscape(runID)
	}
	if search != "" {
		endpoint += "&search=" + url.QueryEscape(search)
	}
	if status != "" {
		endpoint += "&status=" + url.QueryEscape(status)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getJSURLs failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response GetJSURLsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DomainScan represents a domain scan from the fetchScans API
type DomainScan struct {
	ID            string      `json:"_id"`
	RunID         string      `json:"runId"`
	Category      string      `json:"category"`
	Asset         string      `json:"asset"` // This is the domain name
	LastScannedOn string      `json:"lastScannedOn"`
	ThreatScore   int         `json:"threatScore"`
	Status        string      `json:"status"`
	ScanID        string      `json:"scanId"`
	WkspID        string      `json:"wkspId"`
	UserID        string      `json:"userId"`
	Info          string      `json:"info"`
	Monitoring    bool        `json:"monitoring"`
	CronData      interface{} `json:"cronData"`
	CreatedAt     string      `json:"createdAt"`
	UpdatedAt     string      `json:"updatedAt"`
}

// DomainScansData represents the data object in the fetchScans API response
type DomainScansData struct {
	Scans      []DomainScan `json:"scans"`
	Pagination struct {
		CurrentPage int  `json:"currentPage"`
		TotalPages  int  `json:"totalPages"`
		TotalItems  int  `json:"totalItems"`
		Limit       int  `json:"limit"`
		HasNext     bool `json:"hasNext"`
		HasPrev     bool `json:"hasPrev"`
	} `json:"pagination"`
	Filters struct {
		Category   string `json:"category"`
		Status     string `json:"status"`
		Monitoring string `json:"monitoring"`
		ScoreMin   string `json:"scoreMin"`
		ScoreMax   string `json:"scoreMax"`
		DateFrom   string `json:"dateFrom"`
		DateTo     string `json:"dateTo"`
		Search     string `json:"search"`
	} `json:"filters"`
	Stats struct {
		ID              interface{} `json:"_id"`
		TotalScans      int         `json:"totalScans"`
		URLScans        int         `json:"urlScans"`
		FileScans       int         `json:"fileScans"`
		DomainScans     int         `json:"domainScans"`
		SuccessCount    int         `json:"successCount"`
		FailedCount     int         `json:"failedCount"`
		InProgressCount int         `json:"inProgressCount"`
		AvgThreatScore  float64     `json:"avgThreatScore"`
	} `json:"stats"`
}

// GetDomainScansResponse represents the response from fetchScans API
type GetDomainScansResponse struct {
	Success bool            `json:"success"`
	Data    DomainScansData `json:"data"`
}

// GetDomainScans retrieves domain scans for a workspace
func (c *Client) GetDomainScans(workspaceID string, page int, status, search, scoreMin, scoreMax, dateFrom, dateTo, limit, monitoring string) (*GetDomainScansResponse, error) {
	endpoint := APIBaseURL + "/fetchScans?wkspId=" + url.QueryEscape(workspaceID) + "&page=" + fmt.Sprintf("%d", page) + "&category=domainScan"

	if status != "" {
		endpoint += "&status=" + url.QueryEscape(status)
	}
	if search != "" {
		endpoint += "&search=" + url.QueryEscape(search)
	}
	if scoreMin != "" {
		endpoint += "&scoreMin=" + url.QueryEscape(scoreMin)
	}
	if scoreMax != "" {
		endpoint += "&scoreMax=" + url.QueryEscape(scoreMax)
	}
	if dateFrom != "" {
		endpoint += "&dateFrom=" + url.QueryEscape(dateFrom)
	}
	if dateTo != "" {
		endpoint += "&dateTo=" + url.QueryEscape(dateTo)
	}
	if limit != "" {
		endpoint += "&limit=" + url.QueryEscape(limit)
	}
	if monitoring != "" {
		endpoint += "&monitoring=" + url.QueryEscape(monitoring)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getDomainScans failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response GetDomainScansResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Secret represents a secret from the keysAndSecrets API
type Secret struct {
	MatchedWord string `json:"matchedWord"`
	Severity    string `json:"severity"`
	CreatedAt   string `json:"createdAt"`
	Occurrences int    `json:"occurrences"`
	ModuleName  string `json:"moduleName"`
	Source      string `json:"source"`
}

// GetSecretsResponse represents the response from keysAndSecrets API
type GetSecretsResponse struct {
	Data       []Secret   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// GetSecrets retrieves secrets for a workspace
func (c *Client) GetSecrets(workspaceID string, page int, runID, lastScannedOn, formDate, toDate, limit, search string) (*GetSecretsResponse, error) {
	endpoint := APIBaseURL + "/keysAndSecrets?wkspId=" + url.QueryEscape(workspaceID) + "&page=" + fmt.Sprintf("%d", page)

	if runID != "" {
		endpoint += "&runId=" + url.QueryEscape(runID)
	}
	if lastScannedOn != "" {
		endpoint += "&lastScannedOn=" + url.QueryEscape(lastScannedOn)
	}
	if formDate != "" {
		endpoint += "&formDate=" + url.QueryEscape(formDate)
	}
	if toDate != "" {
		endpoint += "&toDate=" + url.QueryEscape(toDate)
	}
	if limit != "" {
		endpoint += "&limit=" + url.QueryEscape(limit)
	}
	if search != "" {
		endpoint += "&search=" + url.QueryEscape(search)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getSecrets failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response GetSecretsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// JSIntelligenceItem represents an item from the reconnaissance API
type JSIntelligenceItem struct {
	Value string `json:"value"`
	// Add other fields as needed based on API response
}

// GetJSIntelligenceResponse represents the response from reconnaissance API
type GetJSIntelligenceResponse struct {
	Data       []JSIntelligenceItem `json:"data"`
	Pagination Pagination           `json:"pagination"`
}

// mapFieldToOptions converts user-provided field names to API options parameter values
func mapFieldToOptions(field string) string {
	fieldMap := map[string]string{
		"apipaths":        "apipaths",
		"urls":            "urls",
		"domains":         "domains",
		"ip":              "ipaddresses",
		"emails":          "emails",
		"s3buckets":       "s3domains",
		"s3takeovers":     "s3invalid",
		"gqlqueries":      "gqlqueries",
		"gqlmutaions":     "gqlmutations",
		"gqlfragments":    "gqlfragments",
		"param":           "parameters",
		"npmpackages":     "validnodemodules",
		"npmconfusion":    "invalidnodemodules",
		"guids":           "guids",
		"localhost":       "localhost",
		"activedomains":   "domainsstatus",
		"inactivedomains": "domainsstatus",
		"awsassets":       "awsassets",
		"queryparam":      "queryparams",
		"socialurls":      "socialmediaurls",
		"porturls":        "filteredporturls",
		"extensionurls":   "fileextensionurls",
	}

	// Check if there's a mapping, otherwise use the field name as-is
	if mapped, exists := fieldMap[strings.ToLower(field)]; exists {
		return mapped
	}
	// Default: use the field name as-is (for fields that match exactly)
	return strings.ToLower(field)
}

// getStatusForField returns the status parameter value for fields that require it
func getStatusForField(field string) string {
	fieldLower := strings.ToLower(field)
	if fieldLower == "activedomains" {
		return "active"
	}
	if fieldLower == "inactivedomains" {
		return "inactive"
	}
	return ""
}

// GetJSIntelligence retrieves reconnaissance data for a workspace
func (c *Client) GetJSIntelligence(workspaceID, field string, page int, runID, search, status string) (*GetJSIntelligenceResponse, error) {
	// Map field name to API options parameter
	optionsValue := mapFieldToOptions(field)
	endpoint := APIBaseURL + "/intelligence?wkspId=" + url.QueryEscape(workspaceID) + "&options=" + url.QueryEscape(optionsValue) + "&page=" + fmt.Sprintf("%d", page)

	if runID != "" {
		endpoint += "&runId=" + url.QueryEscape(runID)
	}
	if search != "" {
		endpoint += "&search=" + url.QueryEscape(search)
	}
	// Use field-specific status if provided, otherwise use the passed status parameter
	fieldStatus := getStatusForField(field)
	if fieldStatus != "" {
		status = fieldStatus
	}
	if status != "" {
		endpoint += "&status=" + url.QueryEscape(status)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getJSIntelligence failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response GetJSIntelligenceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetJSIntelligenceRaw retrieves reconnaissance data for a workspace and returns the raw JSON response
func (c *Client) GetJSIntelligenceRaw(workspaceID, field string, page int, runID, search, status string) ([]byte, error) {
	// Map field name to API options parameter
	optionsValue := mapFieldToOptions(field)
	endpoint := APIBaseURL + "/intelligence?wkspId=" + url.QueryEscape(workspaceID) + "&options=" + url.QueryEscape(optionsValue) + "&page=" + fmt.Sprintf("%d", page)

	if runID != "" {
		endpoint += "&runId=" + url.QueryEscape(runID)
	}
	if search != "" {
		endpoint += "&search=" + url.QueryEscape(search)
	}
	// Use field-specific status if provided, otherwise use the passed status parameter
	fieldStatus := getStatusForField(field)
	if fieldStatus != "" {
		status = fieldStatus
	}
	if status != "" {
		endpoint += "&status=" + url.QueryEscape(status)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("getJSIntelligence failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ReverseSearchRequest represents the request body for reverse search
type ReverseSearchRequest struct {
	Value string `json:"value"`
}

// ReverseSearchResponse represents the response from intelligenceSearch API
type ReverseSearchResponse struct {
	Success bool                     `json:"success,omitempty"`
	Data    []map[string]interface{} `json:"data"`
	Message string                   `json:"message,omitempty"`
}

// ReverseSearch performs a reverse search for a given field and value
func (c *Client) ReverseSearch(workspaceID, field, searchValue string) (*ReverseSearchResponse, error) {
	// Map field name to API options parameter (same as reconnaissance)
	optionsValue := mapFieldToOptions(field)
	endpoint := APIBaseURL + "/intelligenceSearch?wkspId=" + url.QueryEscape(workspaceID) + "&options=" + url.QueryEscape(optionsValue)

	// Create request body
	requestBody := ReverseSearchRequest{
		Value: searchValue,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Debug: The JSON marshaler will properly escape newlines as \n in the JSON string
	// This is correct - when the JSON is parsed by the API, \n will become actual newlines
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jsmon-Key", strings.TrimSpace(c.APIKey))

	// Apply custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("reverseSearch failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response - try to handle different response structures
	var response ReverseSearchResponse

	// First, try parsing as an object with data field
	if err := json.Unmarshal(body, &response); err == nil {
		// If we successfully parsed and have data, return it
		if response.Data != nil {
			return &response, nil
		}
	}

	// If that didn't work or data is empty, try parsing as a direct data array
	var directData []map[string]interface{}
	if err := json.Unmarshal(body, &directData); err == nil {
		response.Data = directData
		return &response, nil
	}

	// If both fail, try parsing as a generic map to see the structure
	var genericResponse map[string]interface{}
	if err := json.Unmarshal(body, &genericResponse); err == nil {
		// Check if there's a "data" field
		if data, ok := genericResponse["data"]; ok {
			if dataArray, ok := data.([]interface{}); ok {
				response.Data = make([]map[string]interface{}, 0, len(dataArray))
				for _, item := range dataArray {
					if itemMap, ok := item.(map[string]interface{}); ok {
						response.Data = append(response.Data, itemMap)
					}
				}
				return &response, nil
			}
		}
		// If no data field, try to use the entire response as data
		if len(genericResponse) > 0 {
			response.Data = []map[string]interface{}{genericResponse}
			return &response, nil
		}
	}

	// If all parsing attempts fail, return error with the raw body for debugging
	return nil, fmt.Errorf("failed to parse response. Raw response: %s", string(body))
}
