package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/puneetbharti/navcalc/internal/nav"
	"github.com/puneetbharti/navcalc/internal/output"
)

var (
	dates     []string
	isins     []string
	isinFile  string
	outputFmt string
	verbose   bool
)

const version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:     "navcalc",
	Version: version,
	Short:   "Fetch and display Indian mutual fund NAV data with EUR conversion",
	Long: `navcalc is a CLI tool for fetching NAV (Net Asset Value) data for Indian mutual funds.

Features:
- Fetches NAV from AMFI (primary) with MFapi.in fallback
- Supports multiple ISINs and dates
- Converts NAV to EUR using ECB exchange rates
- Handles market holidays gracefully
- Outputs as table or CSV`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := validateFlags(); err != nil {
			log.Fatalf("Error: %v", err)
		}

		if err := execute(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.Flags().StringSliceVar(&dates, "date", []string{}, "Date(s) in YYYY-MM-DD format (can be specified multiple times)")
	rootCmd.Flags().StringSliceVar(&isins, "isin", []string{}, "ISIN(s) to fetch (can be specified multiple times)")
	rootCmd.Flags().StringVar(&isinFile, "isin-file", "", "File containing ISINs (one per line)")
	rootCmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or csv")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	rootCmd.MarkFlagRequired("date")
}

func validateFlags() error {
	if len(dates) == 0 {
		return fmt.Errorf("at least one date is required")
	}

	if len(isins) == 0 && isinFile == "" {
		return fmt.Errorf("at least one ISIN or --isin-file is required")
	}

	if outputFmt != "table" && outputFmt != "csv" {
		return fmt.Errorf("invalid output format: %s (must be 'table' or 'csv')", outputFmt)
	}

	return nil
}

func execute() error {
	// Parse dates
	parsedDates, err := parseDates(dates)
	if err != nil {
		return err
	}

	// Load ISINs
	parsedISINs, err := loadISINs()
	if err != nil {
		return err
	}

	if !verbose {
		log.SetOutput(io.Discard)
	}

	// Resolve NAVs
	resolver := nav.NewResolver()
	results := resolver.ResolveNAVBatch(parsedISINs, parsedDates)

	// Format and print output
	formatter := output.NewFormatter(outputFmt)
	if err := formatter.PrintResults(results); err != nil {
		return err
	}

	return nil
}

func parseDates(dateStrs []string) ([]time.Time, error) {
	var result []time.Time
	for _, dateStr := range dateStrs {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateStr)
		}
		result = append(result, t)
	}
	return result, nil
}

func loadISINs() ([]string, error) {
	var result []string

	// Add ISINs from --isin flags
	result = append(result, isins...)

	// Load from file if specified
	if isinFile != "" {
		f, err := os.Open(isinFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open ISIN file: %w", err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				result = append(result, line)
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading ISIN file: %w", err)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no ISINs found")
	}

	return result, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
