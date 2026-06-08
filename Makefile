.PHONY: build clean test run version

# Build variables
BINARY_NAME=mimo
BUILD_DIR=./bin
VERSION=0.1.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date +%Y-%m-%d)

# Go path for Windows
GO=go

# Build the binary
build:
	$(GO) build -ldflags "-X github.com/mimo-cli/mimo-cli/internal/version.Version=$(VERSION) \
		-X github.com/mimo-cli/mimo-cli/internal/version.GitCommit=$(GIT_COMMIT) \
		-X github.com/mimo-cli/mimo-cli/internal/version.BuildDate=$(BUILD_DATE)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GO) test ./... -v

# Run the binary
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Install dependencies
deps:
	$(GO) mod tidy

# Show version
version: build
	$(BUILD_DIR)/$(BINARY_NAME) version
