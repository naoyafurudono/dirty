# Package-Aware JSON Effect Definitions Design

## Problem Statement

Current JSON effect definitions only use function names as keys, causing conflicts when different packages have functions with the same name. All functions named `Process` across different packages receive the same effects.

## Design Goals

1. **Backward Compatibility**: Existing JSON definitions without package specifications must continue working
2. **Simplicity**: Keep the design simple and intuitive
3. **Flexibility**: Support various use cases (full paths, relative paths, wildcards)
4. **Performance**: Minimal impact on lookup performance

## Proposed Solution

### JSON Schema Extension

Add optional `package` field to function definitions:

```json
{
  "functions": [
    {
      "name": "Process",
      "package": "github.com/example/pkg/a",
      "effects": "{ file_write[a] }"
    },
    {
      "name": "Process", 
      "package": "github.com/example/pkg/b",
      "effects": "{ file_write[b] }"
    },
    {
      "name": "GlobalHelper",
      "effects": "{ log }"
    }
  ]
}
```

### Matching Rules (Priority Order)

1. **Exact Match**: Package path + function name
2. **Function-Only Match**: Function name without package (backward compatibility)
3. **No Match**: Function has no declared effects in JSON

### Package Path Formats

Support multiple formats for flexibility:

1. **Full Import Path**: `github.com/user/repo/pkg`
2. **Package Name Only**: `pkg` (matches any package ending with `/pkg`)
3. **Wildcard Patterns**: 
   - `github.com/user/repo/*` - all packages in repo
   - `*/models` - all packages named models
   - `github.com/user/*/pkg` - specific package across repos

### Method Support

For methods, use receiver type notation:

```json
{
  "name": "(*UserService).GetUser",
  "package": "github.com/example/services",
  "effects": "{ select[users] }"
}
```

## Implementation Plan

### Phase 1: Basic Package Support
1. Extend JSON parser to handle `package` field
2. Update `ParsedEffects` data structure
3. Implement exact match lookup
4. Add backward compatibility for package-less entries

### Phase 2: Pattern Matching
1. Implement wildcard support
2. Add package-name-only matching
3. Optimize lookup performance

### Phase 3: Advanced Features
1. Method receiver support
2. Import alias handling
3. Vendor directory support

## Data Structure Changes

Current:
```go
type ParsedEffects map[string]*ast.EffectExpr
```

Proposed:
```go
type ParsedEffects struct {
    // Exact matches: "pkg.path.Function" -> EffectExpr
    exact map[string]*ast.EffectExpr
    
    // Function-only matches: "Function" -> []EffectEntry
    functionOnly map[string][]*EffectEntry
    
    // Pattern matches: for wildcard support
    patterns []*EffectPattern
}

type EffectEntry struct {
    Package string
    Function string
    Effects *ast.EffectExpr
}

type EffectPattern struct {
    Pattern string  // e.g., "github.com/user/*"
    Regex   *regexp.Regexp
    Effects map[string]*ast.EffectExpr
}
```

## Lookup Algorithm

```go
func (pe *ParsedEffects) Lookup(pkg, function string) *ast.EffectExpr {
    // 1. Try exact match
    if effects, ok := pe.exact[pkg + "." + function]; ok {
        return effects
    }
    
    // 2. Try pattern matches
    for _, pattern := range pe.patterns {
        if pattern.Matches(pkg, function) {
            return pattern.Effects[function]
        }
    }
    
    // 3. Try function-only match (backward compatibility)
    if entries, ok := pe.functionOnly[function]; ok {
        // Return first match or implement precedence rules
        return entries[0].Effects
    }
    
    return nil
}
```

## Migration Strategy

1. Deploy with backward compatibility
2. Encourage package specifications for new entries
3. Provide tooling to auto-generate package paths
4. Eventually deprecate package-less entries (optional)

## Example Use Cases

### Case 1: Database Models
```json
{
  "functions": [
    {
      "name": "Save",
      "package": "*/models",
      "effects": "{ insert[*] | update[*] }"
    }
  ]
}
```

### Case 2: Service Layer
```json
{
  "functions": [
    {
      "name": "(*UserService).CreateUser",
      "package": "github.com/myapp/services",
      "effects": "{ insert[users] | event[user.created] }"
    }
  ]
}
```

### Case 3: Third-party Libraries
```json
{
  "functions": [
    {
      "name": "Query",
      "package": "github.com/jmoiron/sqlx",
      "effects": "{ select[*] }"
    }
  ]
}
```

## Testing Strategy

1. Unit tests for new lookup logic
2. Integration tests with real packages
3. Performance benchmarks
4. Backward compatibility tests
5. Edge cases (vendor, internal packages)

## Future Considerations

1. **Import Aliases**: Handle `import foo "github.com/bar"`
2. **Local Packages**: Better support for `./pkg` style imports
3. **Caching**: Optimize repeated lookups
4. **Precedence Rules**: When multiple patterns match
5. **Effect Inheritance**: Package-level default effects