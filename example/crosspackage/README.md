# Cross-Package Analysis Example

This example demonstrates how dirty can analyze effects across package boundaries.

## Structure

```
.
├── db/
│   └── queries.go      # Database queries with effects
├── service/
│   └── user.go         # Service layer using db package
├── handler/
│   └── api.go          # HTTP handlers using service package
└── effect-registry.json # Effect definitions for external packages
```

## Running the Analysis

```bash
# From the example/crosspackage directory
go run ../../cmd/dirty/main.go ./...
```

## Expected Behavior

The analyzer should:

1. Track effects from `db` package functions
2. Propagate those effects through `service` package 
3. Ensure `handler` package declares all transitive effects
4. Report errors when effects are missing