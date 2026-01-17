.PHONY: build run test clean dev web-dev web-build screenshots playwright-install

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go binary
BINARY = zeitpolizei

# Build for current platform
build:
	@echo "Building $(BINARY)..."
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/zeitpolizei

# Build for all platforms
build-all: build-linux build-linux-arm build-darwin build-darwin-arm

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/zeitpolizei

build-linux-arm:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64 ./cmd/zeitpolizei

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 ./cmd/zeitpolizei

build-darwin-arm:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/zeitpolizei

# Build for UDM (ARM64 Linux with static linking)
build-udm:
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 \
		go build $(LDFLAGS) -o bin/$(BINARY)-udm ./cmd/zeitpolizei

# Run the application
run:
	go run ./cmd/zeitpolizei -config config.yaml

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air -c .air.toml

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build web UI
web-build:
	cd web && npm install && npm run build

# Run web UI in dev mode
web-dev:
	cd web && npm run dev

# Build with embedded web UI
build-with-web: web-build build

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Docker build
docker-build:
	docker build -t zeitpolizei:$(VERSION) -f deploy/docker/Dockerfile .

# Docker run
docker-run:
	docker run -p 8765:8765 -v $(PWD)/config.yaml:/app/config.yaml zeitpolizei:$(VERSION)

# Generate screenshots for documentation
screenshots:
	cd web && npm install && npx playwright install chromium && npm run screenshots

# Install Playwright browsers
playwright-install:
	cd web && npx playwright install chromium
