package basic

// Valid: function declares its own effect
// dirty: { select[user] }
func GetUser(id int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// Valid: function declares all effects from called functions
// dirty: { select[user] | select[member] }
func GetUserWithMembers(userID int64) error {
	err := GetUser(userID)
	if err != nil {
		return err
	}

	// SELECT * FROM members WHERE user_id = ?
	return nil
}

// Invalid: missing effect declaration from called function
// dirty: { select[member] }
func GetMemberOnly(userID int64) error {
	err := GetUser(userID) // want "function calls GetUser which has effects \\[select\\[user\\]\\] not declared in this function"
	if err != nil {
		return err
	}

	// SELECT * FROM members WHERE user_id = ?
	return nil
}

// Valid: function with no effects
func NoEffects() error {
	// Pure computation
	return nil
}

// Valid: function without effect declaration is not checked
func ImplicitEffects(userID int64) error {
	return GetUserWithMembers(userID) // No error - function has no // dirty: comment
}

// Invalid: function with empty effect declaration calling function with effects
// dirty: { }
func EmptyEffects(userID int64) error {
	return GetUserWithMembers(userID) // want "function calls GetUserWithMembers which has effects \\[select\\[member\\], select\\[user\\]\\] not declared in this function"
}
