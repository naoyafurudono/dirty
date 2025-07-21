package nofacts

// Test basic functionality without cross-package dependencies

// dirty: { select[user] }
func GetUser(id int) string {
	return "user"
}

// dirty: { select[user] | transform }
func ProcessUser(id int) string {
	user := GetUser(id) // OK: has select[user]
	return "processed " + user
}

// dirty: { transform }
func ProcessUserIncorrect(id int) string {
	// This should fail because GetUser has select[user] effect
	user := GetUser(id) // want `function calls GetUser which has effects \[select\[user\]\] not declared in this function`
	return "processed " + user
}
