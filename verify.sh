#!/bin/bash
# verify.sh - Verify navcalc against live AMFI and ECB data

set -e

echo "==================================================================="
echo "AMFI NAV Verification Script"
echo "==================================================================="
echo ""
echo "This script fetches NAV data directly from AMFI and ECB to allow"
echo "manual verification of the tool's output against authoritative sources."
echo ""
echo "Expected ISIN Test List:"
echo "  INF846K01CH7  Axis Focused 25 Fund - Gr"
echo "  INF179KA1RZ8  HDFC Small Cap Fund - Gr"
echo "  INF767K01139  Motilal Oswal Multi-Asset Fund - Gr"
echo "  INF769K01EY2  SBI Emerging Businesses Fund - Gr"
echo "  INF247L01700  Mirae Asset Large Cap Fund - Gr"
echo "  INF247L01908  Mirae Asset Hybrid Balanced Fund - Gr"
echo "  INF247L01AH0  Mirae Asset Overnight Fund - Gr"
echo "  INF200K01537  ICICI Prudential Bluechip Fund - Gr"
echo ""
echo "==================================================================="
echo ""

echo "1. AMFI NAV Data for 02-Jan-2024"
echo "———————————————————————————————————————————————————————"
curl -s "https://portal.amfiindia.com/DownloadNAVHistoryReport_Po.aspx?tp=1&frmdt=02-Jan-2024&todt=02-Jan-2024" \
  | grep -E "INF846K01CH7|INF179KA1RZ8|INF767K01139|INF769K01EY2|INF247L01700|INF247L01908|INF247L01AH0|INF200K01537" | head -20

echo ""
echo ""
echo "2. AMFI NAV Data for 31-Dec-2024"
echo "———————————————————————————————————————————————————————"
curl -s "https://portal.amfiindia.com/DownloadNAVHistoryReport_Po.aspx?tp=1&frmdt=31-Dec-2024&todt=31-Dec-2024" \
  | grep -E "INF846K01CH7|INF179KA1RZ8|INF767K01139|INF769K01EY2|INF247L01700|INF247L01908|INF247L01AH0|INF200K01537" | head -20

echo ""
echo ""
echo "3. ECB EUR/INR Rate for 2024-01-02"
echo "———————————————————————————————————————————————————————"
curl -s "https://data-api.ecb.europa.eu/service/data/EXR/D.INR.EUR.SP00.A?startPeriod=2024-01-02&endPeriod=2024-01-02&format=csvdata" \
  | grep -v "^KEY\|^#" | head -10

echo ""
echo ""
echo "4. ECB EUR/INR Rate for 2024-12-31"
echo "———————————————————————————————————————————————————————"
curl -s "https://data-api.ecb.europa.eu/service/data/EXR/D.INR.EUR.SP00.A?startPeriod=2024-12-31&endPeriod=2024-12-31&format=csvdata" \
  | grep -v "^KEY\|^#" | head -10

echo ""
echo ""
echo "==================================================================="
echo "Verification complete!"
echo ""
echo "Use the above data to manually verify the navcalc output:"
echo ""
echo "  navcalc --date 2024-01-02 --date 2024-12-31 \\"
echo "    --isin INF846K01CH7 --isin INF179KA1RZ8 \\"
echo "    ... (other ISINs) ..."
echo "    --output table"
echo ""
echo "Or in CSV format:"
echo ""
echo "  navcalc --date 2024-01-02 --date 2024-12-31 \\"
echo "    ... (ISINs) ... \\"
echo "    --output csv"
echo ""
echo "==================================================================="
