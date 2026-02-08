package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/jmoiron/sqlx"
)

type AuthCodeRepository struct {
	db *sqlx.DB
}

const (
	YEARS  = 0
	MONTHS = 0
	DAYS   = 7
)

// StoreCode saves the generated code
func (r *AuthCodeRepository) StoreCode(code string, userID []byte, clientID []byte, redirectURI string) error {
	query := `INSERT INTO authorization_codes (code, user_id, client_id, redirect_uri, expires_at) 
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
              FROM authorization_codes WHERE code = ? FOR UPDATE`

	err = tx.Get(&authCode, query, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("code not found")
		}
		return nil, err
	}

	// SECURITY CHECKS
	if authCode.UsedAt.Valid {
		return nil, errors.New("code already exchanged")
	}
	if time.Now().After(authCode.ExpiresAt) {
		return nil, errors.New("code expired")
	}

	// CONSUME: Mark as used
	_, err = tx.Exec("UPDATE authorization_codes SET used_at = NOW() WHERE code = ?", code)
	if err != nil {
		return nil, err
	}

	// Commit the transaction to release the lock
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &authCode, nil
}

func (r *AuthCodeRepository) GetUserForAuth(email string) (*models.UserClaims, string, error) {
	var row struct {
		ID           []byte `db:"id"`
		Username     string `db:"username"`
		FirstName    string `db:"first_name"`
		MiddleName   string `db:"middle_name"`
		LastName     string `db:"last_name"`
		Email        string `db:"email"`
		PasswordHash string `db:"password_hash"`
		Status       string `db:"status"`
		RolesString  string `db:"roles"`
	}

	query := `CALL GetUserForAuth(?)`
	err := r.db.Get(&row, query, email)
	if err != nil {
		return nil, "", err
	}

	// Convert comma-separated string back to a slice for UserClaims
	roles := strings.Split(row.RolesString, ",")

	claims := &models.UserClaims{
		UserID:     row.ID,
		Username:   row.Username,
		Email:      row.Email,
		FirstName:  row.FirstName,
		MiddleName: row.MiddleName,
		LastName:   row.LastName,
		Roles:      roles,
	}

	return claims, row.PasswordHash, nil
}

// VerifyClient checks if the client credentials are valid
func (r *AuthCodeRepository) VerifyClient(clientID []byte, clientSecret string) (bool, error) {
	var storedHash string
	query := `SELECT client_secret FROM clients WHERE id = ?`

	err := r.db.Get(&storedHash, query, clientID)
	if err != nil {
		return false, err
	}

	// Use internal/auth/hash.go utility
	err = auth.CompareSecret(storedHash, clientSecret)
	return err == nil, nil
}

// GetClaimsById is similar to GetUserForAuth but uses user UUID
func (r *AuthCodeRepository) GetClaimsByID(userId []byte) (*models.UserClaims, error) {
	var row struct {
		ID          []byte `db:"id"`
		Username    string `db:"username"`
		FirstName   string `db:"first_name"`
		MiddleName  string `db:"middle_name"`
		LastName    string `db:"last_name"`
		Email       string `db:"email"`
		Status      string `db:"status"`
		RolesString string `db:"roles"`
	}

	// FIX: Filter by u.id and use ? placeholder
	query := `
        SELECT 
            u.id, u.email, u.first_name, u.middle_name, u.last_name,
            u.username, u.status,
            IFNULL((SELECT GROUP_CONCAT(r.role_name) 
            FROM user_roles ur 
            JOIN roles r ON ur.role_id = r.id 
            WHERE ur.user_id = u.id), '') as roles
        FROM users u
        WHERE u.id = ? AND u.status != 'deleted'
        LIMIT 1`

	err := r.db.Get(&row, query, userId)
	if err != nil {
		return nil, err
	}

	var roles []string
	if row.RolesString != "" {
		roles = strings.Split(row.RolesString, ",")
	} else {
		roles = []string{}
	}

	return &models.UserClaims{
		UserID:     row.ID,
		Username:   row.Username,
		Email:      row.Email,
		FirstName:  row.FirstName,
		MiddleName: row.MiddleName,
		LastName:   row.LastName,
		Roles:      roles,
	}, nil
}

func (r *AuthCodeRepository) StoreRefreshToken(token string, userID []byte, clientID []byte) error {
	expirationDate := time.Now().AddDate(YEARS, MONTHS, DAYS)
	query := `
		INSERT INTO refresh_tokens(token, client_id, user_id, expires_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, token, clientID, userID, expirationDate)
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthCodeRepository) RotateRefreshToken(oldToken, newToken string) error {
	newExpiresAt := time.Now().AddDate(YEARS, MONTHS, DAYS)

	query := `CALL RotateRefreshToken(?, ?, ?)`
	_, err := r.db.Exec(query, oldToken, newToken, newExpiresAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthCodeRepository) GetIDsFromToken(token string) ([]byte, []byte, error) {
    var IDs struct {
        UserID   []byte `db:"user_id"`
        ClientID []byte `db:"client_id"`
    }

    query := `
        SELECT user_id, client_id FROM refresh_tokens
        WHERE token = ?
    `

    err := r.db.Get(&IDs, query, token)
    if err != nil {
        return nil, nil, err
    }

    return IDs.UserID, IDs.ClientID, nil
}

func NewAuthCodeRepository(db *sqlx.DB) *AuthCodeRepository {
	return &AuthCodeRepository{
		db: db,
	}
}
