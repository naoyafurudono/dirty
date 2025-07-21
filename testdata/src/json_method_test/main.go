package main

// dirty: { print }
func main() {
	userSvc := &UserService{}
	orderSvc := &OrderService{}

	// Same method name on different types
	userSvc.GetUser()
	orderSvc.GetUser()

	// Value receiver method
	userSvc.GetUserByValue()
}
