package mfapi

import (
	"testing"
	"time"
)

func makeHistory(entries []struct{ date, nav string }) *NAVHistoryResponse {
	h := &NAVHistoryResponse{}
	h.Meta.SchemeCode = 100001
	h.Meta.SchemeName = "Test Fund - Gr"
	h.Meta.ISIN = "INF000K01000"
	for _, e := range entries {
		h.Data = append(h.Data, struct {
			Date string `json:"date"`
			NAV  string `json:"nav"`
		}{Date: e.date, NAV: e.nav})
	}
	return h
}

func TestFindClosestNAV_ExactMatch(t *testing.T) {
	history := makeHistory([]struct{ date, nav string }{
		{"02-Jan-2024", "50.0000"},
	})
	c := NewClient()
	target := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	record := c.findClosestNAV(history, "INF000K01000", target)
	if record == nil {
		t.Fatal("expected record, got nil")
	}
	if record.NAV != 50.0 {
		t.Errorf("NAV = %v, want 50.0", record.NAV)
	}
	if !record.Date.Equal(target) {
		t.Errorf("Date = %v, want %v", record.Date, target)
	}
}

func TestFindClosestNAV_FallsBackToPastDate(t *testing.T) {
	// Target is 2024-01-01 (holiday), nearest available is 2023-12-29
	history := makeHistory([]struct{ date, nav string }{
		{"29-Dec-2023", "48.0000"},
	})
	c := NewClient()
	target := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	record := c.findClosestNAV(history, "INF000K01000", target)
	if record == nil {
		t.Fatal("expected record, got nil")
	}
	if record.NAV != 48.0 {
		t.Errorf("NAV = %v, want 48.0", record.NAV)
	}
}

func TestFindClosestNAV_NoneWithinWindow(t *testing.T) {
	// Available date is 30 days away — beyond maxDayOffset
	history := makeHistory([]struct{ date, nav string }{
		{"01-Dec-2023", "45.0000"},
	})
	c := NewClient()
	target := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	record := c.findClosestNAV(history, "INF000K01000", target)
	if record != nil {
		t.Errorf("expected nil record for out-of-window date, got %+v", record)
	}
}

func TestFindClosestNAV_SkipsNullNAV(t *testing.T) {
	history := makeHistory([]struct{ date, nav string }{
		{"02-Jan-2024", "null"},
		{"01-Jan-2024", "50.0000"},
	})
	c := NewClient()
	target := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	record := c.findClosestNAV(history, "INF000K01000", target)
	if record == nil {
		t.Fatal("expected record for fallback date, got nil")
	}
	if record.NAV != 50.0 {
		t.Errorf("NAV = %v, want 50.0", record.NAV)
	}
}

func TestFindClosestNAV_EmptyHistory(t *testing.T) {
	history := makeHistory(nil)
	c := NewClient()
	target := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	record := c.findClosestNAV(history, "INF000K01000", target)
	if record != nil {
		t.Errorf("expected nil for empty history, got %+v", record)
	}
}
