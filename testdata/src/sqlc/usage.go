package sqlc

import "context"

// Test cases for sqlc-use integration

// Valid: declares all effects from sqlc
//dirty: select[users], insert[logs]
func ProcessUserWithLogging(ctx context.Context, q *Queries, id int64) error {
	user, err := q.GetUser(ctx, id) // sqlc-use should detect: select[users]
	if err != nil {
		return err
	}
	return logUserAccess(user.ID) // manual: insert[logs]
}

// Invalid: missing select[users] from sqlc
//dirty: insert[logs]
func ProcessUserBroken(ctx context.Context, q *Queries, id int64) error {
	user, err := q.GetUser(ctx, id) // want "function calls GetUser which has effects \\[select\\[users\\]\\] not declared in this function"
	if err != nil {
		return err
	}
	return logUserAccess(user.ID)
}

// Invalid: missing multiple effects from sqlc
//dirty: select[users]
func ComplexOperationBroken(ctx context.Context, q *Queries, orgID int64) error {
	// ComplexQuery has: select[users], select[organizations], update[member_counts], insert[activity_logs]
	return q.ComplexQuery(ctx, orgID) // want "function calls ComplexQuery which has effects \\[insert\\[activity_logs\\], select\\[organizations\\], select\\[users\\], update\\[member_counts\\]\\] not declared in this function"
}

// Valid: no declaration needed for functions without //dirty: comment
func ImplicitSQLCEffects(ctx context.Context, q *Queries, id int64) error {
	return q.UpdateUserStatus(ctx, id, "active") // No error - function has no //dirty: comment
}

// Invalid: empty declaration but calls sqlc function
//dirty:
func EmptyWithSQLCCall(ctx context.Context, q *Queries) error {
	_, err := q.CreateUser(ctx, "test") // want "function calls CreateUser which has effects \\[insert\\[users\\]\\] not declared in this function"
	return err
}

// Helper function with manual effect
//dirty: insert[logs]
func logUserAccess(userID int64) error {
	// INSERT INTO access_logs ...
	return nil
}
