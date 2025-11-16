package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/domain"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type UserRepository struct{ DB *sql.DB }

func (repository *UserRepository) Create(ctx context.Context, name, email, passwordHash string, isActive bool) (int64, error) {
	var id int64
	xid := uuid.NewString()
	err := repository.DB.QueryRowContext(ctx,
		`INSERT INTO users(xid,name,email,password,is_active) VALUES($1,$2,$3,$4,$5) RETURNING id`,
		xid, name, email, passwordHash, isActive,
	).Scan(&id)
	return id, err
}

func (repository *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := repository.DB.QueryRowContext(ctx,
		`SELECT id, xid, name, email, password, is_active, email_verified_at, created_at, updated_at
         FROM users WHERE email=$1 AND deleted_at IS NULL LIMIT 1`, email,
	)
	user, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repository *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	row := repository.DB.QueryRowContext(ctx,
		`SELECT id, xid, name, email, password, is_active, email_verified_at, created_at, updated_at
         FROM users WHERE id=$1 AND deleted_at IS NULL LIMIT 1`, id,
	)
	user, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repository *UserRepository) FindByXID(ctx context.Context, xid string) (*domain.User, error) {
	row := repository.DB.QueryRowContext(ctx,
		`SELECT id, xid, name, email, password, is_active, email_verified_at, created_at, updated_at
         FROM users WHERE xid=$1 AND deleted_at IS NULL LIMIT 1`, xid,
	)
	user, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repository *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	rows, err := repository.DB.QueryContext(ctx,
		`SELECT id, xid, name, email, password, is_active, email_verified_at, created_at, updated_at
         FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (repository *UserRepository) Update(ctx context.Context, xid, name, email string, passwordHash *string, isActive bool) (*domain.User, error) {
	var row rowScanner
	if passwordHash != nil {
		row = repository.DB.QueryRowContext(ctx,
			`UPDATE users
             SET name=$1, email=$2, password=$3, is_active=$4, updated_at=NOW()
             WHERE xid=$5 AND deleted_at IS NULL
             RETURNING id, xid, name, email, password, is_active, email_verified_at, created_at, updated_at`,
			name, email, *passwordHash, isActive, xid,
		)
	} else {
		row = repository.DB.QueryRowContext(ctx,
			`UPDATE users
             SET name=$1, email=$2, is_active=$3, updated_at=NOW()
             WHERE xid=$4 AND deleted_at IS NULL
             RETURNING id, xid, name, email, password, is_active, email_verified_at, created_at, updated_at`,
			name, email, isActive, xid,
		)
	}
	user, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repository *UserRepository) Delete(ctx context.Context, xid string) error {
	result, err := repository.DB.ExecContext(ctx,
		`UPDATE users
         SET deleted_at = NOW(), updated_at = NOW(), is_active = FALSE
         WHERE xid = $1 AND deleted_at IS NULL`, xid,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanUser(row rowScanner) (*domain.User, error) {
	var (
		user            domain.User
		emailVerifiedAt sql.NullTime
	)

	err := row.Scan(&user.ID, &user.XID, &user.Name, &user.Email, &user.Password, &user.IsActive, &emailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if emailVerifiedAt.Valid {
		t := emailVerifiedAt.Time
		user.EmailVerifiedAt = &t
	}
	return &user, nil
}
