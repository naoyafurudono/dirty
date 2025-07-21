package main

type UserService struct{}

func (s *UserService) GetUser() string {
	return "user"
}

func (s UserService) GetUserByValue() string {
	return "user by value"
}

type OrderService struct{}

func (o *OrderService) GetUser() string {
	return "order user"
}
