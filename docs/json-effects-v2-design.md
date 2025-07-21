# JSON Effects V2 Design: Package and Type-Aware Effects

## Overview

Extend JSON effect definitions to support package-aware function effects and type-aware method effects, using exact matching only.

## Design Principles

1. **No Backward Compatibility** - Breaking change is acceptable
2. **Exact Match Only** - No wildcards, no patterns, no fallbacks  
3. **Explicit is Better** - All effects must be fully qualified
4. **Simple Implementation** - Minimize complexity

## JSON Schema

```json
{
  "version": "2.0",
  "effects": [
    {
      "package": "github.com/example/services",
      "function": "CreateUser",
      "effects": "{ insert[users] | event[user.created] }"
    },
    {
      "package": "github.com/example/services",
      "receiver": "*UserService",
      "method": "GetUser",
      "effects": "{ select[users] }"
    },
    {
      "package": "github.com/example/models",
      "receiver": "User",
      "method": "Validate",
      "effects": "{ }"
    }
  ]
}
```

## Effect Entry Types

### 1. Package Functions
For top-level functions in a package:
```json
{
  "package": "github.com/example/utils",
  "function": "HashPassword",
  "effects": "{ cpu_intensive }"
}
```

### 2. Methods with Pointer Receivers
For methods with pointer receivers:
```json
{
  "package": "github.com/example/services",
  "receiver": "*UserService",
  "method": "UpdateUser",
  "effects": "{ update[users] | event[user.updated] }"
}
```

### 3. Methods with Value Receivers
For methods with value receivers:
```json
{
  "package": "github.com/example/models",
  "receiver": "User",
  "method": "String",
  "effects": "{ }"
}
```

## Matching Rules

### Function Matching
- Match requires: `package` + `function`
- No match = no effects from JSON

### Method Matching  
- Match requires: `package` + `receiver` + `method`
- Receiver must match exactly (including pointer/value distinction)
- No match = no effects from JSON

## Data Structure

```go
// EffectKey uniquely identifies a function or method
type EffectKey struct {
    Package  string // Full package path
    Receiver string // Empty for functions, type name for methods
    Name     string // Function or method name
}

// EffectRegistry holds all JSON-defined effects
type EffectRegistry struct {
    effects map[EffectKey]*ast.EffectExpr
}

// Lookup returns effects for exact match only
func (r *EffectRegistry) Lookup(pkg, receiver, name string) *ast.EffectExpr {
    key := EffectKey{
        Package:  pkg,
        Receiver: receiver,
        Name:     name,
    }
    return r.effects[key]
}
```

## Implementation Details

### 1. Identifying Functions and Methods

During analysis, we need to extract:
- Package path from the current compilation unit
- Receiver type (if method) with pointer/value distinction
- Function/method name

```go
// Example extraction logic
func extractCallInfo(call *ast.CallExpr, pkg *types.Package) (pkgPath, receiver, name string) {
    switch fun := call.Fun.(type) {
    case *ast.Ident:
        // Package function in same package
        return pkg.Path(), "", fun.Name
        
    case *ast.SelectorExpr:
        // Could be method call or qualified function
        if recv := getReceiverType(fun.X); recv != "" {
            // Method call
            return pkg.Path(), recv, fun.Sel.Name
        } else {
            // Qualified function from another package
            if pkgName := getPackageName(fun.X); pkgName != "" {
                if importedPkg := findImportedPackage(pkgName); importedPkg != nil {
                    return importedPkg.Path(), "", fun.Sel.Name
                }
            }
        }
    }
    return "", "", ""
}
```

### 2. Receiver Type Extraction

```go
func getReceiverType(expr ast.Expr) string {
    // Use type checker to get the type
    if t := typeOf(expr); t != nil {
        switch t := t.(type) {
        case *types.Named:
            return t.Obj().Name()
        case *types.Pointer:
            if named, ok := t.Elem().(*types.Named); ok {
                return "*" + named.Obj().Name()
            }
        }
    }
    return ""
}
```

### 3. Cross-Package Analysis

For calls to functions/methods in other packages:
1. Resolve the import path from the import statement
2. Use the fully qualified package path for lookup
3. Handle vendored packages correctly

## Migration from V1

Users must update their JSON files:

### Before (V1):
```json
{
  "functions": [
    {
      "name": "Process",
      "effects": "{ file_write }"
    }
  ]
}
```

### After (V2):
```json
{
  "version": "2.0",
  "effects": [
    {
      "package": "github.com/example/processor",
      "function": "Process", 
      "effects": "{ file_write }"
    }
  ]
}
```

## Example: Complete JSON File

```json
{
  "version": "2.0",
  "effects": [
    {
      "package": "github.com/myapp/handlers",
      "function": "HandleLogin",
      "effects": "{ select[users] | insert[sessions] | event[user.login] }"
    },
    {
      "package": "github.com/myapp/services",
      "receiver": "*AuthService",
      "method": "ValidateToken",
      "effects": "{ select[sessions] }"
    },
    {
      "package": "github.com/myapp/models",
      "receiver": "User",
      "method": "FullName",
      "effects": "{ }"
    },
    {
      "package": "github.com/jmoiron/sqlx",
      "receiver": "*DB",
      "method": "Select",
      "effects": "{ select[*] }"
    }
  ]
}
```

## Error Handling

1. **Invalid JSON**: Fatal error, analysis fails
2. **Missing Required Fields**: Entry is skipped with warning
3. **Duplicate Entries**: Last entry wins
4. **Version Mismatch**: Fatal error if version != "2.0"

## Testing Strategy

1. Unit tests for effect registry and lookup
2. Integration tests with real Go packages
3. Tests for cross-package calls
4. Tests for method vs function distinction
5. Tests for pointer vs value receivers

## Future Considerations

While keeping the current design simple, these could be added later:
1. Effect parameters (e.g., `select[users:id,name]`)
2. Conditional effects based on arguments
3. Effect composition/inheritance
4. Support for interface methods

## Implementation Plan

1. **Phase 1: Core Implementation**
   - New JSON parser for V2 format
   - EffectRegistry with exact matching
   - Integration with analyzer

2. **Phase 2: Type System Integration**
   - Accurate receiver type extraction
   - Cross-package import resolution
   - Method vs function disambiguation

3. **Phase 3: Migration and Testing**
   - Migration guide and examples
   - Comprehensive test suite
   - Performance optimization