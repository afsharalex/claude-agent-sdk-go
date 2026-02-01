# Claude Agent SDK Go - Development Automation
# Usage: just <recipe>

# Default recipe - show available recipes
default:
    @just --list

# Run tests with race detection and coverage
test:
    go test -race -cover ./...

# Run tests with verbose output
test-verbose:
    go test -race -cover -v ./...

# Run tests for a specific package
test-pkg pkg:
    go test -race -cover -v ./{{pkg}}/...

# Generate HTML coverage report
coverage:
    go test -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Generate coverage report and open in browser
coverage-open: coverage
    open coverage.html || xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html manually"

# Verify coverage meets threshold (default 75%)
# Note: This checks the main package coverage, excluding examples and internal packages
# that require CLI integration testing.
coverage-check threshold="75":
    #!/usr/bin/env bash
    set -euo pipefail
    go test -race -coverprofile=coverage.out github.com/anthropics/claude-agent-sdk-go
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    threshold={{threshold}}
    echo "Main package coverage: ${coverage}%"
    echo "Threshold: ${threshold}%"
    if (( $(echo "$coverage < $threshold" | bc -l) )); then
        echo "ERROR: Coverage ${coverage}% is below threshold ${threshold}%"
        exit 1
    fi
    echo "OK: Coverage meets threshold"

# Verify coverage meets threshold across all packages (more strict)
coverage-check-all threshold="45":
    #!/usr/bin/env bash
    set -euo pipefail
    go test -race -coverprofile=coverage.out ./...
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    threshold={{threshold}}
    echo "Total coverage (all packages): ${coverage}%"
    echo "Threshold: ${threshold}%"
    if (( $(echo "$coverage < $threshold" | bc -l) )); then
        echo "ERROR: Coverage ${coverage}% is below threshold ${threshold}%"
        exit 1
    fi
    echo "OK: Coverage meets threshold"

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run staticcheck
staticcheck:
    staticcheck ./...

# Run go vet
vet:
    go vet ./...

# Run all linting checks
check: vet lint staticcheck

# Run goimports and gofmt
format:
    goimports -w .
    gofmt -s -w .

# Check formatting (for CI - fails if changes needed)
format-check:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -n "$(goimports -l .)" ]; then
        echo "The following files need goimports:"
        goimports -l .
        exit 1
    fi
    if [ -n "$(gofmt -l .)" ]; then
        echo "The following files need gofmt:"
        gofmt -l .
        exit 1
    fi
    echo "Formatting check passed"

# Build the SDK
build:
    go build ./...

# Build example programs
build-examples:
    @echo "Building examples..."
    @for dir in examples/*/; do \
        if [ -f "${dir}main.go" ]; then \
            echo "Building ${dir}..."; \
            go build -o "${dir}$(basename ${dir})" "${dir}"; \
        fi \
    done
    @echo "Examples built successfully"

# Run an example (default: simple)
run example="simple":
    @if [ -d "examples/{{example}}" ]; then \
        go run ./examples/{{example}}; \
    else \
        echo "Example '{{example}}' not found. Available examples:"; \
        ls -d examples/*/ 2>/dev/null | xargs -n1 basename || echo "No examples found"; \
        exit 1; \
    fi

# Hot reload development with air
dev:
    @if command -v air >/dev/null 2>&1; then \
        air; \
    else \
        echo "air not installed. Run 'just install-tools' first."; \
        exit 1; \
    fi

# Install development tools
install-tools:
    @echo "Installing development tools..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install honnef.co/go/tools/cmd/staticcheck@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/air-verse/air@latest
    @echo "All tools installed successfully"

# Check if required tools are installed
check-tools:
    @echo "Checking for required tools..."
    @command -v golangci-lint >/dev/null 2>&1 && echo "✓ golangci-lint" || echo "✗ golangci-lint (run 'just install-tools')"
    @command -v staticcheck >/dev/null 2>&1 && echo "✓ staticcheck" || echo "✗ staticcheck (run 'just install-tools')"
    @command -v goimports >/dev/null 2>&1 && echo "✓ goimports" || echo "✗ goimports (run 'just install-tools')"
    @command -v air >/dev/null 2>&1 && echo "✓ air" || echo "✗ air (run 'just install-tools')"

# Clean build artifacts
clean:
    rm -f coverage.out coverage.html
    rm -f examples/*/$(ls examples/ 2>/dev/null | head -1) 2>/dev/null || true
    go clean ./...
    @echo "Cleaned build artifacts"

# Full CI pipeline
ci: format-check vet lint staticcheck test coverage-check

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Run benchmarks for a specific package
bench-pkg pkg:
    go test -bench=. -benchmem ./{{pkg}}/...

# Update dependencies
deps-update:
    go get -u ./...
    go mod tidy

# Verify dependencies
deps-verify:
    go mod verify

# Run tests with short flag (skip long-running tests)
test-short:
    go test -short -race -cover ./...

# Generate mocks (if using mockgen)
mocks:
    @echo "Generating mocks..."
    @if command -v mockgen >/dev/null 2>&1; then \
        go generate ./...; \
    else \
        echo "mockgen not installed. Install with: go install github.com/golang/mock/mockgen@latest"; \
    fi

# Show test coverage by function
coverage-func:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run tests and show only failed tests
test-fail:
    go test -race ./... 2>&1 | grep -E "(FAIL|---)" || echo "All tests passed!"

# Tidy go modules
tidy:
    go mod tidy

# Download dependencies
deps:
    go mod download

# Print Go environment info
env:
    go env
