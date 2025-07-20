.PHONY: all build test clean install

all: build test

build:
	go build -o bin/dirty ./cmd/dirty

test:
	go test -v ./...

install:
	go install ./cmd/dirty

clean:
	rm -rf bin/

# Run the analyzer on example code
check-examples:
	go run ./cmd/dirty ./testdata/...

# Run with race detector
test-race:
	go test -race -v ./...

# Generate test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run all checks
check: fmt test lint
