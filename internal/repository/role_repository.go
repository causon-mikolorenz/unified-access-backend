package repository

import (
    "fmt"
    "github.com/causon-mikolorenz/unified-access-backend/models"
    "github.com/jmoiron/sqlx"
)

type RoleRepository struct {
    db *sqlx.DB
}

// CreateRole adds a new role to the system.
// @Summary Create Role
// @ID create-role
func (r *RoleRepository) CreateRole(role models.Role) error {
    query := `
        INSERT INTO roles (role_name, description)
        VALUES (?, ?)
    `
    // Auto-commits immediately if not called within a transaction
    _, err := r.db.Exec(query, role.RoleName, role.Description)
    if err != nil {
        return fmt.Errorf("failed to create role: %w", err)
    }
    return nil
}

// GetByID retrieves a single role by its integer ID.
// @Summary Get Role by ID
// @ID get-role-by-id
func (r *RoleRepository) GetByID(id int) (*models.Role, error) {
    var role models.Role
    query := `
        SELECT (id, role_name, description, created_at, updated_at) 
        FROM roles WHERE id = ? AND deleted_at IS NULL`
    err := r.db.Get(&role, query, id)
    return &role, err
}

// ListRoles returns a paginated list of active roles.
// @Summary List Roles
// @ID list-roles
func (r *RoleRepository) ListRoles(limit, offset int) ([]models.Role, error) {
    var roles []models.Role
    query := `
        SELECT (id, role_name, description, created_at, updated_at) FROM roles 
        WHERE deleted_at IS NULL 
        ORDER BY id DESC 
        LIMIT ? OFFSET ?`
    err := r.db.Select(&roles, query, limit, offset)
    return roles, err
}

// UpdateRole modifies the description or name of an existing role.
// @Summary Update Role
// @ID update-role
func (r *RoleRepository) UpdateRole(role models.Role) error {
    query := `
        UPDATE roles 
        SET role_name = ?, description = ? 
        WHERE id = ? AND deleted_at IS NULL`
    _, err := r.db.Exec(query, role.RoleName, role.Description, role.ID)
    return err
}

// SoftDelete ensures audit integrity by hiding the role instead of purging it.
// @Summary Soft Delete Role
// @ID delete-role
func (r *RoleRepository) SoftDelete(id int) error {
    query := `UPDATE roles SET deleted_at = NOW() WHERE id = ?`
    _, err := r.db.Exec(query, id)
    return err
}

// CountRoles counts the number of roles that are not deleted.
// @Summary Count all active non deleted roles
// @ID count-roles
func (r *RoleRepository) CountRoles() (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM roles WHERE deleted_at IS NULL`
    err := r.db.Get(&count, query)
    return count, err
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
    return &RoleRepository{db: db}
}