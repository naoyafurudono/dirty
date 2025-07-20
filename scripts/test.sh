#!/bin/bash

set -e

echo "Running dirty effect analyzer tests..."
echo "======================================"

# Run unit tests
echo "Running unit tests..."
go test -v ./analyzer/...

# Run analyzer on test data
echo -e "\nRunning analyzer on test data..."
go run ./cmd/dirty ./testdata/src/basic
go run ./cmd/dirty ./testdata/src/complex || true

# Check test coverage
echo -e "\nGenerating test coverage..."
go test -coverprofile=coverage.out ./analyzer/...
go tool cover -func=coverage.out

echo -e "\nTest run complete!"