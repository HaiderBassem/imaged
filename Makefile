.PHONY: build test bench clean install

BINARY_NAME=imaged
BUILD_DIR=bin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/imaged-cli

install:
	@go install ./cmd/imaged-cli

test:
	@echo "Running tests..."
	@go test ./... -v

bench:
	@echo "Running benchmarks..."
	@go test ./... -bench=. -benchmem

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) coverage.txt

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  install   - Install the binary"
	@echo "  test      - Run tests"
	@echo "  bench     - Run benchmarks"
	@echo "  clean     - Clean build artifacts"