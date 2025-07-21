// Package example demonstrates complex effect chaining scenarios
package example

// LoadConfig loads configuration with config selection effects
// dirty: { select[config] }
func LoadConfig() error {
	return nil
}

// GetUserData retrieves user data with user selection effects
// dirty: { select[user] }
func GetUserData(_ int64) error {
	return nil
}

// WriteAuditLog writes audit logs with audit insertion effects
// dirty: { insert[audit] }
func WriteAuditLog(_ string) error {
	return nil
}

// LoadUserWithConfig loads config and retrieves user data (chain 1: load config and get user)
func LoadUserWithConfig(id int64) error {
	if err := LoadConfig(); err != nil {
		return err
	}
	return GetUserData(id)
}

// GetUserWithAudit gets user data and writes audit log (chain 2: get user and write audit)
func GetUserWithAudit(id int64) error {
	if err := GetUserData(id); err != nil {
		return err
	}
	return WriteAuditLog("user accessed")
}

// ComplexOperation combines all operations (chain 3: combine everything)
func ComplexOperation(id int64) error {
	if err := LoadUserWithConfig(id); err != nil {
		return err
	}
	return GetUserWithAudit(id)
}

// BrokenDeepChain demonstrates missing effects in deep call chains (error: missing effects in deep call chain)
// dirty: { insert[audit] }
func BrokenDeepChain(id int64) error {
	// ComplexOperationは以下のエフェクトを持つ:
	// - select[config] (LoadConfig経由)
	// - select[user] (GetUserData経由、2回)
	// - insert[audit] (WriteAuditLog経由)
	return ComplexOperation(id)
}
