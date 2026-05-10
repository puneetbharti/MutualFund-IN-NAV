# navcalc – Indian Mutual Fund NAV Calculator

A production-quality Go CLI tool for fetching and displaying Net Asset Value (NAV) data for Indian mutual funds with EUR conversion.

## Features

- **Multi-source NAV fetching**: AMFI (primary) with automatic fallback to MFapi.in
- **Multi-date support**: Fetch NAV for multiple dates in a single query
- **EUR conversion**: Automatically converts INR NAV to EUR using ECB exchange rates
- **Market holiday handling**: Gracefully handles Indian market holidays by finding nearest trading day
- **Flexible output**: Table format (human-readable) or CSV (machine-readable)
- **Efficient caching**: In-memory caching of scheme codes and exchange rates to minimize API calls

## Installation

### From Source

```bash
git clone https://github.com/puneetbharti/navcalc.git
cd navcalc
go build -o navcalc ./cmd/navcalc
```

Or install directly:

```bash
go install github.com/puneetbharti/navcalc/cmd/navcalc@latest
```

### Using Make

```bash
make build
make install
```

## Quick Start

### Basic Usage

```bash
# Fetch NAV for two ISINs on a specific date
navcalc --date 2024-01-02 --isin INF846K01CH7 --isin INF179KA1RZ8

# Fetch NAV for multiple dates
navcalc --date 2024-01-02 --date 2024-12-31 --isin INF846K01CH7

# Output as CSV
navcalc --date 2024-01-02 --isin INF846K01CH7 --output csv

# Load ISINs from a file
navcalc --date 2024-01-02 --isin-file my-funds.txt --output csv > output.csv
```

### Full Example

```bash
navcalc \
  --date 2024-01-02 \
  --date 2024-12-31 \
  --isin INF846K01CH7 \
  --isin INF179KA1RZ8 \
  --isin INF767K01139 \
  --isin INF769K01EY2 \
  --isin INF247L01700 \
  --isin INF247L01908 \
  --isin INF247L01AH0 \
  --isin INF200K01537 \
  --output table
```

## CLI Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--date` | | string | Date(s) in YYYY-MM-DD format (can be specified multiple times, required) |
| `--isin` | | string | ISIN code(s) (can be specified multiple times) |
| `--isin-file` | | string | Path to file with ISINs (one per line) |
| `--output` | | string | Output format: `table` or `csv` (default: `table`) |
| `--verbose` | | bool | Enable verbose logging for debugging |
| `--help` | | | Show help message |
| `--version` | | | Show version |

## Output Formats

### Table Format (Default)

Human-readable ASCII table:

```
┌─────────────────────────────────────┬────────────────┬────────────┬────────────┬──────────────┬────────────┐
│ Fund Name                           │ ISIN           │ Date       │ NAV (INR)  │ EUR/INR Rate │ NAV (EUR)  │
├─────────────────────────────────────┼────────────────┼────────────┼────────────┼──────────────┼────────────┤
│ Axis Focused 25 Fund - Gr           │ INF846K01CH7   │ 2024-01-02 │   39.9169  │    91.8295   │    0.4347  │
│ HDFC Small Cap Fund - Gr            │ INF179KA1RZ8   │ 2024-01-02 │   82.7580  │    91.8295   │    0.9010  │
└─────────────────────────────────────┴────────────────┴────────────┴────────────┴──────────────┴────────────┘
```

### CSV Format

Machine-readable comma-separated values:

```csv
fund_name,isin,date,nav_actual_date,nav_inr,eur_inr_rate,eur_rate_date,nav_eur,error
"Axis Focused 25 Fund - Gr",INF846K01CH7,2024-01-02,2024-01-02,39.9169,91.8295,2024-01-02,0.4347,
"HDFC Small Cap Fund - Gr",INF179KA1RZ8,2024-01-02,2024-01-02,82.7580,91.8295,2024-01-02,0.9010,
```

## Data Sources

### AMFI (Association of Mutual Funds in India)

- **Primary source** for NAV data
- **URL**: https://portal.amfiindia.com/DownloadNAVHistoryReport_Po.aspx
- **Reliability**: Official, authoritative source
- **Coverage**: Daily NAV for all AMFI-registered mutual funds
- **Fallback behavior**: If AMFI unavailable, automatically falls back to MFapi.in

### MFapi.in

- **Fallback source** for NAV data
- **URL**: https://api.mfapi.in
- **Reliability**: Community-maintained, generally consistent with AMFI
- **When used**: Only if AMFI request fails or times out (>15 seconds)

### ECB (European Central Bank)

- **Source** for EUR/INR exchange rates
- **URL**: https://data-api.ecb.europa.eu
- **Rate type**: Spot exchange rates (official ECB daily rates)
- **Frequency**: Working days (Monday–Friday)
- **Fallback behavior**: If exact date unavailable, uses nearest prior trading day rate

## Market Holidays & Special Dates

The tool handles market holidays automatically:

- **No trading on holidays**: Returns data from the next available trading day
- **Weekends**: Automatically skips to the next Monday
- **Examples**:
  - 1 Jan 2024 (New Year) → Uses NAV from 2 Jan 2024
  - 26 Jan 2024 (Republic Day) → Uses NAV from 25 Jan or 29 Jan
  - 2024-01-06 (Saturday) → ECB rate uses Friday's rate (2024-01-05)

## Makefile Targets

```bash
make build            # Build the binary
make test             # Run tests
make run-example      # Run with example data (8 funds, 2 dates)
make run-example-csv  # Generate example CSV output
make lint             # Run go vet
make clean            # Remove built binary
make install          # Install to GOPATH/bin
```

## Verification

To verify the tool against live API data, use the included `verify.sh` script:

```bash
chmod +x verify.sh
./verify.sh
```

This script queries AMFI and ECB directly and greps for the test ISINs, allowing you to manually cross-check the tool's output.

## Error Handling

The tool provides graceful error handling:

- **Per-fund errors**: Failure for one fund doesn't stop others
- **Clear error messages**: "No NAV found for ISIN X on date Y (checked ±7 days)"
- **HTTP timeouts**: 15-second timeout per request with single retry on transient errors (5xx)
- **CSV error column**: CSV output includes an `error` column for any per-fund failures

## Performance

- **Caching**: Scheme codes and exchange rates cached in-memory
- **Batch requests**: Multiple ISINs and dates processed efficiently
- **Parallel independence**: Different API clients can fetch independently (no shared state except caching)

## Project Structure

```
navcalc/
├── cmd/navcalc/              # CLI entry point
│   └── main.go
├── internal/
│   ├── amfi/                 # AMFI API client
│   │   └── client.go
│   ├── mfapi/                # MFapi.in fallback client
│   │   └── client.go
│   ├── ecb/                  # ECB exchange rate client
│   │   └── client.go
│   ├── nav/                  # NAV resolver and types
│   │   ├── types.go
│   │   └── resolver.go
│   └── output/               # Output formatting
│       └── formatter.go
├── skills/                   # Skill documentation
│   ├── amfi-api.md
│   ├── mfapi-fallback.md
│   └── ecb-exchange-rate.md
├── go.mod                    # Go module file
├── go.sum                    # Go module dependencies
├── Makefile                  # Build targets
├── verify.sh                 # Verification script
└── README.md                 # This file
```

## Dependencies

```
github.com/spf13/cobra@v1.7.0              # CLI framework
github.com/olekukonko/tablewriter@v0.0.5   # Table formatting
```

No external databases, ORMs, or heavy frameworks required.

## License

MIT

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## Support

For issues, questions, or suggestions, please open an issue on the GitHub repository.

---

**Last Updated**: 10 May 2026
