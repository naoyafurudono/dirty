package example

// GetUserByID retrieves a user from the database by ID
// dirty: { select[user] }
func GetUserByID(_ int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// WriteLog writes a log message
// dirty: { insert[log] }
func WriteLog(_ string) error {
	// INSERT INTO logs (message) VALUES (?)
	return nil
}

// ProcessUser demonstrates correct effect declaration
// dirty: { select[user] | insert[log] }
func ProcessUser(id int64) error {
	if err := GetUserByID(id); err != nil {
		return err
	}
	return WriteLog("user processed")
}

// ProcessUserBroken demonstrates missing effect error
// dirty: { insert[log] }
func ProcessUserBroken(id int64) error {
	if err := GetUserByID(id); err != nil {
		return err
	}
	return WriteLog("user processed")
}

// HelperFunction is an undeclared function with implicit effects
func HelperFunction(id int64) error {
	return GetUserByID(id)
}

// UseHelper demonstrates missing implicit effects error
// dirty: { insert[log] }
func UseHelper(id int64) error {
	if err := HelperFunction(id); err != nil {
		return err
	}
	return WriteLog("done")
}
