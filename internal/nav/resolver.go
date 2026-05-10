package nav

import (
	"fmt"
	"log"
	"time"

	"github.com/puneetbharti/navcalc/internal/amfi"
	"github.com/puneetbharti/navcalc/internal/ecb"
	"github.com/puneetbharti/navcalc/internal/mfapi"
	"github.com/puneetbharti/navcalc/internal/types"
)

// Resolver orchestrates NAV lookup with fallback logic
type Resolver struct {
	amfiClient  *amfi.Client
	mfapiClient *mfapi.Client
	ecbClient   *ecb.Client
}

// NewResolver creates a new resolver
func NewResolver() *Resolver {
	return &Resolver{
		amfiClient:  amfi.NewClient(),
		mfapiClient: mfapi.NewClient(),
		ecbClient:   ecb.NewClient(),
	}
}

// ResolveNAV resolves NAV for an ISIN on a given date with EUR conversion
func (r *Resolver) ResolveNAV(isin string, targetDate time.Time) *types.NAVResult {
	result := &types.NAVResult{
		ISIN:          isin,
		RequestedDate: targetDate,
	}

	// Step 1: Try to fetch NAV from AMFI
	navRecord, err := r.amfiClient.FetchNAVForISIN(isin, targetDate)
	if err == nil && navRecord != nil {
		result.FundName = navRecord.SchemeName
		result.ActualNAVDate = navRecord.Date
		result.NAVINR = navRecord.NAV
		result.Source = "AMFI"
		log.Printf("✓ AMFI: %s (%s) on %s", isin, navRecord.SchemeName, navRecord.Date.Format("2006-01-02"))
	} else {
		// Step 2: Fall back to MFapi
		log.Printf("⚠ AMFI failed for %s: %v, trying MFapi...", isin, err)
		navRecord, err = r.mfapiClient.GetNAVForISIN(isin, targetDate)
		if err == nil && navRecord != nil {
			result.FundName = navRecord.SchemeName
			result.ActualNAVDate = navRecord.Date
			result.NAVINR = navRecord.NAV
			result.Source = "MFapi"
			log.Printf("✓ MFapi: %s (%s) on %s", isin, navRecord.SchemeName, navRecord.Date.Format("2006-01-02"))
		} else {
			result.Error = fmt.Errorf("no NAV found for ISIN %s on date %s (checked AMFI and MFapi)",
				isin, targetDate.Format("2006-01-02"))
			return result
		}
	}

	// Step 3: Fetch exchange rate for the NAV date
	rate, err := r.ecbClient.GetExchangeRate(result.ActualNAVDate)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch exchange rate: %w", err)
		return result
	}

	result.EURINRRate = rate.INRperEUR
	result.EURRateDate = rate.Date

	// Step 4: Calculate EUR NAV
	result.NAVEUR = result.NAVINR / result.EURINRRate

	return result
}

// ResolveNAVBatch resolves NAV for multiple ISINs on multiple dates
func (r *Resolver) ResolveNAVBatch(isins []string, dates []time.Time) []*types.NAVResult {
	var results []*types.NAVResult

	for _, date := range dates {
		for _, isin := range isins {
			result := r.ResolveNAV(isin, date)
			results = append(results, result)
		}
	}

	return results
}
