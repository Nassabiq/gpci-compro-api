package postgres

import (
	"context"
	"database/sql"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/catalog/domain"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) ListPrograms(ctx context.Context) ([]domain.Program, error) {
	const query = `
SELECT id, code, name
FROM public.lkp_product_program
ORDER BY id`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []domain.Program
	for rows.Next() {
		var program domain.Program
		if err := rows.Scan(&program.ID, &program.Code, &program.Name); err != nil {
			return nil, err
		}
		programs = append(programs, program)
	}
	return programs, rows.Err()
}

func (r *Repository) CreateProgram(ctx context.Context, payload domain.ProgramPayload) (domain.Program, error) {
	const query = `
INSERT INTO public.lkp_product_program (id, code, name)
VALUES ($1, $2, $3)
RETURNING id, code, name`

	var program domain.Program
	err := r.DB.QueryRowContext(ctx, query, payload.ID, payload.Code, payload.Name).Scan(&program.ID, &program.Code, &program.Name)
	return program, err
}

func (r *Repository) UpdateProgram(ctx context.Context, id int16, payload domain.ProgramPayload) (domain.Program, error) {
	const query = `
UPDATE public.lkp_product_program
SET code = $1,
    name = $2
WHERE id = $3
RETURNING id, code, name`

	var program domain.Program
	err := r.DB.QueryRowContext(ctx, query, payload.Code, payload.Name, id).Scan(&program.ID, &program.Code, &program.Name)
	return program, err
}

func (r *Repository) DeleteProgram(ctx context.Context, id int16) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM public.lkp_product_program WHERE id = $1`, id)
	return err
}

func (r *Repository) ListStatuses(ctx context.Context) ([]domain.CertificationStatus, error) {
	const query = `
SELECT id, code, name
FROM public.lkp_cert_status
ORDER BY id`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []domain.CertificationStatus
	for rows.Next() {
		var status domain.CertificationStatus
		if err := rows.Scan(&status.ID, &status.Code, &status.Name); err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	return statuses, rows.Err()
}

func (r *Repository) CreateStatus(ctx context.Context, payload domain.StatusPayload) (domain.CertificationStatus, error) {
	const query = `
INSERT INTO public.lkp_cert_status (id, code, name)
VALUES ($1, $2, $3)
RETURNING id, code, name`

	var status domain.CertificationStatus
	err := r.DB.QueryRowContext(ctx, query, payload.ID, payload.Code, payload.Name).Scan(&status.ID, &status.Code, &status.Name)
	return status, err
}

func (r *Repository) UpdateStatus(ctx context.Context, id int16, payload domain.StatusPayload) (domain.CertificationStatus, error) {
	const query = `
UPDATE public.lkp_cert_status
SET code = $1,
    name = $2
WHERE id = $3
RETURNING id, code, name`

	var status domain.CertificationStatus
	err := r.DB.QueryRowContext(ctx, query, payload.Code, payload.Name, id).Scan(&status.ID, &status.Code, &status.Name)
	return status, err
}

func (r *Repository) DeleteStatus(ctx context.Context, id int16) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM public.lkp_cert_status WHERE id = $1`, id)
	return err
}
