package pkg2

import "simpletest/pkg1"

// dirty: { transform }
func ProcessUser() string {
	user := pkg1.GetUser() // want `function calls simpletest/pkg1.GetUser which has effects \[select\[users\]\] not declared in this function`
	return "processed: " + user
}
