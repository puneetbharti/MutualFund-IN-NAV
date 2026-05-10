package ecb

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/puneetbharti/navcalc/internal/types"
)

const (
	ecbBaseURL   = "https://data-api.ecb.europa.eu/service/data/EXR/D.INR.EUR.SP00.A"
	httpTimeout  = 15 * time.Second
	maxDayOffset = 7
)

// Client for ECB exchange rate API
type Client struct {
	httpClient *http.Client
	rateMu     sync.RWMutex
	rateCache  map[string]float64 // date (YYYY-MM-DD) -> rate
}

// NewClient creates a new ECB client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		rateCache: make(map[string]float64),
	}
}

// GetExchangeRate gets EUR/INR rate for a specific date
// Falls back to nearest prior available date if exact date unavailable
func (c *Client) GetExchangeRate(targetDate time.Time) (*types.ExchangeRate, error) {
	// Check cache first
	c.rateMu.RLock()
	if rate, exists := c.rateCache[targetDate.Format("2006-01-02")]; exists {
		c.rateMu.RUnlock()
		return &types.ExchangeRate{
			Date:      targetDate,
			INRperEUR: rate,
		}, nil
	}
	c.rateMu.RUnlock()

	// Fetch from API for a date range
	rate, actualDate, err := c.fetchExchangeRate(targetDate)
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.rateMu.Lock()
	c.rateCache[actualDate.Format("2006-01-02")] = rate
	c.rateMu.Unlock()

	return &types.ExchangeRate{
		Date:      actualDate,
		INRperEUR: rate,
	}, nil
}

// fetchExchangeRate fetches rate from ECB API
func (c *Client) fetchExchangeRate(targetDate time.Time) (float64, time.Time, error) {
	// First try exact date
	rate, actualDate, err := c.fetchRateForDate(targetDate)
	if err == nil {
		return rate, actualDate, nil
	}

	// Search backwards for available rate (prefer past dates)
	for offset := 1; offset <= maxDayOffset; offset++ {
		pastDate := targetDate.AddDate(0, 0, -offset)
		rate, actualDate, err := c.fetchRateForDate(pastDate)
		if err == nil {
			return rate, actualDate, nil
		}
	}

	return 0, time.Time{}, fmt.Errorf("no exchange rate found for date %s (checked ±%d days)",
		targetDate.Format("2006-01-02"), maxDayOffset)
}

// fetchRateForDate fetches rate for a specific date
func (c *Client) fetchRateForDate(targetDate time.Time) (float64, time.Time, error) {
	dateStr := targetDate.Format("2006-01-02")
	url := fmt.Sprintf("%s?startPeriod=%s&endPeriod=%s&format=csvdata",
		ecbBaseURL, dateStr, dateStr)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, time.Time{}, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	rate, actualDate, err := parseECBResponse(resp.Body)
	if err != nil {
		return 0, time.Time{}, err
	}

	return rate, actualDate, nil
}

// parseECBResponse parses the CSV response from ECB.
// The ECB API returns a multi-column CSV; column positions are derived from
// the header row by looking up TIME_PERIOD and OBS_VALUE by name.
func parseECBResponse(reader io.Reader) (float64, time.Time, error) {
	scanner := bufio.NewScanner(reader)

	dateCol, rateCol := -1, -1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")

		// Locate column indices from the header row
		if dateCol == -1 {
			for i, f := range fields {
				switch strings.TrimSpace(f) {
				case "TIME_PERIOD":
					dateCol = i
				case "OBS_VALUE":
					rateCol = i
				}
			}
			continue
		}

		if dateCol >= len(fields) || rateCol >= len(fields) {
			continue
		}

		dateStr := strings.TrimSpace(fields[dateCol])
		rateStr := strings.TrimSpace(fields[rateCol])

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil || rate == 0 {
			continue
		}

		return rate, date, nil
	}

	if err := scanner.Err(); err != nil {
		return 0, time.Time{}, fmt.Errorf("error reading response: %w", err)
	}

	return 0, time.Time{}, fmt.Errorf("no rate data found in response")
}
