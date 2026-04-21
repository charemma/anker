# Go Developer Rules

## Code Quality

- All `fmt.Fprintf`, `fmt.Fprintln`, `fmt.Fprint` calls to `io.Writer` must have return values handled
- For non-actionable writes (logging to stderr): use `_, _ =` prefix
- errcheck linter is active -- no unchecked errors

## Modern Go (1.21+)

- `slices.Contains` over manual loops
- `maps.Keys` over manual key collection
- `log/slog` over `log` for structured logging
- `errors.Join` for combining errors

## Style

- Accept interfaces, return structs
- Short variable names in small scopes
- Error values, not exceptions
- Table-driven tests for multiple input/output combos
- `t.TempDir()` for filesystem tests, never write to real paths

## Testing

- Tests live next to code (`foo_test.go` alongside `foo.go`)
- Use `IKNO_HOME` env var pointed at `t.TempDir()` to isolate state
- Run: `go test ./...` or `nix flake check`

## CLI

- Cobra for command structure
- Commands register via `init()` in their own file
- CLI args > config file > defaults
- Output formats: simple, detailed, json, markdown

## Build

- Quick: `go build -o bin/ikno .`
- Reproducible: `nix build`
- Checks: `nix flake check` (tests + lint + build)
