package db

// User represents a user in the database
type User struct {
	ID    int64
	Name  string
	Email string
}

// dirty: { select[users] }
func GetUserByID(id int64) (*User, error) {
	// Simulated database query
	return &User{ID: id, Name: "John Doe", Email: "john@example.com"}, nil
}

// dirty: { insert[users] }
func CreateUser(name, email string) (*User, error) {
	// Simulated database insert
	return &User{ID: 1, Name: name, Email: email}, nil
}

// dirty: { update[users] }
func UpdateUserEmail(id int64, email string) error {
	// Simulated database update
	return nil
}

// dirty: { delete[users] }
func DeleteUser(id int64) error {
	// Simulated database delete
	return nil
}
