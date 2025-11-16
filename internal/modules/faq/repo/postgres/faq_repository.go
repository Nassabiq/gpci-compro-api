package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/faq/domain"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type FAQRepository struct {
	DB *sql.DB
}

func NewFAQRepository(db *sql.DB) *FAQRepository {
	return &FAQRepository{DB: db}
}

const (
	baseSelectFAQ = `
SELECT id, question, answer, created_at, updated_at
FROM public.faqs
WHERE deleted_at IS NULL
`
)

func (repository *FAQRepository) ListFAQs(ctx context.Context, filter domain.FAQFilter) ([]domain.FAQ, int, error) {
	var total int
	if err := repository.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM public.faqs WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.PageSize
	offset := 0
	if filter.Page > 0 && filter.PageSize > 0 {
		offset = (filter.Page - 1) * filter.PageSize
	}
	if limit > 0 && offset >= total {
		return []domain.FAQ{}, total, nil
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseSelectFAQ)
	queryBuilder.WriteString("ORDER BY created_at DESC, id DESC")

	var args []any
	if limit > 0 {
		args = append(args, limit)
		queryBuilder.WriteString(fmt.Sprintf("\nLIMIT $%d", len(args)))
		if offset > 0 {
			args = append(args, offset)
			queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
		}
	}

	rows, err := repository.DB.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var faqs []domain.FAQ
	for rows.Next() {
		faq, scanErr := scanFAQ(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		faqs = append(faqs, faq)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return faqs, total, nil
}

func (repository *FAQRepository) GetFAQByID(ctx context.Context, id int64) (*domain.FAQ, error) {
	row := repository.DB.QueryRowContext(ctx, baseSelectFAQ+`AND id = $1 LIMIT 1`, id)
	faq, err := scanFAQ(row)
	if err != nil {
		return nil, err
	}
	return &faq, nil
}

func (repository *FAQRepository) mustGetFAQByID(ctx context.Context, id int64) (domain.FAQ, error) {
	faq, err := repository.GetFAQByID(ctx, id)
	if err != nil {
		return domain.FAQ{}, err
	}
	return *faq, nil
}

func (repository *FAQRepository) CreateFAQ(ctx context.Context, payload domain.FAQPayload) (domain.FAQ, error) {
	var id int64
	if err := repository.DB.QueryRowContext(ctx,
		`INSERT INTO public.faqs(question, answer) VALUES ($1, $2) RETURNING id`,
		payload.Question, payload.Answer,
	).Scan(&id); err != nil {
		return domain.FAQ{}, err
	}
	return repository.mustGetFAQByID(ctx, id)
}

func (repository *FAQRepository) UpdateFAQ(ctx context.Context, id int64, payload domain.FAQPayload) (domain.FAQ, error) {
	var updatedID int64
	err := repository.DB.QueryRowContext(ctx,
		`UPDATE public.faqs
         SET question = $1,
             answer = $2,
             updated_at = NOW()
         WHERE id = $3 AND deleted_at IS NULL
         RETURNING id`,
		payload.Question, payload.Answer, id,
	).Scan(&updatedID)
	if err != nil {
		return domain.FAQ{}, err
	}
	return repository.mustGetFAQByID(ctx, updatedID)
}

func (repository *FAQRepository) DeleteFAQ(ctx context.Context, id int64) error {
	result, err := repository.DB.ExecContext(ctx,
		`UPDATE public.faqs
         SET deleted_at = NOW(), updated_at = NOW()
         WHERE id = $1 AND deleted_at IS NULL`,
		id,
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

func scanFAQ(row rowScanner) (domain.FAQ, error) {
	var (
		faq       domain.FAQ
		createdAt time.Time
		updatedAt time.Time
	)

	if err := row.Scan(&faq.ID, &faq.Question, &faq.Answer, &createdAt, &updatedAt); err != nil {
		return domain.FAQ{}, err
	}

	faq.CreatedAt = createdAt
	faq.UpdatedAt = updatedAt
	return faq, nil
}
