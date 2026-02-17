package repository

import (
	"encoding/json"
	"fmt"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

// GetUserList retrieves a paginated list of non-deleted users.
func (r *UserRepository) GetUserList(limit, offset int) ([]models.User, error) {
	var users []models.User
	query := `
		SELECT id, username, first_name, middle_name, last_name,
		       email, status, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL 
		LIMIT ? OFFSET ?`

	err := r.db.Select(&users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// GetUserByEmail finds a user by email, including the hash for auth logic.
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, first_name, middle_name, last_name, 
		       email, password_hash, status, created_at, updated_at
		FROM users
		WHERE email = ? AND deleted_at IS NULL`

	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserById retrieves a specific user by binary UUID.
func (r *UserRepository) GetUserById(id []byte) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, first_name, middle_name, last_name, 
		       email, status, created_at, updated_at
		FROM users
		WHERE id = ? AND deleted_at IS NULL`

	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser executes a stored procedure to handle User and Roles atomically.
func (r *UserRepository) CreateUser(u *models.User) error {
	rolesJSON, err := json.Marshal(u.Roles)
	if err != nil {
		return fmt.Errorf("failed to marshal user roles: %w", err)
	}

	query := `CALL CreateUser(?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.Exec(query,
		u.ID,
		u.Username,
		u.FirstName,
		u.MiddleName,
		u.LastName,
		u.Email,
		u.PasswordHash,
		rolesJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to execute CreateUser procedure: %w", err)
	}
	return nil
}

// UpdateStatus changes the user's active/inactive state.
func (r *UserRepository) UpdateStatus(id []byte, status string) error {
	query := `UPDATE users SET status = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

// GetRoles fetches the roles assigned to a user via the junction table.
func (r *UserRepository) GetRoles(userID []byte) ([]models.Role, error) {
	var roles []models.Role
	query := `
		SELECT r.id, r.role_name, r.description 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.deleted_at IS NULL`

	err := r.db.Select(&roles, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// SoftDelete marks the user as deleted for forensic record keeping.
func (r *UserRepository) SoftDelete(id []byte) error {
	query := `CALL ArchiveUser(?)`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UserRepository) CountUsers() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := r.db.Get(&count, query)
	return count, err
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}
