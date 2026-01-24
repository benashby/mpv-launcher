# Makefile for mpv-launcher

# Binary name
BINARY_NAME=mpv-launcher

# Installation directory (XDG-compatible)
INSTALL_DIR=$(HOME)/.local/bin

# Build directory
BUILD_DIR=build

# Go build flags
GO_BUILD_FLAGS=-ldflags="-s -w"

# CGO is required for go-astiav
export CGO_ENABLED=1

.PHONY: all build install uninstall clean test

all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mpv-launcher
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install the binary to XDG-compatible location
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete!"
	@echo "Make sure $(INSTALL_DIR) is in your PATH"

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstallation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./cmd/... ./internal/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./cmd/... ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "Dependencies downloaded!"

# Tidy up dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Dependencies tidied!"
