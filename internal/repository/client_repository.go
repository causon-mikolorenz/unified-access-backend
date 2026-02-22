package repository

import (
	"fmt"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/jmoiron/sqlx"
)

type ClientRepository struct {
	db *sqlx.DB
}

// GetByID retrieves a specific client by its binary UUID.
// @Summary Get Client by ID
// @ID get-client-by-id
// @Accept json
// @Produce json
// @Param id query []byte true "Client Binary ID"
// @Success 200 {object} models.Client
func (r *ClientRepository) GetByID(id []byte) (*models.Client, error) {
	var client models.Client
	query := `
        SELECT id, client_name, abbreviation, description, 
               image_location, base_url, redirect_uri, logout_uri, updated_at
        FROM clients 
        WHERE id = ? AND deleted_at IS NULL`

	err := r.db.Get(&client, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return &client, nil
}

// ListClients returns a paginated list of non-deleted service providers.
// @Summary List Clients
// @ID list-clients
// @Param limit query int true "Limit"
// @Param offset query int true "Offset"
// @Success 200 {array} models.Client
func (r *ClientRepository) ListClients(limit, offset int) ([]models.Client, error) {
	var clients []models.Client
	query := `
        SELECT id, client_name, abbreviation, description, image_location 
        FROM clients 
        WHERE deleted_at IS NULL 
        LIMIT ? OFFSET ?`

	err := r.db.Select(&clients, query, limit, offset)
	return clients, err
}

// CreateClient handles atomic insertion of client, grants, and prefixed roles.
// @Summary Create Client
// @ID create-client
func (r *ClientRepository) CreateClient(
	client *models.Client,
	grants []string,
) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Client
	q1 := `INSERT INTO clients (id, client_name, abbreviation, client_secret, 
               base_url, redirect_uri, logout_uri, description, image_location) 
           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(q1, client.ID, client.ClientName, client.Abbreviation,
		client.ClientSecret, client.BaseUrl, client.RedirectUri,
		client.LogoutUri, client.Description, client.ImageLocation)
	if err != nil {
		return err
	}

	// 2. Insert Grant Types
	q2 := `INSERT INTO client_grant_types (client_id, grant_type) VALUES (?, ?)`
	for _, g := range grants {
		if _, err = tx.Exec(q2, client.ID, g); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateClient modifies safe fields only. Abbreviation and Secret are locked.
// @Summary Update Client
// @ID update-client
func (r *ClientRepository) UpdateClient(c *models.Client) error {
	query := `
        UPDATE clients 
        SET client_name = ?, description = ?, image_location = ?, 
            base_url = ?, redirect_uri = ?, logout_uri = ?
        WHERE id = ? AND deleted_at IS NULL`

	_, err := r.db.Exec(query, c.ClientName, c.Description, c.ImageLocation,
		c.BaseUrl, c.RedirectUri, c.LogoutUri, c.ID)
	return err
}

// SoftDelete marks a client as deleted without removing the audit trail.
// @Summary Soft Delete Client
// @ID delete-client
func (r *ClientRepository) SoftDelete(id []byte) error {
	query := `UPDATE clients SET deleted_at = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ClientRepository) GetGrantTypes(clientID []byte) ([]string, error) {
	var grants []string
	query := `SELECT grant_type FROM client_grant_types WHERE client_id = ?`
	err := r.db.Select(&grants, query, clientID)
	return grants, err
}

func (r *ClientRepository) GetClientRoles(abbr string) ([]string, error) {
	var roles []string
	// Using LIKE to find all roles prefixed with the abbreviation
	query := `SELECT role_name FROM roles WHERE role_name LIKE ?`
	err := r.db.Select(&roles, query, abbr+":%")
	return roles, err
}

func (r *ClientRepository) ListClientBaseURLS() ([]string, error) {
	var baseURLS []string
	query := `SELECT base_url FROM clients WHERE deleted_at IS NULL`
	err := r.db.Select(&baseURLS, query)
	if err != nil {
		return nil, err
	}
	return baseURLS, nil
}

func NewClientRepository(db *sqlx.DB) *ClientRepository {
	return &ClientRepository{db: db}
}
