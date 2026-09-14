package user

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) FindByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	const query = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = ?
		LIMIT 1
	`

	var u User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) Create(
	ctx context.Context,
	email string,
	passwordHash string,
) (*User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES (?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		email,
		passwordHash,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uint64(id),
		Email:        email,
		PasswordHash: passwordHash,
	}, nil
}
