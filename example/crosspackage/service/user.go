package service

import (
	"fmt"

	"github.com/naoyafurudono/dirty/example/crosspackage/db"
)

// UserService provides user-related operations
type UserService struct{}

// dirty: { select[users] }
func (s *UserService) GetUser(id int64) (*db.User, error) {
	// Correctly declares the effect from db.GetUserByID
	return db.GetUserByID(id)
}

// dirty: { insert[users] | event[user.created] }
func (s *UserService) CreateUser(name, email string) (*db.User, error) {
	user, err := db.CreateUser(name, email)
	if err != nil {
		return nil, err
	}

	// Emit event (simulated)
	fmt.Printf("Event: user.created for %s\n", user.Name)

	return user, nil
}

// This function is missing the required effect declaration
// It should declare { select[users] | update[users] }
// dirty: { update[users] }
func (s *UserService) UpdateUserEmailIfExists(id int64, email string) error {
	// ERROR: Missing { select[users] } effect
	user, err := db.GetUserByID(id)
	if err != nil {
		return err
	}

	if user != nil {
		return db.UpdateUserEmail(id, email)
	}

	return fmt.Errorf("user not found")
}
