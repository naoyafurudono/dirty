# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"dirty" is an effect system for Go that works as a static analysis tool. It tracks function side effects through comment-based annotations and ensures effect declarations are consistent across function calls.

## Effect System Design

Functions are annotated with effects using `//dirty:` comments:

```go
//dirty: select[user]
func GetUser(id int64) (User, error) { ... }

//dirty: select[user], select[member]
func GetUserWithMember(userID int64) (User, []Member, error) { ... }
```

**Key Rules:**

- A function must declare ALL effects it produces (directly or through called functions)
- Effects from called functions must be a subset of the calling function's declared effects
- Multiple effects are comma-separated
- Effect labels follow the pattern: `action[target]` (e.g., `select[user]`, `insert[member]`)

**Internal Implementation Note:**
While the syntax remains `action[target]`, internally the analyzer treats each effect label as a simple opaque token. The brackets and structure are preserved for readability and future extensibility, but the current implementation does not parse or interpret the action/target components separately.

## Development Setup

Since this project is in design phase, initial setup will involve:

```bash
# Initialize Go module
go mod init github.com/naoyafurudono/dirty

# Create the analyzer package structure
mkdir -p analyzer
mkdir -p cmd/dirty
```

## Implementation Architecture

When implementing:

1. **AST Parser**: Extract `//dirty:` annotations from Go source files
2. **Effect Checker**: Verify effect consistency across function calls
3. **CLI Tool**: Provide command-line interface for running checks
4. **Integration**: Consider implementing as a `go vet` compatible analyzer

## Current Limitations (from design)

- No cross-module effect checking
- No support for higher-order functions
- Effects must be explicitly declared (no inference)

## Testing Strategy

When implementing tests:

- Unit tests for effect parsing logic
- Integration tests with sample Go code containing various effect patterns
- Test cases for both valid and invalid effect declarations

## Important Notes

- This is a static analysis tool, not a runtime library
- Effects are purely for compile-time checking
- The tool should integrate smoothly with existing Go toolchains
