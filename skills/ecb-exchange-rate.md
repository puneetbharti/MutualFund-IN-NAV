# Skill: ECB Exchange Rate API

## Overview
The European Central Bank (ECB) provides a free, public data API for exchange rates. We use this to fetch EUR/INR rates for converting NAVs from INR to EUR.

## API Endpoint

**Base URL:** `https://data-api.ecb.europa.eu/service/data/EXR/D.INR.EUR.SP00.A`

This endpoint returns daily exchange rates where:
- `D` = Daily frequency
- `INR` = Indian Rupee
- `EUR` = Euro
- `SP00` = Spot exchange rate
- `A` = Average (official rate)

## Query Parameters

| Parameter     | Type   | Required | Description              | Example      |
|---------------|--------|----------|--------------------------|--------------|
| `startPeriod` | string | Yes      | Start date (ISO 8601)    | `2024-01-02` |
| `endPeriod`   | string | Yes      | End date (ISO 8601)      | `2024-01-02` |
| `format`      | string | Yes      | Response format          | `csvdata`    |

## Date Format

- Use **ISO 8601 format: `YYYY-MM-DD`** (not DD-Mon-YYYY)
- This differs from AMFI and MFapi date formats
- Always use GMT/UTC (no timezone offset needed)

## Response Format

CSV format with the following structure:

```
KEY,0
TIME_PERIOD,OBS_VALUE
2024-01-02,91.8295
```

**Parsing:**
- First line: `KEY,0` (skip)
- Second line: `TIME_PERIOD,OBS_VALUE` (header — skip)
- Data lines: `YYYY-MM-DD,<rate>`
  - `TIME_PERIOD` = date in `YYYY-MM-DD` format
  - `OBS_VALUE` = exchange rate (INR per 1 EUR)

## Example Request

```bash
curl "https://data-api.ecb.europa.eu/service/data/EXR/D.INR.EUR.SP00.A?startPeriod=2024-01-02&endPeriod=2024-01-02&format=csvdata"
```

### Example Response

```
KEY,0
TIME_PERIOD,OBS_VALUE
2024-01-02,91.8295
```

## Conversion Formula

To convert NAV from INR to EUR:

```
NAV_EUR = NAV_INR / OBS_VALUE
```

Where:
- `NAV_INR` = Net Asset Value in Indian Rupees
- `OBS_VALUE` = exchange rate (INR per 1 EUR)
- `NAV_EUR` = equivalent value in Euros

**Example:**
- NAV_INR = 39.9169
- OBS_VALUE = 91.8295
- NAV_EUR = 39.9169 / 91.8295 = 0.4347...

## Market Hours & Holidays

- ECB data is typically available only for **working days** (Monday–Friday)
- No data for weekends or ECB holidays
- **Christmas (25 Dec), New Year (1 Jan, 2 Jan)**, etc. may have no rates

## Fallback Strategy

If no rate exists for the exact requested date:
1. Search **backwards** from the requested date
2. Find the **nearest prior working day** with an available rate
3. Use that rate (typically 1 day before for weekends)

**Do not search forward** — always use the most recent available rate.

**Example:**
- Requested date: 2024-01-06 (Saturday) → no ECB data
- Search backwards: 2024-01-05 (Friday) → has rate 91.8295 ✓

## Caching Strategy

### Exchange Rate Cache

Maintain an **in-memory map** to cache rates:

```go
var rateCache = make(map[string]float64) // date (YYYY-MM-DD) -> rate
```

**Cache behavior:**
1. Check cache for exact date
2. If not in cache, check the database (or API)
3. On API response, cache all dates returned
4. **Cache duration:** entire application lifetime (within-process)

**Example:**
```go
rateCache["2024-01-02"] = 91.8295
rateCache["2024-01-03"] = 91.7890
```

## Error Handling

| Scenario | Action |
|----------|--------|
| HTTP 400 (bad date) | Return `ErrInvalidDateFormat` |
| HTTP 404 | Return `ErrRateNotFound` |
| HTTP 5xx | Retry once after 5 seconds |
| Network timeout (>15s) | Return `ErrHTTPTimeout` |
| Invalid CSV or no data | Return `ErrInvalidResponse` |

## HTTP Timeout

Set a **15-second timeout** for all ECB requests.

## Example Workflow

1. **Fetch rate for single date:**
   ```bash
   curl "https://data-api.ecb.europa.eu/service/data/EXR/D.INR.EUR.SP00.A?startPeriod=2024-01-02&endPeriod=2024-01-02&format=csvdata"
   ```
   Response:
   ```
   KEY,0
   TIME_PERIOD,OBS_VALUE
   2024-01-02,91.8295
   ```

2. **Parse CSV:**
   - Split by newline
   - Skip first two lines (KEY and header)
   - For each data line: split by comma
   - Extract date and rate

3. **Cache the rate:**
   ```go
   rateCache["2024-01-02"] = 91.8295
   ```

4. **Convert NAV:**
   ```go
   navEUR = 39.9169 / 91.8295
   ```

## Batch Fetching

For multiple dates, you can request a range:

```bash
curl "https://data-api.ecb.europa.eu/service/data/EXR/D.INR.EUR.SP00.A?startPeriod=2024-01-02&endPeriod=2024-12-31&format=csvdata"
```

This is more efficient than individual requests. Cache all returned rates.

## Rate Availability

- Rates are typically available from **1999-01-01 onwards**
- Daily rates (Monday–Friday only, with rare exceptions for holidays)
- Historical rates do not change; caching is safe

## Precision

- Exchange rates are typically provided to **4 decimal places** (e.g., 91.8295)
- When converting NAV: use at least **4 decimal places** in the output
