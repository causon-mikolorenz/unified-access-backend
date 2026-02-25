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
	rolesJSON, err := json.Marshal(u.RoleString)
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
func (r *UserRepository) UpdateStatus(user *models.User) error {
	query := `UPDATE users SET status = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.Exec(query, string(user.Status), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

// UpdateUserPassword calls the procedure for updating user's password.
func (r *UserRepository) UpdateUserPassword(user *models.User) error {
	query := `CALL UpdateUserPassword(?, ?)`

	_, err := r.db.Exec(query, user.ID, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// UpdateUserRoles updates the roles of a specific user based on the role IDs
func (r *UserRepository) UpdateUserRoles(userID []byte, roleIDs []int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// if some roles are deleted but not all
	if len(roleIDs) > 0 {
		deleteQuery, args, _ := sqlx.In(
			`DELETE FROM user_roles WHERE user_id = ? AND role_id NOT IN (?)`, 
			userID, 
			roleIDs,
		)
		if _, err := tx.Exec(tx.Rebind(deleteQuery), args...); err != nil {
            return fmt.Errorf("failed to delete user roles: %w", err)
        }
    } else {
        // If roleIDs is empty, remove all roles for the user
        deleteAll := "DELETE FROM user_roles WHERE user_id = ?"
        if _, err := tx.Exec(deleteAll, userID); err != nil {
            return fmt.Errorf("failed to delete all roles from user: %w", err)
        }
    }

	insertQuery := `
        INSERT INTO user_roles (user_id, role_id) 
        VALUES (?, ?) 
        ON DUPLICATE KEY UPDATE role_id = role_id`

    for _, rid := range roleIDs {
        if _, err := tx.Exec(insertQuery, userID, rid); err != nil {
            return fmt.Errorf("upsert error: %w", err)
        }
    }
	
	return tx.Commit()
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
