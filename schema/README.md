# Effect Registry Schema

This directory contains the JSON Schema definition for Effect Registry files.

## Schema File

- `effect-registry.schema.json` - The official JSON Schema for `effect-registry.json` files

## Usage

### Validating JSON Files

You can use any JSON Schema validator to validate your effect declaration files.

#### Using ajv-cli

```bash
# Install ajv-cli
npm install -g ajv-cli

# Validate a file
ajv validate -s schema/effect-registry.schema.json -d effect-registry.json
```

#### Using Python jsonschema

```python
import json
import jsonschema

# Load schema
with open('schema/effect-registry.schema.json') as f:
    schema = json.load(f)

# Load data
with open('effect-registry.json') as f:
    data = json.load(f)

# Validate
jsonschema.validate(data, schema)
```

### VS Code Integration

Add this to your VS Code settings to get IntelliSense and validation:

```json
{
  "json.schemas": [
    {
      "fileMatch": ["effect-registry.json", "**/effect-registry.json"],
      "url": "./schema/effect-registry.schema.json"
    }
  ]
}
```

Or add this to the top of your `effect-registry.json` file:

```json
{
  "$schema": "../schema/effect-registry.schema.json",
  "version": "1.0",
  "effects": {
    // Your effects here
  }
}
```

## Schema Details

### Structure

```typescript
{
  version: "1.0",  // Required, currently only "1.0" is supported
  effects: {       // Required, mapping of function names to effect expressions
    [functionName: string]: string  // Effect expression in set notation
  }
}
```

### Effect Expression Pattern

The effect expression must follow this pattern:
- Wrapped in curly braces: `{ ... }`
- Empty set: `{ }`
- Single effect: `{ operation[target] }`
- Multiple effects separated by `|`: `{ effect1[target1] | effect2[target2] }`

Where:
- `operation` and `target` can contain letters, numbers, underscore, hyphen, and dot
- Must start with a letter or underscore

### Valid Examples

```json
{
  "version": "1.0",
  "effects": {
    "EmptyEffects": "{ }",
    "SimpleSelect": "{ select[users] }",
    "MultipleEffects": "{ select[users] | insert[logs] }",
    "ComplexIdentifiers": "{ network[external-api] | io[file.system] }",
    "ManyEffects": "{ select[users] | update[balance] | insert[transactions] | network[payment-api] }"
  }
}
```

### Invalid Examples

```json
{
  "version": "2.0",  // ❌ Unsupported version
  "effects": {
    "BadSyntax1": "select[users]",  // ❌ Missing curly braces
    "BadSyntax2": "{ select users }",  // ❌ Missing brackets
    "BadSyntax3": "{ select[] }",  // ❌ Empty target
    "BadSyntax4": "{ [users] }",  // ❌ Missing operation
    "BadSeparator": "{ select[users], insert[logs] }"  // ❌ Wrong separator (comma instead of |)
  }
}
```

## Future Extensions

The schema is designed to be extensible. Future versions might support:
- Effect references: `{ $common_effects | custom[operation] }`
- Effect operators: `(A | B) & C`
- Effect parameters: `{ network[api, timeout=30s] }`

The `version` field ensures backward compatibility when introducing new features.
