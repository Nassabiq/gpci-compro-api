package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/domain"
)

type ProductCertificationRepository struct {
	DB *sql.DB
}

func NewProductCertificationRepository(db *sql.DB) *ProductCertificationRepository {
	return &ProductCertificationRepository{DB: db}
}

func (repository *ProductCertificationRepository) ListProductCertifications(ctx context.Context, productSlug string, filter domain.ProductCertificationFilter) ([]domain.ProductCertification, int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM public.product_has_certification pc
		JOIN public.products p ON p.id = pc.product_id
		WHERE p.slug = $1`

	var total int
	if err := repository.DB.QueryRowContext(ctx, countQuery, productSlug).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.PageSize
	offset := 0
	if filter.Page > 0 && filter.PageSize > 0 {
		offset = (filter.Page - 1) * filter.PageSize
	}
	if limit > 0 && offset >= total {
		return []domain.ProductCertification{}, total, nil
	}

	query := `
		SELECT
			pc.certificate_no,
			pc.issue_date,
			pc.expiry_date,
			pc.document_file,
			pc.meta_json,
			pc.updated_at,
			c.id,
			c.name,
			c.image,
			pp.id,
			pp.code,
			pp.name,
			cs.id,
			cs.code,
			cs.name
		FROM public.product_has_certification pc
		JOIN public.products p ON p.id = pc.product_id
		JOIN public.certifications c ON c.id = pc.certification_id
		JOIN public.lkp_product_program pp ON pp.id = c.program_id
		LEFT JOIN public.lkp_cert_status cs ON cs.id = pc.status_id
		WHERE p.slug = $1
		ORDER BY c.name`

	args := []any{productSlug}
	if limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf("\nLIMIT $%d", len(args))
		if offset > 0 {
			args = append(args, offset)
			query += fmt.Sprintf(" OFFSET $%d", len(args))
		}
	}

	rows, err := repository.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	var result []domain.ProductCertification
	for rows.Next() {
		cert, scanErr := scanProductCertification(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		result = append(result, cert)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (repository *ProductCertificationRepository) CreateProductCertification(ctx context.Context, productSlug string, payload domain.ProductCertificationPayload) (domain.ProductCertification, error) {
	productID, err := repository.productIDBySlug(ctx, productSlug)
	if err != nil {
		return domain.ProductCertification{}, err
	}

	meta := payload.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return domain.ProductCertification{}, err
	}

	var certificateNo any
	if payload.CertificateNo != nil {
		certificateNo = *payload.CertificateNo
	}
	var documentFile any
	if payload.DocumentFile != nil {
		documentFile = *payload.DocumentFile
	}

	const query = `
		INSERT INTO public.product_has_certification (
			product_id,
			certification_id,
			certificate_no,
			issue_date,
			expiry_date,
			status_id,
			document_file,
			meta_json
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)`

	_, err = repository.DB.ExecContext(
		ctx,
		query,
		productID,
		payload.CertificationID,
		certificateNo,
		payload.IssueDate,
		payload.ExpiryDate,
		payload.StatusID,
		documentFile,
		metaJSON,
	)
	if err != nil {
		return domain.ProductCertification{}, err
	}

	return repository.getProductCertification(ctx, productID, payload.CertificationID)
}

func (repository *ProductCertificationRepository) UpdateProductCertification(
	ctx context.Context,
	productSlug string,
	certificationID int64,
	payload domain.ProductCertificationPayload,
) (domain.ProductCertification, error) {
	// BEGIN FUNCTION
	productID, err := repository.productIDBySlug(ctx, productSlug)
	if err != nil {
		return domain.ProductCertification{}, err
	}

	meta := payload.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return domain.ProductCertification{}, err
	}

	var certificateNo any
	if payload.CertificateNo != nil {
		certificateNo = *payload.CertificateNo
	}
	var documentFile any
	if payload.DocumentFile != nil {
		documentFile = *payload.DocumentFile
	}

	const query = `
		UPDATE public.product_has_certification
		SET certificate_no = $3,
			issue_date = $4,
			expiry_date = $5,
			status_id = $6,
			document_file = $7,
			meta_json = $8::jsonb,
			updated_at = NOW()
		WHERE product_id = $1 AND certification_id = $2`

	if _, err := repository.DB.ExecContext(
		ctx,
		query,
		productID,
		certificationID,
		certificateNo,
		payload.IssueDate,
		payload.ExpiryDate,
		payload.StatusID,
		documentFile,
		metaJSON,
	); err != nil {
		return domain.ProductCertification{}, err
	}

	return repository.getProductCertification(ctx, productID, certificationID)
}

func (repository *ProductCertificationRepository) DeleteProductCertification(ctx context.Context, productSlug string, certificationID int64) error {
	productID, err := repository.productIDBySlug(ctx, productSlug)
	if err != nil {
		return err
	}
	_, err = repository.DB.ExecContext(ctx, `DELETE FROM public.product_has_certification WHERE product_id = $1 AND certification_id = $2`, productID, certificationID)
	return err
}

func (repository *ProductCertificationRepository) productIDBySlug(ctx context.Context, slug string) (int64, error) {
	const query = `SELECT id FROM public.products WHERE slug = $1`
	var id int64
	err := repository.DB.QueryRowContext(ctx, query, slug).Scan(&id)
	return id, err
}

func (repository *ProductCertificationRepository) getProductCertification(ctx context.Context, productID, certificationID int64) (domain.ProductCertification, error) {
	const query = `
		SELECT
			pc.certificate_no,
			pc.issue_date,
			pc.expiry_date,
			pc.document_file,
			pc.meta_json,
			pc.updated_at,
			c.id,
			c.name,
			c.image,
			pp.id,
			pp.code,
			pp.name,
			cs.id,
			cs.code,
			cs.name
		FROM public.product_has_certification pc
		JOIN public.certifications c ON c.id = pc.certification_id
		JOIN public.lkp_product_program pp ON pp.id = c.program_id
		LEFT JOIN public.lkp_cert_status cs ON cs.id = pc.status_id
		WHERE pc.product_id = $1 AND pc.certification_id = $2`

	row := repository.DB.QueryRowContext(ctx, query, productID, certificationID)

	var (
		cert          domain.ProductCertification
		metaRaw       []byte
		issue         sql.NullTime
		expiry        sql.NullTime
		statusID      sql.NullInt64
		statusCode    sql.NullString
		statusName    sql.NullString
		documentFile  sql.NullString
		certificateNo sql.NullString
		certImage     sql.NullString
	)

	if err := row.Scan(
		&certificateNo,
		&issue,
		&expiry,
		&documentFile,
		&metaRaw,
		&cert.UpdatedAt,
		&cert.Certification.ID,
		&cert.Certification.Name,
		&certImage,
		&cert.Certification.Program.ID,
		&cert.Certification.Program.Code,
		&cert.Certification.Program.Name,
		&statusID,
		&statusCode,
		&statusName,
	); err != nil {
		return domain.ProductCertification{}, err
	}

	if certificateNo.Valid {
		cert.CertificateNo = certificateNo.String
	}
	if issue.Valid {
		t := issue.Time
		cert.IssueDate = &t
	}
	if expiry.Valid {
		t := expiry.Time
		cert.ExpiryDate = &t
	}
	if documentFile.Valid {
		cert.DocumentFile = documentFile.String
	}
	if certImage.Valid {
		cert.Certification.Image = certImage.String
	}
	if len(metaRaw) > 0 {
		if err := json.Unmarshal(metaRaw, &cert.Meta); err != nil {
			return domain.ProductCertification{}, err
		}
	} else {
		cert.Meta = map[string]any{}
	}
	if statusID.Valid && statusCode.Valid && statusName.Valid {
		status := domain.CertificationStatus{
			ID:   int16(statusID.Int64),
			Code: statusCode.String,
			Name: statusName.String,
		}
		cert.Status = &status
	}
	return cert, nil
}

func scanProductCertification(rows *sql.Rows) (domain.ProductCertification, error) {
	var (
		cert          domain.ProductCertification
		metaRaw       []byte
		issue         sql.NullTime
		expiry        sql.NullTime
		statusID      sql.NullInt64
		statusCode    sql.NullString
		statusName    sql.NullString
		documentFile  sql.NullString
		certificateNo sql.NullString
		certImage     sql.NullString
	)

	if err := rows.Scan(
		&certificateNo,
		&issue,
		&expiry,
		&documentFile,
		&metaRaw,
		&cert.UpdatedAt,
		&cert.Certification.ID,
		&cert.Certification.Name,
		&certImage,
		&cert.Certification.Program.ID,
		&cert.Certification.Program.Code,
		&cert.Certification.Program.Name,
		&statusID,
		&statusCode,
		&statusName,
	); err != nil {
		return domain.ProductCertification{}, err
	}

	if certificateNo.Valid {
		cert.CertificateNo = certificateNo.String
	}
	if issue.Valid {
		t := issue.Time
		cert.IssueDate = &t
	}
	if expiry.Valid {
		t := expiry.Time
		cert.ExpiryDate = &t
	}
	if documentFile.Valid {
		cert.DocumentFile = documentFile.String
	}
	if certImage.Valid {
		cert.Certification.Image = certImage.String
	}
	if len(metaRaw) > 0 {
		if err := json.Unmarshal(metaRaw, &cert.Meta); err != nil {
			return domain.ProductCertification{}, err
		}
	} else {
		cert.Meta = map[string]any{}
	}
	if statusID.Valid && statusCode.Valid && statusName.Valid {
		status := domain.CertificationStatus{
			ID:   int16(statusID.Int64),
			Code: statusCode.String,
			Name: statusName.String,
		}
		cert.Status = &status
	}
	return cert, nil
}
