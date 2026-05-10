package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/puneetbharti/navcalc/internal/types"
)

// Formatter handles output formatting
type Formatter struct {
	format string // "table" or "csv"
}

// NewFormatter creates a new formatter
func NewFormatter(format string) *Formatter {
	return &Formatter{
		format: format,
	}
}

// PrintResults prints NAV results in the specified format
func (f *Formatter) PrintResults(results []*types.NAVResult) error {
	switch f.format {
	case "csv":
		return f.printCSV(results, os.Stdout)
	case "table":
		return f.printTable(results)
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}
}

// printTable prints results as an ASCII table
func (f *Formatter) printTable(results []*types.NAVResult) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Fund Name",
		"ISIN",
		"Date",
		"NAV (INR)",
		"EUR/INR Rate",
		"NAV (EUR)",
		"Source",
	})

	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetCenterSeparator("|")
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_CENTER,
	})

	for _, result := range results {
		if result.Error != nil {
			table.Append([]string{
				"ERROR",
				result.ISIN,
				result.RequestedDate.Format("2006-01-02"),
				"—",
				"—",
				fmt.Sprintf("Error: %v", result.Error),
				"—",
			})
			continue
		}

		table.Append([]string{
			result.FundName,
			result.ISIN,
			result.ActualNAVDate.Format("2006-01-02"),
			formatFloat(result.NAVINR, 4),
			formatFloat(result.EURINRRate, 4),
			formatFloat(result.NAVEUR, 4),
			result.Source,
		})
	}

	table.Render()
	return nil
}

// printCSV prints results as CSV
func (f *Formatter) printCSV(results []*types.NAVResult, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"fund_name",
		"isin",
		"date",
		"nav_actual_date",
		"nav_inr",
		"eur_inr_rate",
		"eur_rate_date",
		"nav_eur",
		"source",
		"error",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.FundName,
			result.ISIN,
			result.RequestedDate.Format("2006-01-02"),
			result.ActualNAVDate.Format("2006-01-02"),
			formatFloat(result.NAVINR, 4),
			formatFloat(result.EURINRRate, 4),
			result.EURRateDate.Format("2006-01-02"),
			formatFloat(result.NAVEUR, 4),
			result.Source,
			"",
		}

		if result.Error != nil {
			row[9] = result.Error.Error()
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// formatFloat formats a float64 with specified decimal places
func formatFloat(val float64, decimals int) string {
	if val == 0 {
		return "—"
	}
	format := "%." + strconv.Itoa(decimals) + "f"
	return fmt.Sprintf(format, val)
}
