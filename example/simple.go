package example

// GetUserByID performs database access to retrieve user data
// dirty: { select[user] }
func GetUserByID(_ int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// WriteLog records log messages
// dirty: { insert[log] }
func WriteLog(_ string) error {
	// INSERT INTO logs (message) VALUES (?)
	return nil
}

// ProcessUser correctly declares all required effects
// dirty: { select[user] | insert[log] }
func ProcessUser(id int64) error {
	if err := GetUserByID(id); err != nil {
		return err
	}
	return WriteLog("user processed")
}

// ProcessUserBroken demonstrates missing effects (error: missing select[user] effect)
// dirty: { insert[log] }
func ProcessUserBroken(id int64) error {
	if err := GetUserByID(id); err != nil {
		return err
	}
	return WriteLog("user processed")
}

// HelperFunction has no declarations but implicitly has effects (not checked but has implicit effects)
func HelperFunction(id int64) error {
	return GetUserByID(id)
}

// UseHelper demonstrates missing implicit effects (error: missing HelperFunction's implicit effects)
// dirty: { insert[log] }
func UseHelper(id int64) error {
	if err := HelperFunction(id); err != nil {
		return err
	}
	return WriteLog("done")
}
