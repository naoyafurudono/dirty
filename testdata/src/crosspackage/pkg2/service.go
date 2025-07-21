package pkg2

import (
	"crosspackage/pkg1"
	"fmt"
)

type ProcessedUser struct {
	ID       int
	FullName string
}

// dirty: { select[users] | transform }
func ProcessUser(id int) ProcessedUser {
	user := pkg1.GetUser(id) // Uses pkg1's GetUser which has { select[users] }
	return ProcessedUser{
		ID:       user.ID,
		FullName: fmt.Sprintf("Processed: %s", user.Name),
	}
}

// dirty: { transform }
func ProcessUserIncorrect(id int) ProcessedUser {
	// ERROR: Missing { select[users] } effect from pkg1.GetUser
	user := pkg1.GetUser(id) // want `function ProcessUserIncorrect requires effects \{ select\[users\] \} but declares only \{ transform \}`
	return ProcessedUser{
		ID:       user.ID,
		FullName: fmt.Sprintf("Processed: %s", user.Name),
	}
}

// dirty: { insert[users] | event[user.created] }
func CreateAndNotify(name string) {
	user := pkg1.CreateUser(name) // Uses pkg1's CreateUser which has { insert[users] }
	fmt.Printf("User created: %v\n", user)
	// Simulated event emission
}

// No dirty comment - should be calculated implicitly
func GetProcessedUserName(id int) string {
	// Should require { select[users] } from pkg1.GetUserName
	return pkg1.GetUserName(id)
}
