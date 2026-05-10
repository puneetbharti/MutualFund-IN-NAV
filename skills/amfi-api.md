# Skill: AMFI NAV History API

## Overview
The Association of Mutual Funds in India (AMFI) provides a public NAV history API that returns semicolon-separated plain text data for mutual funds.

## API Endpoint

**Base URL:** `https://portal.amfiindia.com/DownloadNAVHistoryReport_Po.aspx`

## Query Parameters

| Parameter | Type   | Required | Description                      | Example        |
|-----------|--------|----------|----------------------------------|-----------------|
| `tp`      | int    | Yes      | Report type (always `1`)         | `1`             |
| `frmdt`   | string | Yes      | From date in `DD-Mon-YYYY` format| `02-Jan-2024`   |
| `todt`    | string | Yes      | To date in `DD-Mon-YYYY` format  | `02-Jan-2024`   |
| `Sch`     | string | No       | Scheme code (optional filter)    | `119551`        |

## Date Format Rules

- Always use **`DD-Mon-YYYY`** format where Mon is the 3-letter month abbreviation (Jan, Feb, Mar, etc.)
- Month abbreviations must be capitalized
- Day must be zero-padded (e.g., `02` not `2`)
- Year must be 4 digits

## Response Format

Plain text with semicolon-separated fields. Header row:
```
Scheme Code;ISIN Div Payout/ ISIN Growth;ISIN Div Reinvestment;Scheme Name;Net Asset Value;Repurchase Price;Sale Price;Date
```

Data rows follow the same format. Example:
```
119551;INF846K01CH7;INF846K01CH7;Axis Focused 25 Fund - Gr;39.9169;39.8838;39.9500;02-Jan-2024
```

## Column Reference

| Index | Field Name                        | Description                    |
|-------|-----------------------------------|--------------------------------|
| 0     | Scheme Code                       | AMFI scheme code               |
| 1     | ISIN Div Payout / ISIN Growth     | ISIN for growth option         |
| 2     | ISIN Div Reinvestment             | ISIN for dividend reinvestment |
| 3     | Scheme Name                       | Fund name                      |
| 4     | Net Asset Value                   | NAV as float64                 |
| 5     | Repurchase Price                  | Buyback price                  |
| 6     | Sale Price                        | Selling price                  |
| 7     | Date                              | NAV date in `DD-Mon-YYYY`      |

## Filtering by ISIN

To filter results by ISIN:
1. Parse each data row by splitting on `;`
2. Check if `fields[1]` (ISIN Growth) or `fields[2]` (ISIN Div Reinvestment) matches the target ISIN
3. Extract NAV from `fields[4]` and parse as float64
4. Extract date from `fields[7]` and parse as time.Time

## Market Holiday Behavior

- **No trading on market holidays** — the AMFI API returns no NAV data for 1 Jan, 26 Jan (Republic Day), 15 Aug (Independence Day), etc.
- **Strategy:** Request data for the same date; if empty, retry for the next trading day (scan forward by 1 day at a time)
- **Example:** For 1 Jan 2024 (holiday), request NAV for 2 Jan 2024 instead

## Example Request

### Single date
```bash
curl "https://portal.amfiindia.com/DownloadNAVHistoryReport_Po.aspx?tp=1&frmdt=02-Jan-2024&todt=02-Jan-2024"
```

### Response excerpt
```
Scheme Code;ISIN Div Payout/ ISIN Growth;ISIN Div Reinvestment;Scheme Name;Net Asset Value;Repurchase Price;Sale Price;Date
119551;INF846K01CH7;INF846K01CH7;Axis Focused 25 Fund - Gr;39.9169;39.8838;39.9500;02-Jan-2024
120209;INF179KA1RZ8;INF179KA1RZ8;HDFC Small Cap Fund - Gr;82.7580;82.4980;83.0190;02-Jan-2024
...
```

## Implementation Notes

### Parsing
- Skip the header row (detect by checking if first field is "Scheme Code")
- Trim whitespace from each field after splitting
- Parse NAV as `float64` using `strconv.ParseFloat`
- Parse date as `time.Time` using `time.Parse("02-Jan-2006", dateStr)`

### Error Handling
- Return `ErrNoNAVForDate` if no data rows are returned or no ISIN match
- Return `ErrHTTPTimeout` if request takes > 15 seconds
- Log HTTP response status codes for debugging

### Performance
- Prefer fetching a range (e.g., 02-Jan-2024 to 02-Jan-2024) rather than querying per fund
- The optional `Sch` parameter can filter on the server side but is not required
- **Do not use the `Sch` parameter** — always fetch full date range and filter by ISIN client-side

## Reference ISINs (Test Data)

These ISINs were used to verify the tool:
```
INF846K01CH7  Axis Focused 25 Fund - Gr
INF179KA1RZ8  HDFC Small Cap Fund - Gr
INF767K01139  Motilal Oswal Multi-Asset Fund - Gr
INF769K01EY2  SBI Emerging Businesses Fund - Gr
INF247L01700  Mirae Asset Large Cap Fund - Gr
INF247L01908  Mirae Asset Hybrid Balanced Fund - Gr
INF247L01AH0  Mirae Asset Overnight Fund - Gr
INF200K01537  ICICI Prudential Bluechip Fund - Gr
```
