.PHONY: build test test-cover lint lint-fix fmt clean run install deps-update outdated audit

APP_NAME := dotset
BUILD_DIR := bin

# Go is expected on PATH; override with `make GO=/path/to/go` if needed.
GO ?= go

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOFLAGS := -ldflags="-s -w -X main.Version=$(VERSION)"

build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)/

test:
	$(GO) test -race -count=1 ./...

test-cover:
	$(GO) test -race -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

run: build
	./$(BUILD_DIR)/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR)/ coverage.out coverage.html

install:
	$(GO) install ./cmd/$(APP_NAME)/

deps-update:
	$(GO) get -u=patch ./...
	$(GO) mod tidy

outdated:
	$(GO) list -u -m -json all | go-mod-outdated -update -direct

audit:
	govulncheck ./...