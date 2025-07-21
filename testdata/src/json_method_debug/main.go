package main

type Service struct{}

// dirty: { network }
func (s *Service) Process() {}

// dirty: { print }
func main() {
	svc := &Service{}
	svc.Process() // This should require { network }

	// Function with same name
	Process() // This should use JSON effects
}

func Process() {}
