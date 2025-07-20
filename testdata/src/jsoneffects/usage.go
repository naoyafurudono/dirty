package jsoneffects

// Test JSON-based effect declarations

// dirty: { custom[effect] }
func ProcessUserCorrect(id int64) error {
	user, err := GetUser(id) // ✓ OK: custom[effect] is declared
	if err != nil {
		return err
	}
	_ = user
	return nil
}

// dirty: { insert[logs] }
func ProcessUserWrong(id int64) error {
	// GetUser is defined locally with custom[effect], source code takes priority
	user, err := GetUser(id) // want `function calls GetUser which has effects \[custom\[effect\]\] not declared in this function`
	if err != nil {
		return err
	}
	_ = user
	return nil
}

// dirty: { select[cache] }
func ComplexOperationWrong() {
	// ComplexQuery defined in JSON with multiple effects
	ComplexQuery() // want `function calls ComplexQuery which has effects \[insert\[activity_logs\], select\[organizations\], select\[users\], update\[member_counts\]\] not declared in this function`
}

// Without dirty declaration - should compute effects implicitly
func ImplicitEffects() {
	CreateUser("test") // This should make ImplicitEffects have insert[users] effect
}

// dirty: { select[data] }
func CallImplicitWrong() {
	ImplicitEffects() // want `function calls ImplicitEffects which has effects \[insert\[users\]\] not declared in this function`
}

// Test empty effects
// dirty: { select[data] | network[api] }
func CallValidateWrong() {
	ValidateInput() // ✓ OK: ValidateInput has empty effects, so any function can call it
}

// Test JSON declaration override by source code
// dirty: { custom[effect] }
func GetUser(id int64) (string, error) {
	// This function is also defined in JSON, but source code declaration takes priority
	return "", nil
}

// dirty: { }
func TestOverride() {
	GetUser(1) // want `function calls GetUser which has effects \[custom\[effect\]\] not declared in this function`
}

// Test JSON-only functions
// dirty: { insert[logs] }
func CallExternalAPI() {
	ExternalAPICall() // want `function calls ExternalAPICall which has effects \[insert\[logs\], network\[external_api\], select\[cache\]\] not declared in this function`
}

// Test calling JSON-defined function correctly
// dirty: { select[users] | update[balance] | insert[transactions] | network[payment_api] }
func CallProcessPaymentCorrect() {
	ProcessPayment() // ✓ OK: all effects are declared
}

// Functions that are defined in JSON but not called (or as stubs)
func CreateUser(name string) {}
func ComplexQuery() {}
func ValidateInput() {}
func ExternalAPICall() {}
func ProcessPayment() {}
