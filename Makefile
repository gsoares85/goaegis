.PHONY: help build test lint clean install

BINARY_NAME=goaegis
GO=go

help:
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build:
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

test:
	$(GO) test -v -race -coverprofile=coverage.txt ./...

test-coverage:
	$(GO) test -v -race -coverprofile=coverage.txt ./...
	$(GO) tool cover -html=coverage.txt -o coverage.html

lint:
	$(GO) vet ./...
	golangci-lint run ./...

fmt:
	gofmt -s -w .

clean:
	$(GO) clean
	rm -rf bin/ coverage.txt coverage.html

install:
	$(GO) mod download
	$(GO) mod tidy

check: fmt lint test

.DEFAULT_GOAL := help
