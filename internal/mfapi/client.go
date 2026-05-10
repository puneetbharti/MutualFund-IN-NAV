package mfapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/puneetbharti/navcalc/internal/types"
)

const (
	mfapiBaseURL = "https://api.mfapi.in/mf"
	httpTimeout  = 15 * time.Second
	maxDayOffset = 7
)

// Client for MFapi.in API
type Client struct {
	httpClient  *http.Client
	schemeMu    sync.RWMutex
	schemeCache map[string]string // ISIN -> SchemeCode
}

// SearchResponse represents the MFapi search response
type SearchResponse struct {
	SchemeCode int    `json:"schemeCode"`
	SchemeName string `json:"schemeName"`
	ISIN       string `json:"isin"`
}

// NAVHistoryResponse represents the full NAV history response
type NAVHistoryResponse struct {
	Meta struct {
		FundHouse  string `json:"fund_house"`
		SchemeType string `json:"scheme_type"`
		SchemeCode int    `json:"scheme_code"`
		SchemeName string `json:"scheme_name"`
		ISIN       string `json:"isin"`
	} `json:"meta"`
	Status string `json:"status"`
	Data   []struct {
		Date string `json:"date"`
		NAV  string `json:"nav"`
	} `json:"data"`
}

// NewClient creates a new MFapi client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		schemeCache: make(map[string]string),
	}
}

// GetNAVForISIN fetches NAV for a specific ISIN and target date
// Falls back to closest available date within ±7 days
func (c *Client) GetNAVForISIN(isin string, targetDate time.Time) (*types.NAVRecord, error) {
	schemeCode, err := c.getSchemeCode(isin)
	if err != nil {
		return nil, err
	}

	history, err := c.fetchNAVHistory(schemeCode)
	if err != nil {
		return nil, err
	}

	// Find NAV for the target date or closest available
	record := c.findClosestNAV(history, isin, targetDate)
	if record == nil {
		return nil, fmt.Errorf("no NAV found for ISIN %s on date %s (checked ±%d days)",
			isin, targetDate.Format("2006-01-02"), maxDayOffset)
	}

	return record, nil
}

// getSchemeCode searches for scheme code by ISIN (with caching)
func (c *Client) getSchemeCode(isin string) (string, error) {
	// Check cache first
	c.schemeMu.RLock()
	if code, exists := c.schemeCache[isin]; exists {
		c.schemeMu.RUnlock()
		return code, nil
	}
	c.schemeMu.RUnlock()

	// Search in API
	url := fmt.Sprintf("%s/search?q=%s", mfapiBaseURL, isin)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search failed with status %d", resp.StatusCode)
	}

	var results []SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", fmt.Errorf("invalid search response: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("ISIN %s not found", isin)
	}

	schemeCode := fmt.Sprintf("%d", results[0].SchemeCode)

	// Cache the result
	c.schemeMu.Lock()
	c.schemeCache[isin] = schemeCode
	c.schemeMu.Unlock()

	return schemeCode, nil
}

// fetchNAVHistory fetches the full NAV history for a scheme
func (c *Client) fetchNAVHistory(schemeCode string) (*NAVHistoryResponse, error) {
	url := fmt.Sprintf("%s/%s", mfapiBaseURL, schemeCode)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("NAV history request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NAV history request failed with status %d", resp.StatusCode)
	}

	var history NAVHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("invalid NAV history response: %w", err)
	}

	return &history, nil
}

// findClosestNAV searches for NAV closest to target date (prefer past dates)
func (c *Client) findClosestNAV(history *NAVHistoryResponse, isin string, targetDate time.Time) *types.NAVRecord {
	// Search backwards from target date, then forwards
	for offset := 0; offset <= maxDayOffset; offset++ {
		// Search backward (past dates)
		if offset == 0 {
			if record := c.findNAVForDate(history, isin, targetDate); record != nil {
				return record
			}
		} else {
			// Search backward
			pastDate := targetDate.AddDate(0, 0, -offset)
			if record := c.findNAVForDate(history, isin, pastDate); record != nil {
				return record
			}

			// Search forward (only if backward didn't find anything)
			futureDate := targetDate.AddDate(0, 0, offset)
			if record := c.findNAVForDate(history, isin, futureDate); record != nil {
				return record
			}
		}
	}

	return nil
}

// findNAVForDate finds NAV for a specific date in history
func (c *Client) findNAVForDate(history *NAVHistoryResponse, isin string, date time.Time) *types.NAVRecord {
	targetDateStr := date.Format("02-Jan-2006")

	for _, entry := range history.Data {
		if entry.Date == targetDateStr && entry.NAV != "" && entry.NAV != "null" {
			navVal, err := strconv.ParseFloat(entry.NAV, 64)
			if err != nil {
				continue
			}

			return &types.NAVRecord{
				SchemeCode: fmt.Sprintf("%d", history.Meta.SchemeCode),
				ISIN:       isin,
				SchemeName: history.Meta.SchemeName,
				NAV:        navVal,
				Date:       date,
			}
		}
	}

	return nil
}
