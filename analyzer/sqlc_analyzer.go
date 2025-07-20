package analyzer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
)

// SQLCOperation represents a single database operation from sqlc-use
type SQLCOperation struct {
	Operation string `json:"operation"`
	Table     string `json:"table"`
}

// SQLCQueryMap maps query names to their operations
type SQLCQueryMap map[string][]SQLCOperation

// LoadSQLCEffects loads effects from sqlc-use JSON file
func LoadSQLCEffects(jsonPath string) (SQLCQueryMap, error) {
	if jsonPath == "" {
		return nil, nil
	}

	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sqlc-use JSON: %w", err)
	}

	var queryMap SQLCQueryMap
	if err := json.Unmarshal(data, &queryMap); err != nil {
		return nil, fmt.Errorf("failed to parse sqlc-use JSON: %w", err)
	}

	return queryMap, nil
}

// ConvertToEffects converts SQL operations to dirty effect labels
func ConvertToEffects(operations []SQLCOperation) []string {
	effectSet := make(map[string]bool)

	for _, op := range operations {
		// Format: operation[table]
		effect := fmt.Sprintf("%s[%s]", op.Operation, op.Table)
		effectSet[effect] = true
	}

	// Convert to sorted slice for deterministic output
	effects := make([]string, 0, len(effectSet))
	for effect := range effectSet {
		effects = append(effects, effect)
	}
	sort.Strings(effects)

	return effects
}
