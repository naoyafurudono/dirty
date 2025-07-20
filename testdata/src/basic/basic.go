package basic

// Valid: function declares its own effect
//dirty: select[user]
func GetUser(id int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// Valid: function declares all effects from called functions
//dirty: select[user], select[member]
func GetUserWithMembers(userID int64) error {
	err := GetUser(userID)
	if err != nil {
		return err
	}
	
	// SELECT * FROM members WHERE user_id = ?
	return nil
}

// Invalid: missing effect declaration from called function
//dirty: select[member]
func GetMemberOnly(userID int64) error {
	err := GetUser(userID) // want "function calls GetUser which has effect select\\[user\\] not declared in this function"
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

// Invalid: calling function with effects without declaring them
func MissingAllEffects(userID int64) error {
	return GetUserWithMembers(userID) // want "function calls GetUserWithMembers which has effects select\\[user\\], select\\[member\\] not declared in this function"
}