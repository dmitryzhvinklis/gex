# Gex Shell Makefile
# High-performance Linux shell written in Go

.PHONY: build install clean test benchmark run help

# Variables
BINARY_NAME=gex
VERSION=1.0.0
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
GO_FLAGS=-ldflags="-s -w -X main.VERSION=$(VERSION)"
BENCHMARK_FLAGS=-benchmem -benchtime=5s

# Default target
all: build

# Build the binary
build:
	@echo "Building Gex v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install to system
install: build
	@echo "Installing Gex to $(INSTALL_DIR)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall from system
uninstall:
	@echo "Uninstalling Gex..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Performance profiling
profile: build
	@echo "Running with CPU profiling..."
	@$(BUILD_DIR)/$(BINARY_NAME) -cpuprofile=cpu.prof
	@go tool pprof cpu.prof

# Memory profiling
memprofile: build
	@echo "Running with memory profiling..."
	@$(BUILD_DIR)/$(BINARY_NAME) -memprofile=mem.prof
	@go tool pprof mem.prof

# Static analysis
lint:
	@echo "Running static analysis..."
	@go vet ./...
	@go fmt ./...
	@golangci-lint run 2>/dev/null || echo "golangci-lint not available"

# Security scan
security:
	@echo "Running security scan..."
	@gosec ./... 2>/dev/null || echo "gosec not available"

# Generate documentation
docs:
	@echo "Generating documentation..."
	@go doc -all . > docs/api.txt
	@echo "Documentation generated in docs/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f *.prof
	@rm -f *.test
	@echo "Clean complete"

# Create distribution package
dist: release
	@echo "Creating distribution package..."
	@mkdir -p dist
	@tar -czf dist/gex-$(VERSION)-linux-amd64.tar.gz \
		-C $(BUILD_DIR) $(BINARY_NAME) \
		-C .. install.sh README.md FEATURES.md
	@echo "Distribution package created: dist/gex-$(VERSION)-linux-amd64.tar.gz"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go mod tidy
	@go mod download
	@echo "Development setup complete"

# Quick install using installer script
quick-install:
	@echo "Running quick install..."
	@GEX_LOCAL_PATH=$(PWD) ./install.sh

# Check dependencies
deps:
	@echo "Checking dependencies..."
	@go mod verify
	@go list -m all

# Show help
help:
	@echo "Gex Shell Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  release      - Build optimized release version"
	@echo "  install      - Install to system (requires sudo)"
	@echo "  uninstall    - Remove from system (requires sudo)"
	@echo "  run          - Build and run the shell"
	@echo "  test         - Run tests"
	@echo "  benchmark    - Run performance benchmarks"
	@echo "  profile      - Run with CPU profiling"
	@echo "  memprofile   - Run with memory profiling"
	@echo "  lint         - Run static analysis"
	@echo "  security     - Run security scan"
	@echo "  docs         - Generate documentation"
	@echo "  clean        - Clean build artifacts"
	@echo "  dist         - Create distribution package"
	@echo "  dev-setup    - Setup development environment"
	@echo "  quick-install- Install using installer script"
	@echo "  deps         - Check dependencies"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION      - Version number (default: $(VERSION))"
	@echo "  INSTALL_DIR  - Installation directory (default: $(INSTALL_DIR))"
