# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

PLUGIN_NAME := vault-plugin-database-db2
PLUGIN_DIR := cmd/$(PLUGIN_NAME)
BINARY_NAME := $(PLUGIN_NAME)

# Build settings
GO := go
GOFMT := gofmt
GOVET := $(GO) vet
GOTEST := $(GO) test
GOBUILD := $(GO) build

# Output directory
BIN_DIR := bin

.PHONY: all build clean test fmt vet dev

all: fmt vet test build

# Build the plugin binary
build:
	@echo "==> Building $(PLUGIN_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) ./$(PLUGIN_DIR)

# Build for development (with race detector)
dev:
	@echo "==> Building $(PLUGIN_NAME) for development..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -race -o $(BIN_DIR)/$(BINARY_NAME) ./$(PLUGIN_DIR)

# Run tests
test:
	@echo "==> Running tests..."
	$(GOTEST) -v ./...

# Format code
fmt:
	@echo "==> Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
vet:
	@echo "==> Running go vet..."
	$(GOVET) ./...

# Clean build artifacts
clean:
	@echo "==> Cleaning..."
	@rm -rf $(BIN_DIR)

# Install dependencies
deps:
	@echo "==> Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Cross-compile for multiple platforms
build-all:
	@echo "==> Building for multiple platforms..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 ./$(PLUGIN_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 ./$(PLUGIN_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(PLUGIN_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(PLUGIN_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(PLUGIN_DIR)
