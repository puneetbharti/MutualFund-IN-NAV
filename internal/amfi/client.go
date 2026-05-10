package amfi

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/puneetbharti/navcalc/internal/types"
)

const (
	amfiBaseURL = "https://portal.amfiindia.com/DownloadNAVHistoryReport_Po.aspx"
	httpTimeout = 15 * time.Second
)

// Client for AMFI API
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new AMFI client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// FetchNAV fetches NAV data for a specific date range
// Returns a map of ISIN -> NAVRecord
func (c *Client) FetchNAV(fromDate, toDate time.Time) (map[string]*types.NAVRecord, error) {
	urlStr := fmt.Sprintf("%s?tp=1&frmdt=%s&todt=%s",
		amfiBaseURL,
		formatAMFIDate(fromDate),
		formatAMFIDate(toDate),
	)

	resp, err := c.httpClient.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return parseAMFIResponse(resp.Body)
}

// FetchNAVForISIN fetches NAV for a specific ISIN on a given date
func (c *Client) FetchNAVForISIN(isin string, targetDate time.Time) (*types.NAVRecord, error) {
	navMap, err := c.FetchNAV(targetDate, targetDate)
	if err != nil {
		return nil, err
	}

	if record, exists := navMap[isin]; exists && record != nil {
		return record, nil
	}

	return nil, fmt.Errorf("no NAV data found for ISIN %s on %s", isin, targetDate.Format("2006-01-02"))
}

// amfiCols holds column indices resolved from the AMFI header row.
type amfiCols struct {
	schemeCode, schemeName, isinGrowth, isinDiv, nav, repurchase, sale, date int
}

// parseAMFIResponse parses the semicolon-delimited AMFI response.
// Column positions are resolved from the header row by name so the parser
// stays correct if AMFI reorders columns in future.
func parseAMFIResponse(reader io.Reader) (map[string]*types.NAVRecord, error) {
	result := make(map[string]*types.NAVRecord)
	scanner := bufio.NewScanner(reader)

	var cols *amfiCols

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, ";")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		// Resolve column positions from the header row
		if cols == nil {
			if !strings.Contains(line, "Scheme Code") {
				continue
			}
			idx := func(name string) int {
				for i, f := range fields {
					if strings.Contains(f, name) {
						return i
					}
				}
				return -1
			}
			cols = &amfiCols{
				schemeCode: idx("Scheme Code"),
				schemeName: idx("Scheme Name"),
				isinGrowth: idx("ISIN Div Payout"),
				isinDiv:    idx("ISIN Div Reinvestment"),
				nav:        idx("Net Asset Value"),
				repurchase: idx("Repurchase"),
				sale:       idx("Sale Price"),
				date:       idx("Date"),
			}
			continue
		}

		if len(fields) < 8 {
			continue
		}

		get := func(i int) string {
			if i < 0 || i >= len(fields) {
				return ""
			}
			return fields[i]
		}

		navVal, err := strconv.ParseFloat(get(cols.nav), 64)
		if err != nil {
			continue
		}

		parsedDate, err := time.Parse("02-Jan-2006", get(cols.date))
		if err != nil {
			continue
		}

		repurchaseVal, _ := strconv.ParseFloat(get(cols.repurchase), 64)
		saleVal, _ := strconv.ParseFloat(get(cols.sale), 64)

		base := types.NAVRecord{
			SchemeCode:      get(cols.schemeCode),
			SchemeName:      get(cols.schemeName),
			NAV:             navVal,
			Date:            parsedDate,
			RepurchasePrice: repurchaseVal,
			SalePrice:       saleVal,
		}

		// Store by both ISIN types — clone the struct for each so mutations don't alias.
		isinGrowth := get(cols.isinGrowth)
		isinDiv := get(cols.isinDiv)

		if isinGrowth != "" {
			r := base
			r.ISIN = isinGrowth
			result[isinGrowth] = &r
		}
		if isinDiv != "" && isinDiv != isinGrowth {
			r := base
			r.ISIN = isinDiv
			result[isinDiv] = &r
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return result, nil
}

// formatAMFIDate formats a time.Time to AMFI format (DD-Mon-YYYY)
func formatAMFIDate(t time.Time) string {
	return t.Format("02-Jan-2006")
}
