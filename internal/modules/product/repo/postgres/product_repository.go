package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/domain"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (repository *ProductRepository) ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int, error) {
	whereClause, args := buildProductWhereClause(filter)
	total, err := repository.countProducts(ctx, whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.PageSize
	offset := 0
	if filter.Page > 0 && filter.PageSize > 0 {
		offset = (filter.Page - 1) * filter.PageSize
	}
	if limit > 0 && offset >= total {
		return []domain.Product{}, total, nil
	}

	products, err := repository.queryProducts(ctx, whereClause, "ORDER BY p.created_at DESC, p.id DESC", limit, offset, args...)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

const (
	baseProductSelect = `
SELECT
    p.id,
    p.name,
    p.slug,
    p.features,
    p.reason,
    p.tshp,
    p.images,
    p.is_active,
    p.created_at,
    p.updated_at
`
	baseProductFrom = `
FROM public.products p
JOIN public.lkp_product_program pp ON pp.id = p.program_id
JOIN public.brands b ON b.id = p.brand_id
JOIN public.brand_categories bc ON bc.id = b.brand_category_id
JOIN public.companies c ON c.id = p.company_id
`
	baseProductCountQuery = `
SELECT COUNT(*)
` + baseProductFrom
)

func (repository *ProductRepository) queryProducts(ctx context.Context, whereClause, orderClause string, limit, offset int, args ...any) ([]domain.Product, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseProductSelect)
	queryBuilder.WriteString(baseProductFrom)
	if strings.TrimSpace(whereClause) != "" {
		queryBuilder.WriteString("WHERE ")
		queryBuilder.WriteString(whereClause)
		queryBuilder.WriteRune('\n')
	}
	if strings.TrimSpace(orderClause) != "" {
		queryBuilder.WriteRune('\n')
		queryBuilder.WriteString(orderClause)
	}

	queryArgs := make([]any, len(args))
	copy(queryArgs, args)

	if limit > 0 {
		queryArgs = append(queryArgs, limit)
		queryBuilder.WriteString(fmt.Sprintf("\nLIMIT $%d", len(queryArgs)))
		if offset > 0 {
			queryArgs = append(queryArgs, offset)
			queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", len(queryArgs)))
		}
	}

	rows, err := repository.DB.QueryContext(ctx, queryBuilder.String(), queryArgs...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		product, scanErr := scanProduct(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (repository *ProductRepository) GetProductBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	whereClause := "p.slug = $1"
	products, err := repository.queryProducts(ctx, whereClause, "ORDER BY p.id", 1, 0, slug)
	if err != nil {
		return nil, err
	}

	for _, product := range products {
		return &product, nil
	}
	return nil, sql.ErrNoRows
}

func (repository *ProductRepository) mustGetProductBySlug(ctx context.Context, slug string) (domain.Product, error) {
	product, err := repository.GetProductBySlug(ctx, slug)
	if err != nil {
		return domain.Product{}, err
	}
	return *product, nil
}

func (repository *ProductRepository) CreateProduct(ctx context.Context, payload domain.ProductPayload) (domain.Product, error) {
	tshpJSON, imagesJSON, isActive, err := prepareProductJSON(payload)
	if err != nil {
		return domain.Product{}, err
	}

	const query = `
		INSERT INTO public.products (
			company_id,
			brand_id,
			program_id,
			name,
			slug,
			features,
			reason,
			tshp,
			images,
			is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9::jsonb,$10)
		RETURNING slug`

	var slug string
	err = repository.DB.QueryRowContext(
		ctx,
		query,
		payload.CompanyID,
		payload.BrandID,
		payload.ProgramID,
		payload.Name,
		payload.Slug,
		payload.Features,
		payload.Reason,
		tshpJSON,
		imagesJSON,
		isActive,
	).Scan(&slug)

	if err != nil {
		return domain.Product{}, err
	}

	return repository.mustGetProductBySlug(ctx, slug)
}

func (repository *ProductRepository) UpdateProduct(ctx context.Context, slug string, payload domain.ProductPayload) (domain.Product, error) {
	tshpJSON, imagesJSON, isActive, err := prepareProductJSON(payload)
	if err != nil {
		return domain.Product{}, err
	}

	const query = `
		UPDATE public.products
		SET company_id = $1,
			brand_id = $2,
			program_id = $3,
			name = $4,
			slug = $5,
			features = $6,
			reason = $7,
			tshp = $8::jsonb,
			images = $9::jsonb,
			is_active = $10,
			updated_at = NOW()
		WHERE slug = $11
		RETURNING slug`

	var newSlug string
	err = repository.DB.QueryRowContext(
		ctx,
		query,
		payload.CompanyID,
		payload.BrandID,
		payload.ProgramID,
		payload.Name,
		payload.Slug,
		payload.Features,
		payload.Reason,
		tshpJSON,
		imagesJSON,
		isActive,
		slug,
	).Scan(&newSlug)
	if err != nil {
		return domain.Product{}, err
	}
	return repository.mustGetProductBySlug(ctx, newSlug)
}

func (repository *ProductRepository) DeleteProduct(ctx context.Context, slug string) error {
	_, err := repository.DB.ExecContext(ctx, `DELETE FROM public.products WHERE slug = $1`, slug)
	return err
}

func buildProductWhereClause(filter domain.ProductFilter) (string, []any) {
	var (
		clauses []string
		args    []any
		pos     = 1
	)

	appendClause := func(condition string, value any) {
		clauses = append(clauses, fmt.Sprintf(condition, pos))
		args = append(args, value)
		pos++
	}

	if filter.ProgramCode != "" {
		appendClause("pp.code = $%d", filter.ProgramCode)
	}
	if filter.BrandSlug != "" {
		appendClause("b.slug = $%d", filter.BrandSlug)
	}
	if filter.CategorySlug != "" {
		appendClause("bc.slug = $%d", filter.CategorySlug)
	}
	if filter.Search != "" {
		condition := "(p.name ILIKE '%%' || $%[1]d || '%%' OR c.name ILIKE '%%' || $%[1]d || '%%')"
		clauses = append(clauses, fmt.Sprintf(condition, pos))
		args = append(args, filter.Search)
		pos++
	}
	if filter.IsActiveOnly {
		appendClause("p.is_active = $%d", true)
	}

	return strings.Join(clauses, " AND "), args
}

func (repository *ProductRepository) countProducts(ctx context.Context, whereClause string, args ...any) (int, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseProductCountQuery)
	if strings.TrimSpace(whereClause) != "" {
		queryBuilder.WriteString("WHERE ")
		queryBuilder.WriteString(whereClause)
	}

	var total int
	if err := repository.DB.QueryRowContext(ctx, queryBuilder.String(), args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func scanProduct(rows *sql.Rows) (domain.Product, error) {
	var (
		product   domain.Product
		features  sql.NullString
		reason    sql.NullString
		tshpRaw   []byte
		imagesRaw []byte
		createdAt time.Time
		updatedAt time.Time
	)

	if err := rows.Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&features,
		&reason,
		&tshpRaw,
		&imagesRaw,
		&product.IsActive,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.Product{}, err
	}

	if len(tshpRaw) == 0 {
		product.TSHP = map[string]any{}
	} else if err := json.Unmarshal(tshpRaw, &product.TSHP); err != nil {
		return domain.Product{}, err
	}
	if len(imagesRaw) == 0 {
		product.Images = []string{}
	} else if err := json.Unmarshal(imagesRaw, &product.Images); err != nil {
		return domain.Product{}, err
	}

	product.Features = features.String
	product.Reason = reason.String
	product.CreatedAt = createdAt
	product.UpdatedAt = updatedAt

	return product, nil
}

func prepareProductJSON(payload domain.ProductPayload) ([]byte, []byte, bool, error) {
	tshp := payload.TSHP
	if tshp == nil {
		tshp = map[string]any{}
	}
	images := payload.Images
	if images == nil {
		images = []string{}
	}

	tshpJSON, err := json.Marshal(tshp)
	if err != nil {
		return nil, nil, false, err
	}
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return nil, nil, false, err
	}

	isActive := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}

	return tshpJSON, imagesJSON, isActive, nil
}
