package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/brand/domain"
)

type BrandRepository struct {
	DB *sql.DB
}

func NewBrandRepository(db *sql.DB) *BrandRepository {
	return &BrandRepository{DB: db}
}

func (r *BrandRepository) ListBrands(ctx context.Context, filter domain.BrandFilter) ([]domain.Brand, int, error) {
	categorySlug := strings.TrimSpace(filter.CategorySlug)

	const countQuery = `
SELECT COUNT(*)
FROM public.brands b
JOIN public.brand_categories bc ON bc.id = b.brand_category_id
WHERE ($1 = '' OR bc.slug = $1)`

	var total int
	if err := r.DB.QueryRowContext(ctx, countQuery, categorySlug).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.PageSize
	offset := 0
	if filter.Page > 0 && filter.PageSize > 0 {
		offset = (filter.Page - 1) * filter.PageSize
	}
	if limit > 0 && offset >= total {
		return []domain.Brand{}, total, nil
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
SELECT
    b.id,
    b.name,
    b.slug,
    bc.id,
    bc.name,
    bc.slug
FROM public.brands b
JOIN public.brand_categories bc ON bc.id = b.brand_category_id
WHERE ($1 = '' OR bc.slug = $1)
ORDER BY b.name`)

	args := []any{categorySlug}
	if limit > 0 {
		args = append(args, limit)
		queryBuilder.WriteString(fmt.Sprintf("\nLIMIT $%d", len(args)))
		if offset > 0 {
			args = append(args, offset)
			queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
		}
	}

	rows, err := r.DB.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var brands []domain.Brand
	for rows.Next() {
		var brand domain.Brand
		if err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Slug,
			&brand.Category.ID,
			&brand.Category.Name,
			&brand.Category.Slug,
		); err != nil {
			return nil, 0, err
		}
		brands = append(brands, brand)
	}
	return brands, total, rows.Err()
}

func (r *BrandRepository) CreateBrand(ctx context.Context, payload domain.BrandPayload) (domain.Brand, error) {
	const query = `
INSERT INTO public.brands (brand_category_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id`

	var brandID int64
	if err := r.DB.QueryRowContext(ctx, query, payload.CategoryID, payload.Name, payload.Slug).Scan(&brandID); err != nil {
		return domain.Brand{}, err
	}
	return r.getBrandByID(ctx, brandID)
}

func (r *BrandRepository) UpdateBrand(ctx context.Context, id int64, payload domain.BrandPayload) (domain.Brand, error) {
	const query = `
UPDATE public.brands
SET brand_category_id = $1,
    name = $2,
    slug = $3,
    updated_at = NOW()
WHERE id = $4
RETURNING id`

	var brandID int64
	if err := r.DB.QueryRowContext(ctx, query, payload.CategoryID, payload.Name, payload.Slug, id).Scan(&brandID); err != nil {
		return domain.Brand{}, err
	}
	return r.getBrandByID(ctx, brandID)
}

func (r *BrandRepository) DeleteBrand(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM public.brands WHERE id = $1`, id)
	return err
}

func (r *BrandRepository) getBrandByID(ctx context.Context, id int64) (domain.Brand, error) {
	const query = `
SELECT
    b.id,
    b.name,
    b.slug,
    bc.id,
    bc.name,
    bc.slug
FROM public.brands b
JOIN public.brand_categories bc ON bc.id = b.brand_category_id
WHERE b.id = $1`

	var brand domain.Brand
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.Category.ID,
		&brand.Category.Name,
		&brand.Category.Slug,
	)
	return brand, err
}
