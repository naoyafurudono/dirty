package main

// Process in main package
func Process() string {
	return "main"
}

// SameNameFunc for testing
func SameNameFunc() {}

// dirty: { print }
func main() {
	Process() // This is local Process
	SameNameFunc()
}
