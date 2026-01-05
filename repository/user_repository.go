package repository

import (
	"fmt"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT 
			id, 
			username, 
			first_name, 
			middle_name, 
			last_name, 
			email,
			password_hash, 
			status
		FROM users
		WHERE email = ?
	`
	err := r.db.Get(&user, query, email)

	return &user, err
}

func (r *UserRepository) GetById(id []byte) (*models.User, error) {
	var user models.User
	query := `
		SELECT 
			id, 
			username, 
			first_name, 
			middle_name, 
			last_name, 
			email, 
			status
		FROM users
		WHERE id = ?
	`
	err := r.db.Get(&user, query, id)

	return &user, err
}

func (r *UserRepository) Create(u *models.User) error {
	query := `CALL CreateUser(?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		u.ID,
		u.Username,
		u.FirstName,
		u.MiddleName,
		u.LastName,
		u.Email,
		u.PasswordHash,
	)
	if err != nil {
		return fmt.Errorf("Failed to execute procedure: %w", err)
	}

	return err
}

func (r *UserRepository) UpdateStatus(user *models.User, status models.UserStatus) error {
	query := `UPDATE users SET status = ? WHERE id = ?`

	_, err := r.db.Exec(query, status, user.ID)
	if err != nil {
		return fmt.Errorf("Failed to update status: %w", err)
	}

	return err
}

func (r *UserRepository) GetRoles(user *models.User) ([]models.Role, error) {
	var roles []models.Role
	query := `
		SELECT r.id, r.role_name, r.description 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`

	err := r.db.Select(&roles, query, user.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get user roles: %w", err)
	}

	return roles, nil
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
