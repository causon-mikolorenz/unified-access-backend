package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/jmoiron/sqlx"
)

type AuthCodeRepository struct {
	db *sqlx.DB
}

// StoreCode saves the generated code
func (r *AuthCodeRepository) StoreCode(code string, userID []byte, clientID []byte, redirectURI string) error {
	query := `INSERT INTO auth_codes (code, user_id, client_id, redirect_uri, expires_at) 
              VALUES (?, ?, ?, ?, ?)`
	expiresAt := time.Now().Add(5 * time.Minute) // Codes are very short-lived
	_, err := r.db.Exec(query, code, userID, clientID, redirectURI, expiresAt)
	return err
}

// ExchangeCode uses a transaction to find, lock, and "consume" the code
func (r *AuthCodeRepository) ExchangeCode(code string) (*models.AuthorizationCode, error) {
	tx, err := r.db.Beginx() // Start transaction
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Safety: rollback if we return early without committing

	var authCode models.AuthorizationCode
	// FOR UPDATE locks the row so no other process can read/write it until we are done
	query := `SELECT code, user_id, client_id, redirect_uri, expires_at, used_at 
              FROM auth_codes WHERE code = ? FOR UPDATE`

	err = tx.Get(&authCode, query, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("code not found")
		}
		return nil, err
	}

	// SECURITY CHECKS
	if authCode.Used != true {
		return nil, errors.New("code already exchanged")
	}
	if time.Now().After(authCode.ExpiresAt) {
		return nil, errors.New("code expired")
	}

	// CONSUME: Mark as used
	_, err = tx.Exec("UPDATE auth_codes SET used_at = NOW() WHERE code = ?", code)
	if err != nil {
		return nil, err
	}

	// Commit the transaction to release the lock
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &authCode, nil
}
