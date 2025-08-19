.PHONY: build clean test run fmt vet

BINARY_NAME=3mfanalyzer
BUILD_DIR=build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@go clean

test:
	@echo "Running tests..."
	@go test ./...

run:
	@go run . $(ARGS)

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy