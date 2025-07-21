// Package main demonstrates using JSON-based effect declarations
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

// GetUserFromDB retrieves a user from the database (stub)
func GetUserFromDB(_ string) error { return nil }

// CreateUserInDB creates a new user in the database (stub)
func CreateUserInDB(_ string) error { return nil }

// UpdateUserStatusInDB updates user status in the database (stub)
func UpdateUserStatusInDB(_, _ string) error { return nil }

// DeleteUserFromDB deletes a user from the database (stub)
func DeleteUserFromDB(_ string) error { return nil }

// SendEmailNotification sends an email notification (stub)
func SendEmailNotification(_ string) error { return nil }

// CallExternalAPI is a stub function for external API calls
func CallExternalAPI() (string, error) { return "", nil }

// ValidateUserInput validates user input (stub)
func ValidateUserInput(_ string) error { return nil }

// ComputeHash computes a hash of the input data (stub)
func ComputeHash(_ string) string { return "" }

// LogActivity logs an activity event (stub)
func LogActivity(_ string) {}
