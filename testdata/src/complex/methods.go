package complex

// Test case: methods with effects

type UserRepository struct {
	db string
}

// dirty: { select[user] }
func (r *UserRepository) FindByID(id int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// dirty: { insert[user] }
func (r *UserRepository) Create(name string) error {
	// INSERT INTO users (name) VALUES (?)
	return nil
}

// dirty: { update[user] }
func (r *UserRepository) UpdateName(id int64, name string) error {
	// UPDATE users SET name = ? WHERE id = ?
	return nil
}

type UserService struct {
	repo *UserRepository
}

// Valid: declares all effects from repository methods
// dirty: { select[user] | update[user] | insert[audit_log] }
func (s *UserService) UpdateUserWithAudit(id int64, name string) error {
	// Check user exists
	if err := s.repo.FindByID(id); err != nil {
		return err
	}

	// Update user
	if err := s.repo.UpdateName(id, name); err != nil {
		return err
	}

	// Log the change (INSERT INTO audit_log)
	return nil
}

// Invalid: missing select[user] effect
// dirty: { update[user] }
func (s *UserService) UpdateUserBroken(id int64, name string) error {
	if err := s.repo.FindByID(id); err != nil { // want "function calls FindByID which has effects \\[select\\[user\\]\\] not declared in this function"
		return err
	}

	if err := s.repo.UpdateName(id, name); err != nil {
		return err
	}

	return nil
}
