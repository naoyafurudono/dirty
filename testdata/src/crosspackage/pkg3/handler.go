package pkg3

import (
	"crosspackage/pkg2"
	"fmt"
)

// dirty: { select[users] | transform | http_response }
func HandleGetUser(id int) {
	// Uses pkg2.ProcessUser which requires { select[users] | transform }
	processed := pkg2.ProcessUser(id)
	fmt.Printf("Response: %+v\n", processed)
}

// dirty: { http_response }
func HandleGetUserIncorrect(id int) {
	// ERROR: Missing { select[users] | transform } from pkg2.ProcessUser
	processed := pkg2.ProcessUser(id) // want `function HandleGetUserIncorrect requires effects \{ select\[users\] \| transform \} but declares only \{ http_response \}`
	fmt.Printf("Response: %+v\n", processed)
}

// dirty: { insert[users] | event[user.created] | http_response }
func HandleCreateUser(name string) {
	// Uses pkg2.CreateAndNotify which requires { insert[users] | event[user.created] }
	pkg2.CreateAndNotify(name)
	fmt.Println("User creation handled")
}

// Test transitive effect propagation
// No dirty comment - should calculate { select[users] } implicitly
func GetUserNameViaService(id int) string {
	// pkg2.GetProcessedUserName -> pkg1.GetUserName -> pkg1.GetUser
	// Should require { select[users] } transitively
	return pkg2.GetProcessedUserName(id)
}
