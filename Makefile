.PHONY: build test run-example lint clean help

help:
	@echo "navcalc - Indian Mutual Fund NAV Calculator"
	@echo ""
	@echo "Available targets:"
	@echo "  build         Build the navcalc binary"
	@echo "  test          Run tests"
	@echo "  run-example   Run with example ISINs for 2024-01-02 and 2024-12-31"
	@echo "  lint          Run go vet"
	@echo "  clean         Remove built binaries"
	@echo "  install       Install the binary to GOPATH/bin"

build:
	@echo "Building navcalc..."
	@CGO_ENABLED=0 go build -o navcalc ./cmd/navcalc
	@echo "✓ Built: ./navcalc"

test:
	@echo "Running tests..."
	@CGO_ENABLED=0 go test ./...

run-example: build
	@echo "Running example: NAVs for 8 funds on 2024-01-02 and 2024-12-31"
	./navcalc \
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

run-example-csv: build
	@echo "Running example with CSV output..."
	./navcalc \
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
		--output csv > nav_output_2024.csv
	@echo "✓ Saved to: nav_output_2024.csv"

lint:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ No issues found"

clean:
	@echo "Cleaning..."
	@rm -f navcalc
	@echo "✓ Cleaned"

install: build
	@echo "Installing navcalc..."
	@go install ./cmd/navcalc
	@echo "✓ Installed to $(go env GOPATH)/bin/navcalc"
