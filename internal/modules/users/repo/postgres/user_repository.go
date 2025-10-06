package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/domain"
)

type UserRepository struct{ DB *sql.DB }

func (repository *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (int64, error) {
	var id int64
	xid := uuid.NewString()
	err := repository.DB.QueryRowContext(ctx,
		`INSERT INTO users(xid,name,email,password) VALUES($1,$2,$3,$4) RETURNING id`,
		xid, name, email, passwordHash,
	).Scan(&id)
	return id, err
}

func (repository *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var (
		createdAt sql.NullTime
		updatedAt sql.NullTime
		user      domain.User
	)
	err := repository.DB.QueryRowContext(ctx,
		`SELECT id, xid, name, email, password, is_active, created_at, updated_at
         FROM users WHERE email=$1 AND deleted_at IS NULL LIMIT 1`, email,
	).Scan(&user.ID, &user.XID, &user.Name, &user.Email, &user.Password, &user.IsActive, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		t := createdAt.Time
		user.CreatedAt = &t
	}
	if updatedAt.Valid {
		t := updatedAt.Time
		user.UpdatedAt = &t
	}
	return &user, nil
}
