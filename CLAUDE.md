# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

"dirty" is an effect system for Go implemented as a static analysis tool. It tracks function side effects through `// dirty:` comment annotations.

## Key Concepts

```go
// dirty: { select[user] }
func GetUser(id int64) (User, error) { ... }

// dirty: { select[user] | select[member] }
func GetUserWithMember(userID int64) (User, []Member, error) { ... }
```

**Syntax:**

- Effects are declared within `{ }` braces
- Multiple effects are separated by `|` (pipe)
- Empty set is `{ }`
- Effect labels are treated as opaque tokens (e.g., `select[user]`)

**Rules:**

- Functions must declare ALL effects (including those from called functions)
- The new syntax is based on set notation for future extensibility

## Implementation Status

**Completed:** ✅
- Basic effect checking at call sites
- Implicit effect calculation for undeclared functions
- Method call support
- Circular dependency handling
- Error reporting with detailed diagnostics (DIRTY_VERBOSE=1)
- sqlc-use integration for automatic database effect detection
- Extensible effect syntax with set notation (Issue #2)

**Next:** Performance optimization, cross-package analysis

## Project Structure

```
analyzer/       # Core analysis logic
cmd/dirty/      # CLI entry point
testdata/       # Test cases
DESIGN.md       # Architecture details
```

## Testing

Run tests with: `make test` or `go test ./...`

Keep code base to pass all tests.

## Release

Increment git tag version when new feature added.
