package postgres

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/brand/domain"
)

func (r *BrandRepository) ListBrandCategories(ctx context.Context, filter domain.BrandCategoryFilter) ([]domain.BrandCategory, int, error) {
	const countQuery = `SELECT COUNT(*) FROM public.brand_categories`

	var total int
	if err := r.DB.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.PageSize
	offset := 0
	if filter.Page > 0 && filter.PageSize > 0 {
		offset = (filter.Page - 1) * filter.PageSize
	}
	if limit > 0 && offset >= total {
		return []domain.BrandCategory{}, total, nil
	}

	query := `
		SELECT id, name, slug
		FROM public.brand_categories
		ORDER BY name`

	var args []any
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET $2"
			args = append(args, offset)
		}
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var categories []domain.BrandCategory
	for rows.Next() {
		var category domain.BrandCategory
		if err := rows.Scan(&category.ID, &category.Name, &category.Slug); err != nil {
			return nil, 0, err
		}
		categories = append(categories, category)
	}
	return categories, total, rows.Err()
}

func (r *BrandRepository) CreateBrandCategory(ctx context.Context, payload domain.BrandCategoryPayload) (domain.BrandCategory, error) {
	const query = `
		INSERT INTO public.brand_categories (name, slug)
		VALUES ($1, $2)
		RETURNING id, name, slug`

	var category domain.BrandCategory
	err := r.DB.QueryRowContext(ctx, query, payload.Name, payload.Slug).Scan(&category.ID, &category.Name, &category.Slug)
	return category, err
}

func (r *BrandRepository) UpdateBrandCategory(ctx context.Context, id int64, payload domain.BrandCategoryPayload) (domain.BrandCategory, error) {
	const query = `
		UPDATE public.brand_categories
		SET name = $1,
			slug = $2,
			updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, slug`

	var category domain.BrandCategory
	err := r.DB.QueryRowContext(ctx, query, payload.Name, payload.Slug, id).Scan(&category.ID, &category.Name, &category.Slug)
	return category, err
}

func (r *BrandRepository) DeleteBrandCategory(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM public.brand_categories WHERE id = $1`, id)
	return err
}
