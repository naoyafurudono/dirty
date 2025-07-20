package complex

// Test case: nested function calls with effect propagation

// dirty: { select[config] }
func LoadConfig() error {
	// SELECT * FROM config
	return nil
}

// dirty: { select[user] }
func GetUserByID(id int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// dirty: { select[user] | insert[audit_log] }
func GetUserWithAudit(id int64) error {
	err := GetUserByID(id)
	if err != nil {
		return err
	}
	// INSERT INTO audit_log ...
	return nil
}

// dirty: { select[config] | select[user] | insert[audit_log] }
func InitializeUserSession(userID int64) error {
	// Load configuration first
	if err := LoadConfig(); err != nil {
		return err
	}

	// Get user with audit logging
	if err := GetUserWithAudit(userID); err != nil {
		return err
	}

	return nil
}

// Invalid: missing config effect
// dirty: { select[user] | insert[audit_log] }
func InitializeUserSessionBroken(userID int64) error {
	if err := LoadConfig(); err != nil { // want "function calls LoadConfig which has effects \\[select\\[config\\]\\] not declared in this function"
		return err
	}

	if err := GetUserWithAudit(userID); err != nil {
		return err
	}

	return nil
}
