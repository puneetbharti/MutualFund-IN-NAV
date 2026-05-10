package amfi

import (
	"strings"
	"testing"
	"time"
)

func TestParseAMFIResponse(t *testing.T) {
	body := `Scheme Code;Scheme Name;ISIN Div Payout/ISIN Growth;ISIN Div Reinvestment;Net Asset Value;Repurchase Price;Sale Price;Date
100001;Axis Focused 25 Fund - Gr;INF846K01CH7;INF846K01CI5;39.9169;39.9169;39.9169;02-Jan-2024
100002;HDFC Small Cap Fund - Gr;INF179KA1RZ8;;82.7580;82.7580;82.7580;02-Jan-2024
`

	result, err := parseAMFIResponse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both ISINs from the first row should be present
	growth, ok := result["INF846K01CH7"]
	if !ok {
		t.Fatal("expected growth ISIN INF846K01CH7 in result")
	}
	div, ok := result["INF846K01CI5"]
	if !ok {
		t.Fatal("expected dividend ISIN INF846K01CI5 in result")
	}

	// Pointer aliasing fix: each entry must carry its own ISIN
	if growth.ISIN != "INF846K01CH7" {
		t.Errorf("growth ISIN field = %q, want INF846K01CH7", growth.ISIN)
	}
	if div.ISIN != "INF846K01CI5" {
		t.Errorf("div ISIN field = %q, want INF846K01CI5", div.ISIN)
	}

	// They must be distinct pointers
	if growth == div {
		t.Error("growth and dividend records must be separate structs, got same pointer")
	}

	// NAV and scheme name
	if growth.NAV != 39.9169 {
		t.Errorf("NAV = %v, want 39.9169", growth.NAV)
	}
	if growth.SchemeName != "Axis Focused 25 Fund - Gr" {
		t.Errorf("SchemeName = %q, want \"Axis Focused 25 Fund - Gr\"", growth.SchemeName)
	}

	wantDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if !growth.Date.Equal(wantDate) {
		t.Errorf("Date = %v, want %v", growth.Date, wantDate)
	}

	// Fund with no dividend ISIN
	hdfc, ok := result["INF179KA1RZ8"]
	if !ok {
		t.Fatal("expected ISIN INF179KA1RZ8 in result")
	}
	if hdfc.NAV != 82.7580 {
		t.Errorf("HDFC NAV = %v, want 82.7580", hdfc.NAV)
	}
}

func TestParseAMFIResponse_SkipsInvalidLines(t *testing.T) {
	body := `Scheme Code;Scheme Name;ISIN Div Payout/ISIN Growth;ISIN Div Reinvestment;Net Asset Value;Repurchase Price;Sale Price;Date
bad line with too few fields
100001;Axis Fund;INF846K01CH7;;not-a-float;0;0;02-Jan-2024
100002;HDFC Fund;INF179KA1RZ8;;82.7580;0;0;bad-date
100003;Valid Fund;INF000K01000;;50.0000;50.0000;50.0000;02-Jan-2024
`

	result, err := parseAMFIResponse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the valid row should be present
	if len(result) != 1 {
		t.Errorf("expected 1 valid record, got %d", len(result))
	}
	if _, ok := result["INF000K01000"]; !ok {
		t.Error("expected INF000K01000 in result")
	}
}

func TestParseAMFIResponse_Empty(t *testing.T) {
	result, err := parseAMFIResponse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d records", len(result))
	}
}

func TestFormatAMFIDate(t *testing.T) {
	tests := []struct {
		input time.Time
		want  string
	}{
		{time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), "02-Jan-2024"},
		{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), "31-Dec-2024"},
	}
	for _, tc := range tests {
		got := formatAMFIDate(tc.input)
		if got != tc.want {
			t.Errorf("formatAMFIDate(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
