package implicit

// Test case: implicit effect propagation through functions without declarations

// Base functions with effects
// dirty: { select[user] }
func GetUser(id int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// dirty: { insert[log] }
func LogAction(action string) error {
	// INSERT INTO logs (action) VALUES (?)
	return nil
}

// Function without declaration - should have implicit effects
func GetUserAndLog(id int64) error {
	if err := GetUser(id); err != nil {
		return err
	}
	return LogAction("get_user")
}

// Valid: declares all effects including those from implicit function
// dirty: { select[user] | insert[log] | select[session] }
func ValidCaller(userID int64) error {
	// This has its own effect
	// SELECT * FROM sessions WHERE user_id = ?

	// Call function with implicit effects
	return GetUserAndLog(userID)
}

// Invalid: missing effects from implicit function
// dirty: { select[session] }
func InvalidCaller(userID int64) error {
	// SELECT * FROM sessions WHERE user_id = ?

	// This call brings in select[user] and insert[log] implicitly
	return GetUserAndLog(userID) // want "function calls GetUserAndLog which has effects \\[insert\\[log\\], select\\[user\\]\\] not declared in this function"
}

// Chain of implicit effects
func ChainA() error {
	return GetUser(1)
}

func ChainB() error {
	return ChainA()
}

func ChainC() error {
	return ChainB()
}

// Invalid: missing effect from deep chain
// dirty: { insert[log] }
func InvalidChain() error {
	LogAction("start")
	return ChainC() // want "function calls ChainC which has effects \\[select\\[user\\]\\] not declared in this function"
}

// Circular dependencies
func CircularA() error {
	return CircularB()
}

func CircularB() error {
	GetUser(1)
	return CircularA()
}

// Invalid: missing effect from circular dependency
// dirty: { insert[log] }
func InvalidCircular() error {
	LogAction("circular")
	return CircularA() // want "function calls CircularA which has effects \\[select\\[user\\]\\] not declared in this function"
}
