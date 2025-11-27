#!/bin/bash

set -e

echo "Building ImageD..."

# Create binaries directory
mkdir -p bin

# Build main CLI
echo "Building CLI..."
go build -o bin/imaged ./cmd/imaged-cli

# Build examples
echo "Building examples..."
go build -o bin/examples/basic_scan ./examples/basic_scan
go build -o bin/examples/quality_check ./examples/quality_check

# Run tests
echo "Running tests..."
go test ./...

echo "Build completed successfully!"
echo "Binary location: ./bin/imaged"