package repository

import (
	"fmt"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/jmoiron/sqlx"
)

type ClientRepository struct {
	db *sqlx.DB
}

func (r *ClientRepository) GetByID(id []byte) (*models.Client, error) {
	var client models.Client
	query := `
		SELECT id, client_name FROM clients WHERE id = ?
	`

	err := r.db.Get(&client, query, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get client: %w", err)
	}

	return &client, nil
}

func (r *ClientRepository) Authenticate(id []byte,
	hash string,
) (*models.Client, error) {
	var client models.Client
	query := `
		SELECT id, client_name, client_secret FROM clients WHERE id = ? AND client_secret = ?
	`
	err := r.db.Get(&client, query, id, hash)
	if err != nil {
		return nil, fmt.Errorf("Failed to authenticate: %w", err)
	}

	return &client, err
}

func (r *ClientRepository) GetRedirectUrls(client *models.Client) ([]models.ClientUrl, error) {
	var urls []models.ClientUrl
	query := `
		SELECT id, client_id, redirect_url
		FROM client_urls WHERE client_id = ?
	`

	err := r.db.Select(&urls, query, client.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get client urls: %w", err)
	}

	return urls, err
}

func (r *ClientRepository) GetGrantTypes(client *models.Client) ([]models.ClientGrantTypes, error) {
	var grantTypes []models.ClientGrantTypes
	query := `
		SELECT client_id, grant_type
		FROM client_grant_types WHERE client_id = ?
	`

	err := r.db.Select(&grantTypes, query, client.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get client grant types: %w", err)
	}

	return grantTypes, err
}

func NewClientRepository(db *sqlx.DB) *ClientRepository {
	return &ClientRepository{
		db: db,
	}
}
