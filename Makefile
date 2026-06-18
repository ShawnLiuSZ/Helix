.PHONY: build test lint install release clean

APP_NAME   := helix
BUILD_DIR  := bin
CMD_DIR    := cmd/helix

GO         := go
GOFLAGS    := -ldflags="-s -w"

# Build
build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

# Test
test:
	$(GO) test ./... -count=1

test-cover:
	$(GO) test ./... -cover -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

test-cover-html:
	$(GO) test ./... -cover -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out

# Lint
lint:
	golangci-lint run ./...

# Install to local
install:
	$(GO) install ./$(CMD_DIR)

# Cross-compile
release:
	goreleaser build --snapshot --clean

# Clean
clean:
	rm -rf $(BUILD_DIR) dist/ coverage.out

# Dev setup
dev-setup:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
