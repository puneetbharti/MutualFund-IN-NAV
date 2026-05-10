# Skill: MFapi.in Fallback API

## Overview
MFapi.in provides a REST API for fetching Indian mutual fund NAV data. This is used as a **fallback** when the AMFI API returns no data for a specific date.

## API Endpoints

### 1. Search for Scheme Code by ISIN

**Endpoint:** `GET https://api.mfapi.in/mf/search`

**Query Parameters:**

| Parameter | Type   | Description      | Example            |
|-----------|--------|------------------|--------------------|
| `q`       | string | Search query     | `INF846K01CH7`     |

**Response:** JSON array of matching schemes

```json
[
  {
    "schemeCode": "119551",
    "schemeName": "Axis Focused 25 Fund - Gr",
    "isin": "INF846K01CH7"
  }
]
```

### 2. Fetch NAV History

**Endpoint:** `GET https://api.mfapi.in/mf/<schemeCode>`

**Path Parameters:**

| Parameter    | Type   | Description           | Example |
|--------------|--------|------------------------|---------|
| `schemeCode` | string | AMFI scheme code       | `119551` |

**Response:** JSON object with metadata and NAV history

```json
{
  "meta": {
    "fund_house": "Axis Mutual Fund",
    "scheme_type": "Growth",
    "scheme_code": "119551",
    "scheme_name": "Axis Focused 25 Fund - Gr",
    "isin": "INF846K01CH7"
  },
  "status": "success",
  "data": [
    {
      "date": "02-Jan-2024",
      "nav": "39.9169"
    },
    {
      "date": "01-Jan-2024",
      "nav": null
    },
    {
      "date": "31-Dec-2023",
      "nav": "38.5425"
    }
  ]
}
```

## Date Format

- Dates in response are in **`DD-Mon-YYYY`** format (e.g., `02-Jan-2024`)
- NAV is returned as a string; parse as `float64`
- A `null` NAV indicates no trading on that date (market holiday or weekend)

## Finding Closest Trading Day

When requested date has no NAV:
1. Search the `data` array for the requested date
2. If exact date not found, search within **±7 days**
3. **Prefer past dates** (search backwards from the requested date)
4. Return the first match with a non-null NAV

**Example logic:**
```
requestedDate = 02-Jan-2024
if not found in data:
  search 01-Jan-2024, 31-Dec-2023, 30-Dec-2023, ... (backwards)
  search 03-Jan-2024, 04-Jan-2024, ... (forwards if still not found)
```

## Caching Strategy

### Scheme Code Cache

To avoid repeated API calls for the same ISIN:
1. Maintain an **in-memory map**: `ISIN → SchemeCode`
2. Before searching, check the cache
3. On successful search, store the result
4. Cache duration: entire application lifetime (within-process)

**Implementation:**
```go
var schemeCache = make(map[string]string) // ISIN -> SchemeCode
```

## Error Handling

| Scenario | Status Code | Action |
|----------|-------------|--------|
| ISIN not found | 404 or empty array | Return `ErrISINNotFound` |
| Scheme code not found | 404 | Return `ErrSchemeNotFound` |
| Network timeout | N/A | Retry once, then fail |
| Invalid JSON | N/A | Return `ErrInvalidResponse` |

## HTTP Timeout

Set a **15-second timeout** for all MFapi requests.

## Example Workflow

1. **Search for ISIN:**
   ```bash
   curl "https://api.mfapi.in/mf/search?q=INF846K01CH7"
   ```
   Response: `[{ "schemeCode": "119551", ... }]`

2. **Fetch NAV history for scheme:**
   ```bash
   curl "https://api.mfapi.in/mf/119551"
   ```
   Response: Full NAV history (array of 2000+ records)

3. **Find NAV for requested date:**
   - Look for `"date": "02-Jan-2024"` in the data array
   - Extract `"nav": "39.9169"` and parse as float64

## Rate Limiting

- No official rate limit documented
- Implement **caching aggressively** to minimize requests
- If 429 (Too Many Requests) is received, wait 60 seconds and retry once

## Fallback Trigger

Use MFapi as fallback **only when:**
- AMFI returned no data (empty result or HTTP error)
- Or AMFI timeout exceeded (15 seconds)

Always try AMFI first before falling back to MFapi.

## Reference ISINs

```
INF846K01CH7  Axis Focused 25 Fund - Gr (Scheme: 119551)
INF179KA1RZ8  HDFC Small Cap Fund - Gr (Scheme: 119209)
INF767K01139  Motilal Oswal Multi-Asset Fund - Gr (Scheme: 120357)
INF769K01EY2  SBI Emerging Businesses Fund - Gr (Scheme: 119644)
INF247L01700  Mirae Asset Large Cap Fund - Gr (Scheme: 120358)
INF247L01908  Mirae Asset Hybrid Balanced Fund - Gr (Scheme: 119598)
INF247L01AH0  Mirae Asset Overnight Fund - Gr (Scheme: 120766)
INF200K01537  ICICI Prudential Bluechip Fund - Gr (Scheme: 119019)
```
