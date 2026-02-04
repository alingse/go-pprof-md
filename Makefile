.PHONY: build test clean install

BINARY_NAME=go-pprof-md
BUILD_DIR=bin
GO=go

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

test:
	@echo "Running tests..."
	$(GO) test -v ./...

clean:
	@echo "Cleaning..."
	$(GO) clean
	rm -rf $(BUILD_DIR)

install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install ./cmd/$(BINARY_NAME)

lint:
	@echo "Running linters..."
	$(GO) vet ./...

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

mod-tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy

all: fmt lint test build
