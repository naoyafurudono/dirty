package complex

// Test case: edge cases and special scenarios

// Multiple effects on same target
//dirty: select[user], update[user], delete[user]
func ManageUser(id int64, action string) error {
	// Various operations on user table
	return nil
}

// Empty effect declaration (should be valid but warns about missing effects from calls)
//dirty:
func EmptyEffectDeclaration() error {
	return ManageUser(1, "delete") // want "function calls ManageUser which has effects \\[delete\\[user\\], select\\[user\\], update\\[user\\]\\] not declared in this function"
}

// No effect comment at all - not checked
func NoEffectComment() error {
	return ManageUser(1, "update") // No error - function has no //dirty: comment
}

// Malformed effect syntax (should be reported as error)
//dirty: select(user)
func MalformedEffect() error {
	// Should report syntax error for using () instead of []
	return nil
}

// Effect with special characters in target
//dirty: select[user_profile], update[user-settings], insert[user.preferences]
func SpecialCharacterTargets() error {
	// Valid: underscores, hyphens, dots in target names
	return nil
}

// Recursive call
//dirty: select[tree_node]
func TraverseTree(nodeID int64) error {
	// Base case
	if nodeID == 0 {
		return nil
	}
	
	// Recursive case - effect is already declared
	return TraverseTree(nodeID - 1)
}

// Conditional effects
//dirty: select[user], select[admin], insert[log]
func ConditionalEffects(isAdmin bool, userID int64) error {
	if isAdmin {
		// SELECT * FROM admin WHERE id = ?
		_ = userID
	} else {
		// SELECT * FROM user WHERE id = ?
		_ = userID
	}
	// INSERT INTO log ...
	return nil
}