package pkg1

type User struct {
	ID   int
	Name string
}

// dirty: { select[users] }
func GetUser(id int) User {
	// Database query simulation
	return User{ID: id, Name: "User"}
}

// dirty: { insert[users] }
func CreateUser(name string) User {
	// Database insert simulation
	return User{ID: 1, Name: name}
}

// dirty: { update[users] }
func UpdateUser(user User) error {
	// Database update simulation
	return nil
}

// No dirty comment - should be calculated implicitly
func GetUserName(id int) string {
	user := GetUser(id) // This should require { select[users] }
	return user.Name
}
