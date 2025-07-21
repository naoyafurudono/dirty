// Package example demonstrates complex effect chain patterns
package example

// LoadConfig loads configuration from storage
// dirty: { select[config] }
func LoadConfig() error {
	return nil
}

// GetUserData retrieves user data from the database
// dirty: { select[user] }
func GetUserData(_ int64) error {
	return nil
}

// WriteAuditLog writes an entry to the audit log
// dirty: { insert[audit] }
func WriteAuditLog(_ string) error {
	return nil
}

// LoadUserWithConfig loads config then retrieves user data
// dirty: { select[config] | select[user] }
func LoadUserWithConfig(id int64) error {
	if err := LoadConfig(); err != nil {
		return err
	}
	return GetUserData(id)
}

// GetUserWithAudit retrieves user and writes audit log
// dirty: { select[user] | insert[audit] }
func GetUserWithAudit(id int64) error {
	if err := GetUserData(id); err != nil {
		return err
	}
	return WriteAuditLog("user accessed")
}

// ComplexOperation combines all operations
// dirty: { select[config] | select[user] | insert[audit] }
func ComplexOperation(id int64) error {
	if err := LoadUserWithConfig(id); err != nil {
		return err
	}
	return GetUserWithAudit(id)
}

// BrokenDeepChain demonstrates missing effects in deep call chains
// dirty: { insert[audit] }
func BrokenDeepChain(id int64) error {
	// ComplexOperationは以下のエフェクトを持つ:
	// - select[config] (LoadConfig経由)
	// - select[user] (GetUserData経由、2回)
	// - insert[audit] (WriteAuditLog経由)
	return ComplexOperation(id)
}
