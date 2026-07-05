package repository

import (
	"task-system/internal/models"
	"task-system/internal/utils"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var ErrDuplicateEmail = utils.ErrConflict

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a user using named query mappings and struct scanning, avoiding manual column scanning
func (r *UserRepository) Create(user *models.User) error {
	// Pre-check validation: check if the email already exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := r.db.Get(&exists, checkQuery, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}

	// Insert the record. sqlx NamedQuery automatically maps ":name", ":email", and ":password"
	// to the matching db tags on the "user" struct parameter.
	insertQuery := `
		INSERT INTO users (name, email, password)
		VALUES (:name, :email, :password)
		RETURNING id, created_at, updated_at
	`
	rows, err := r.db.NamedQuery(insertQuery, user)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" { // unique_violation (duplicate email fallback check)
				return ErrDuplicateEmail
			}
		}
		return err
	}
	defer rows.Close()

	if rows.Next() {
		// StructScan automatically reads RETURNING columns (id, created_at, updated_at)
		// and maps them directly back onto your "user" struct in memory!
		return rows.StructScan(user)
	}

	return nil
}


func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := "SELECT id,name,email,password,created_at,updated_at FROM users WHERE email = $1"
	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	query := "SELECT id,name,email,password,created_at,updated_at FROM users WHERE id = $1"
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user's name, email, and password in the database
func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users 
		SET name = :name, email = :email, password = :password, updated_at = extract(epoch from now())::bigint
		WHERE id = :id
	`
	_, err := r.db.NamedExec(query, user)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

// DeleteUser deletes a user from the database by ID
func (r *UserRepository) DeleteUser(id string) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}