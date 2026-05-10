package ecb

import (
	"strings"
	"testing"
	"time"
)

// ecbHeader matches the multi-column format the ECB API actually returns.
const ecbHeader = "KEY,FREQ,CURRENCY,CURRENCY_DENOM,EXR_TYPE,EXR_SUFFIX,TIME_PERIOD,OBS_VALUE,OBS_STATUS"

func TestParseECBResponse(t *testing.T) {
	body := ecbHeader + "\n" +
		"EXR.D.INR.EUR.SP00.A,D,INR,EUR,SP00,A,2024-01-02,91.8295,A\n"

	rate, date, err := parseECBResponse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 91.8295 {
		t.Errorf("rate = %v, want 91.8295", rate)
	}
	want := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if !date.Equal(want) {
		t.Errorf("date = %v, want %v", date, want)
	}
}

func TestParseECBResponse_Empty(t *testing.T) {
	body := ecbHeader + "\n"
	_, _, err := parseECBResponse(strings.NewReader(body))
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
}

func TestParseECBResponse_ZeroRate(t *testing.T) {
	body := ecbHeader + "\n" +
		"EXR.D.INR.EUR.SP00.A,D,INR,EUR,SP00,A,2024-01-02,0,A\n" +
		"EXR.D.INR.EUR.SP00.A,D,INR,EUR,SP00,A,2024-01-03,92.0000,A\n"

	rate, date, err := parseECBResponse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 92.0000 {
		t.Errorf("expected zero rate skipped, got %v", rate)
	}
	want := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	if !date.Equal(want) {
		t.Errorf("date = %v, want %v", date, want)
	}
}

func TestParseECBResponse_InvalidRate(t *testing.T) {
	body := ecbHeader + "\n" +
		"EXR.D.INR.EUR.SP00.A,D,INR,EUR,SP00,A,2024-01-02,not-a-number,A\n"

	_, _, err := parseECBResponse(strings.NewReader(body))
	if err == nil {
		t.Fatal("expected error for invalid rate, got nil")
	}
}

func TestParseECBResponse_MultipleRows(t *testing.T) {
	body := ecbHeader + "\n" +
		"EXR.D.INR.EUR.SP00.A,D,INR,EUR,SP00,A,2024-01-02,91.8295,A\n" +
		"EXR.D.INR.EUR.SP00.A,D,INR,EUR,SP00,A,2024-01-03,92.1000,A\n"

	rate, date, err := parseECBResponse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 91.8295 {
		t.Errorf("rate = %v, want 91.8295 (first row)", rate)
	}
	want := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if !date.Equal(want) {
		t.Errorf("date = %v, want %v", date, want)
	}
}
