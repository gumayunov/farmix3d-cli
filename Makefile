.PHONY: build clean test test-verbose test-coverage run slice volume fmt vet build-all build-linux build-macos build-windows dist

BINARY_NAME=farmix-cli
BUILD_DIR=build

build:
	@echo "Building $(BINARY_NAME) for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Cross-compilation targets
build-all: build-linux build-macos build-windows
	@echo "Built for all platforms"

build-linux:
	@echo "Building $(BINARY_NAME) for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

build-macos:
	@echo "Building $(BINARY_NAME) for macOS (amd64 and arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 .
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 .

build-windows:
	@echo "Building $(BINARY_NAME) for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Create distribution archives
dist: build-all
	@echo "Creating distribution archives..."
	@mkdir -p $(BUILD_DIR)/dist
	@tar -czf $(BUILD_DIR)/dist/$(BINARY_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	@tar -czf $(BUILD_DIR)/dist/$(BINARY_NAME)-macos-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-macos-amd64
	@tar -czf $(BUILD_DIR)/dist/$(BINARY_NAME)-macos-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-macos-arm64
	@zip -j $(BUILD_DIR)/dist/$(BINARY_NAME)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "Distribution archives created in $(BUILD_DIR)/dist/"

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@go clean

test:
	@echo "Running tests..."
	@go test ./...

test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...

run:
	@go run . $(ARGS)

slice:
	@echo "Running slice command..."
	@go run . slice $(ARGS)

volume:
	@echo "Running volume command..."
	@go run . volume $(ARGS)

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...


deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Cross-compilation help
help-cross:
	@echo "Cross-compilation commands:"
	@echo "  make build-all       - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make build-linux     - Build for Linux amd64"
	@echo "  make build-macos     - Build for macOS (both amd64 and arm64)"
	@echo "  make build-windows   - Build for Windows amd64"
	@echo "  make dist            - Build all and create distribution archives"
	@echo ""
	@echo "Manual cross-compilation examples:"
	@echo "  GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux ."
	@echo "  GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-macos ."
	@echo "  GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-macos-arm64 ."
	@echo "  GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe ."