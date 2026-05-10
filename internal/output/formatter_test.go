package output

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/puneetbharti/navcalc/internal/types"
)

var (
	jan2  = time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	jan3  = time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
)

func goodResult() *types.NAVResult {
	return &types.NAVResult{
		FundName:      "Axis Focused 25 Fund - Gr",
		ISIN:          "INF846K01CH7",
		RequestedDate: jan2,
		ActualNAVDate: jan2,
		NAVINR:        39.9169,
		EURINRRate:    91.8295,
		EURRateDate:   jan3,
		NAVEUR:        0.4347,
		Source:        "AMFI",
	}
}

func errorResult() *types.NAVResult {
	return &types.NAVResult{
		ISIN:          "INF000XXXXXX",
		RequestedDate: jan2,
		Error:         errors.New("no NAV found"),
	}
}

// --- CSV tests ---

func TestPrintCSV_Headers(t *testing.T) {
	f := NewFormatter("csv")
	var buf bytes.Buffer
	if err := f.printCSV([]*types.NAVResult{goodResult()}, &buf); err != nil {
		t.Fatalf("printCSV error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv parse error: %v", err)
	}

	wantHeaders := []string{"fund_name", "isin", "date", "nav_actual_date", "nav_inr", "eur_inr_rate", "eur_rate_date", "nav_eur", "source", "error"}
	if len(records) == 0 {
		t.Fatal("no records in CSV output")
	}
	for i, h := range wantHeaders {
		if i >= len(records[0]) || records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}
}

func TestPrintCSV_DataRow(t *testing.T) {
	f := NewFormatter("csv")
	var buf bytes.Buffer
	if err := f.printCSV([]*types.NAVResult{goodResult()}, &buf); err != nil {
		t.Fatalf("printCSV error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, _ := r.ReadAll()
	if len(records) < 2 {
		t.Fatalf("expected header + 1 data row, got %d rows", len(records))
	}

	row := records[1]
	checks := map[int]string{
		0: "Axis Focused 25 Fund - Gr",
		1: "INF846K01CH7",
		2: "2024-01-02",
		3: "2024-01-02",
		4: "39.9169",
		5: "91.8295",
		6: "2024-01-03",
		7: "0.4347",
		8: "AMFI",
		9: "",
	}
	for col, want := range checks {
		if col >= len(row) {
			t.Errorf("column %d missing", col)
			continue
		}
		if row[col] != want {
			t.Errorf("col %d = %q, want %q", col, row[col], want)
		}
	}
}

func TestPrintCSV_ErrorRow(t *testing.T) {
	f := NewFormatter("csv")
	var buf bytes.Buffer
	if err := f.printCSV([]*types.NAVResult{errorResult()}, &buf); err != nil {
		t.Fatalf("printCSV error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, _ := r.ReadAll()
	if len(records) < 2 {
		t.Fatal("expected header + 1 row")
	}

	row := records[1]
	// error column (index 9) should contain the error message
	if row[9] != "no NAV found" {
		t.Errorf("error column = %q, want \"no NAV found\"", row[9])
	}
	// source column (index 8) should be empty for error rows
	if row[8] != "" {
		t.Errorf("source column for error row = %q, want empty", row[8])
	}
}

func TestPrintCSV_MultipleRows(t *testing.T) {
	f := NewFormatter("csv")
	var buf bytes.Buffer
	results := []*types.NAVResult{goodResult(), errorResult()}
	if err := f.printCSV(results, &buf); err != nil {
		t.Fatalf("printCSV error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, _ := r.ReadAll()
	// header + 2 data rows
	if len(records) != 3 {
		t.Errorf("expected 3 rows (header + 2 data), got %d", len(records))
	}
}

// --- formatFloat tests ---

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		val      float64
		decimals int
		want     string
	}{
		{0, 4, "—"},
		{39.9169, 4, "39.9169"},
		{91.8295, 4, "91.8295"},
		{0.4347, 4, "0.4347"},
		{100.0, 2, "100.00"},
	}
	for _, tc := range tests {
		got := formatFloat(tc.val, tc.decimals)
		if got != tc.want {
			t.Errorf("formatFloat(%v, %d) = %q, want %q", tc.val, tc.decimals, got, tc.want)
		}
	}
}

// --- formatter dispatch ---

func TestNewFormatter_UnsupportedFormat(t *testing.T) {
	f := NewFormatter("xml")
	err := f.PrintResults([]*types.NAVResult{goodResult()})
	if err == nil {
		t.Error("expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error message = %q, expected to mention format name", err.Error())
	}
}
