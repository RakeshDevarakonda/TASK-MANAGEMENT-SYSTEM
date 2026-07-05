package repository

import (
	"task-system/internal/models"

	"github.com/jmoiron/sqlx"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create inserts a new session row
func (r *SessionRepository) Create(session *models.Session) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES (:user_id, :token, :expires_at)
		RETURNING id, created_at
	`
	rows, err := r.db.NamedQuery(query, session)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(session)
	}
	return nil
}

// GetByToken fetches a session record by its refresh token string
func (r *SessionRepository) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	query := "SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token = $1"
	err := r.db.Get(&session, query, token)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByUserID fetches the active session record for a given user ID
func (r *SessionRepository) GetByUserID(userID string) (*models.Session, error) {
	var session models.Session
	query := "SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE user_id = $1"
	err := r.db.Get(&session, query, userID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession updates the stored encrypted refresh token and extends its expiration date
func (r *SessionRepository) UpdateSession(id string, token string, expiresAt int64) error {
	query := "UPDATE sessions SET token = $1, expires_at = $2 WHERE id = $3"
	_, err := r.db.Exec(query, token, expiresAt, id)
	return err
}

// DeleteByUserID deletes all active sessions for a user (invalidates other devices)
func (r *SessionRepository) DeleteByUserID(userID string) error {
	query := "DELETE FROM sessions WHERE user_id = $1"
	_, err := r.db.Exec(query, userID)
	return err
}

// DeleteByToken deletes a session by token value (logout this device)
func (r *SessionRepository) DeleteByToken(token string) error {
	query := "DELETE FROM sessions WHERE token = $1"
	_, err := r.db.Exec(query, token)
	return err
}
