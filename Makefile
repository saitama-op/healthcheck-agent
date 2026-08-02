# Variables
BINARY_NAME=healthcheck-agent
MAIN_PATH=cmd/healthcheck-agent/main.go

# Build flags to strip symbols and reduce binary size to < 5MB
BUILD_FLAGS=-ldflags="-s -w" -trimpath

.PHONY: all build clean tidy test release

# Default target
all: tidy build

# Build for the current OS and architecture
build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Clean up binaries and release directories
clean:
	rm -f $(BINARY_NAME)
	rm -rf release/

# Format and tidy dependencies
tidy:
	go mod tidy
	go fmt ./...

# Run tests (useful for when you add internal/checker_test.go later)
test:
	go test -v ./...

# Cross-compile for all target platforms
release: clean tidy
	mkdir -p release
	@echo "Building for Linux (AMD64)..."
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o release/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Building for Linux (ARM64)..."
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o release/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Building for macOS (AMD64)..."
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o release/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Building for macOS (ARM64 / Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o release/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Building for Windows (AMD64)..."
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o release/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Release builds completed in ./release/"
