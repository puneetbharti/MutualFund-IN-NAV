package types

import "time"

// NAVRecord represents a NAV data point from AMFI
type NAVRecord struct {
	SchemeCode      string
	ISIN            string
	SchemeName      string
	NAV             float64
	Date            time.Time
	RepurchasePrice float64
	SalePrice       float64
}

// NAVResult represents the final enriched NAV data with EUR conversion
type NAVResult struct {
	FundName      string
	ISIN          string
	RequestedDate time.Time
	ActualNAVDate time.Time
	NAVINR        float64
	EURINRRate    float64
	EURRateDate   time.Time
	NAVEUR        float64
	Source        string // "AMFI" or "MFapi"
	Error         error
}

// ExchangeRate represents an ECB exchange rate
type ExchangeRate struct {
	Date      time.Time
	INRperEUR float64
}
