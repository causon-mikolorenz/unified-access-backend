package repository

import (
	"fmt"
	"time"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/jmoiron/sqlx"
)

type SessionRepository struct {
	db *sqlx.DB
}

func (r *SessionRepository) Create(s *models.IdPSession) error {
	query := `
        INSERT INTO idp_sessions (session_id, user_id, ip_address, user_agent, expires_at)
        VALUES (?, ?, ?, ?, ?)
    `
	_, err := r.db.Exec(query, s.SessionId, s.UserId, s.IpAddress, s.UserAgent, s.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByID(sessionID string) (*models.IdPSession, error) {
	var session models.IdPSession
	query := `SELECT session_id, user_id, ip_address, user_agent, created_at, expires_at 
              FROM idp_sessions WHERE session_id = ?`

	err := r.db.Get(&session, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

func (r *SessionRepository) Delete(sessionID string) error {
	query := `DELETE FROM idp_sessions WHERE session_id = ?`
	_, err := r.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteExpired() (int64, error) {
	query := `DELETE FROM idp_sessions WHERE expires_at < ?`
	res, err := r.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to clean up expired sessions: %w", err)
	}
	return res.RowsAffected()
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}
