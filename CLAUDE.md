# IntelMesh Go SDK — Development Guidelines

## Project Overview

Public Go client library for the IntelMesh Risk Intelligence Engine API. This SDK is the official Go integration — it defines its own types (does NOT import from the engine codebase) and communicates exclusively via HTTP.

## Code Conventions

- Go 1.22+
- All exported types, functions, and methods need JSDoc-style comments
- Short interfaces (2-4 methods max)
- Error handling: typed errors with sentinel values
- `go get` for dependencies, NEVER edit go.mod manually
- Tests with `-race` flag
- Coverage target: 80%+

## Architecture

- `client.go` — Main IntelMesh client struct
- `resources/` — One file per API resource (events, rules, phases, etc.)
- `errors.go` — Typed error hierarchy
- `types.go` — All API request/response types (standalone, no engine imports)
- `builders/` — Fluent builders (EventBuilder, RuleBuilder)
- `pagination.go` — Cursor-based async pagination

## Testing

- Use `net/http/httptest` for unit tests
- Integration tests use build tag `//go:build integration`
- Pre-commit: golangci-lint + gosec + go test

## Rules

- NEVER import from the IntelMesh engine (`github.com/intelmesh/intelmesh`)
- All types are self-contained — they mirror the JSON API contract
- Errors map 1:1 to HTTP status codes
- Context propagation on all methods
