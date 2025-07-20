package main

import "fmt"

// Example: Using JSON-based effect declarations
//
// This example shows how to use JSON files to declare effects for functions
// that are defined elsewhere (e.g., generated code, external libraries).

// ProcessUserRegistration correctly declares all effects
// dirty: { select[users] | insert[users] | insert[audit_logs] | network[smtp] | io[filesystem] | insert[activity_logs] }
func ProcessUserRegistration(email string) error {
	// Check if user exists
	if err := GetUserFromDB(email); err == nil {
		return fmt.Errorf("user already exists")
	}

	// Validate input (has empty effects)
	if err := ValidateUserInput(email); err != nil {
		return err
	}

	// Create user
	if err := CreateUserInDB(email); err != nil {
		return err
	}

	// Send welcome email
	if err := SendEmailNotification(email); err != nil {
		// Log the error but don't fail registration
		LogActivity("email_send_failed")
	}

	return nil
}

// ProcessUserUpdate is missing some effects
// dirty: { update[users] }
func ProcessUserUpdate(userID, status string) error {
	// This will cause an error because UpdateUserStatusInDB
	// also has insert[audit_logs] effect which is not declared
	return UpdateUserStatusInDB(userID, status)
}

// ProcessWithExternalCall correctly declares network effects
// dirty: { network[external_api] | select[cache] | insert[logs] | insert[activity_logs] }
func ProcessWithExternalCall() error {
	// Call external API
	data, err := CallExternalAPI()
	if err != nil {
		return err
	}

	// Log the activity
	LogActivity(fmt.Sprintf("external_call_result: %s", data))

	return nil
}

// HelperFunction doesn't declare effects, so they are computed implicitly
func HelperFunction(userID string) error {
	// Computed effects will be: { delete[users] | insert[audit_logs] }
	return DeleteUserFromDB(userID)
}

// UseHelper must declare the effects from HelperFunction
// dirty: { delete[users] | insert[audit_logs] | insert[activity_logs] }
func UseHelper(userID string) error {
	if err := HelperFunction(userID); err != nil {
		return err
	}

	LogActivity("user_deleted")
	return nil
}

// Functions that are declared in JSON but defined as stubs here
func GetUserFromDB(email string) error { return nil }
func CreateUserInDB(email string) error { return nil }
func UpdateUserStatusInDB(userID, status string) error { return nil }
func DeleteUserFromDB(userID string) error { return nil }
func SendEmailNotification(email string) error { return nil }
func CallExternalAPI() (string, error) { return "", nil }
func ValidateUserInput(email string) error { return nil }
func ComputeHash(data string) string { return "" }
func LogActivity(activity string) {}
