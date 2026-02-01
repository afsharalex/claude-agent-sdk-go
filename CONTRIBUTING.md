# Contributing

## Prerequisites

- Go 1.25 or later
- Claude Code CLI installed (`curl -fsSL https://claude.ai/install.sh | bash`)
- golangci-lint (optional, for linting)

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/afsharalex/claude-agent-sdk-go.git
   cd claude-agent-sdk-go
   ```

2. Verify your Go version:
   ```bash
   go version
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Run tests to verify setup:
   ```bash
   go test ./...
   ```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/protocol/

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Formatting and Linting

```bash
# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run

# Tidy dependencies
go mod tidy
```

## Adding a New Option

1. Add the field to the `Options` struct in `options.go`:
   ```go
   type Options struct {
       // ... existing fields
       MyNewOption string
   }
   ```

2. Create a `With*` function:
   ```go
   // WithMyNewOption sets the new option value.
   func WithMyNewOption(value string) Option {
       return func(o *Options) {
           o.MyNewOption = value
       }
   }
   ```

3. If the option needs to be passed to the CLI, update `toTransportOptions()` in `claude.go`.

4. Add the option to `internal/transport/options.go` if needed.

5. Update `README.md` to document the new option.

## Adding a New Example

1. Create a directory under `examples/`:
   ```bash
   mkdir examples/my-example
   ```

2. Create `main.go` with a package comment explaining the example:
   ```go
   // Example: Description of what this example demonstrates
   //
   // This example shows how to...
   //
   // Run with: go run main.go
   package main
   ```

3. Follow the pattern of existing examples for signal handling and cleanup.

4. Update `README.md` to list the new example.

5. Verify the example runs:
   ```bash
   cd examples/my-example
   go run main.go
   ```

## Adding a New Message or Content Type

1. Define the type in `types.go`:
   ```go
   type MyNewBlock struct {
       Field string `json:"field"`
   }

   func (MyNewBlock) contentBlock() {}
   ```

2. Add parsing logic in `messages.go`:
   ```go
   case "my_new_type":
       field, _ := block["field"].(string)
       return MyNewBlock{Field: field}, nil
   ```

3. Add tests for the new type.

## Code Style Guidelines

- Follow standard Go conventions (`go fmt`, `golint`)
- Use meaningful variable names; avoid single letters except for loops
- Document all exported types, functions, and methods
- Keep functions focused; prefer small, composable functions
- Handle errors explicitly; avoid ignoring returned errors
- Use `context.Context` for cancellation and timeouts
- Prefer channels for concurrent communication

### Naming Conventions

- Types: `PascalCase` (e.g., `AssistantMessage`)
- Functions: `PascalCase` for exported, `camelCase` for unexported
- Constants: `PascalCase` (e.g., `HookEventPreToolUse`)
- Files: `snake_case.go`

### Comments

- Start comments with the name of the thing being documented
- End sentences with periods
- Keep comments concise and focused on "why" rather than "what"

## Pull Request Process

1. Fork the repository and create a feature branch:
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make your changes, following the code style guidelines.

3. Add or update tests as needed.

4. Run the full test suite:
   ```bash
   go test ./...
   ```

5. Run the linter:
   ```bash
   golangci-lint run
   ```

6. Commit with a descriptive message:
   ```bash
   git commit -m "Add support for X feature"
   ```

7. Push and open a pull request.

8. Ensure CI checks pass.

9. Respond to review feedback.

## Reporting Issues

When reporting bugs:
- Include Go version (`go version`)
- Include Claude Code CLI version (`claude --version`)
- Provide a minimal reproduction case
- Include error messages and stack traces

For feature requests:
- Describe the use case
- Explain why existing functionality doesn't meet the need
- Consider how it aligns with the Python SDK (if applicable)

## Questions?

Open a GitHub issue for questions about contributing.
