package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/domain"
)

type programRowScanner interface {
	Scan(dest ...any) error
}

const programCertificateBaseSelect = `
SELECT
	pc.id,
	pc.created_at,
	pc.updated_at,
	pc.certificate_no,
	pc.issue_date,
	pc.expiry_date,
	pc.document_file,
	pc.meta_json,
	p.id,
	p.name,
	p.slug,
	b.id,
	b.name,
	b.slug,
	bc.id,
	bc.name,
	bc.slug,
	co.id,
	co.name,
	co.slug,
	c.id,
	c.name,
	c.image,
	prog.id,
	prog.code,
	prog.name,
	cs.id,
	cs.code,
	cs.name
FROM public.product_has_certification pc
JOIN public.products p ON p.id = pc.product_id
JOIN public.certifications c ON c.id = pc.certification_id
JOIN public.lkp_product_program prog ON prog.id = c.program_id
JOIN public.brands b ON b.id = p.brand_id
JOIN public.brand_categories bc ON bc.id = b.brand_category_id
JOIN public.companies co ON co.id = p.company_id
LEFT JOIN public.lkp_cert_status cs ON cs.id = pc.status_id
WHERE prog.code = $1
`

func (repository *ProductCertificationRepository) ListProgramCertificates(ctx context.Context, programCode string, filter domain.ProgramCertificateFilter) ([]domain.ProgramCertificate, int, error) {
	countBuilder := strings.Builder{}
	countBuilder.WriteString(`
		SELECT COUNT(*)
		FROM public.product_has_certification pc
		JOIN public.products p ON p.id = pc.product_id
		JOIN public.certifications c ON c.id = pc.certification_id
		JOIN public.lkp_product_program prog ON prog.id = c.program_id
		JOIN public.brands b ON b.id = p.brand_id
		JOIN public.brand_categories bc ON bc.id = b.brand_category_id
		JOIN public.companies co ON co.id = p.company_id
		WHERE prog.code = $1`)

	countArgs := []any{programCode}
	if filter.Search != "" {
		countArgs = append(countArgs, fmt.Sprintf("%%%s%%", filter.Search))
		countBuilder.WriteString(fmt.Sprintf(" AND (p.name ILIKE $%d OR coalesce(pc.certificate_no,'') ILIKE $%d OR co.name ILIKE $%d OR c.name ILIKE $%d)", len(countArgs), len(countArgs), len(countArgs), len(countArgs)))
	}

	var total int
	if err := repository.DB.QueryRowContext(ctx, countBuilder.String(), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.PageSize
	offset := 0
	if filter.Page > 0 && filter.PageSize > 0 {
		offset = (filter.Page - 1) * filter.PageSize
	}
	if limit > 0 && offset >= total {
		return []domain.ProgramCertificate{}, total, nil
	}

	var (
		args    = []any{programCode}
		builder strings.Builder
	)

	builder.WriteString(programCertificateBaseSelect)

	if filter.Search != "" {
		args = append(args, fmt.Sprintf("%%%s%%", filter.Search))
		builder.WriteString(fmt.Sprintf(" AND (p.name ILIKE $%d OR co.name ILIKE $%d OR c.name ILIKE $%d OR coalesce(pc.certificate_no,'') ILIKE $%d)", len(args), len(args), len(args), len(args)))
	}

	builder.WriteString("\nORDER BY pc.updated_at DESC, pc.id DESC")

	if limit > 0 {
		args = append(args, limit)
		builder.WriteString(fmt.Sprintf("\nLIMIT $%d", len(args)))
		if offset > 0 {
			args = append(args, offset)
			builder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
		}
	}

	rows, err := repository.DB.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var certificates []domain.ProgramCertificate
	for rows.Next() {
		record, scanErr := scanProgramCertificate(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		certificates = append(certificates, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return certificates, total, nil
}

func (repository *ProductCertificationRepository) GetProgramCertificate(ctx context.Context, programCode, productSlug string, certificationID int64) (*domain.ProgramCertificate, error) {
	builder := strings.Builder{}
	builder.WriteString(programCertificateBaseSelect)
	builder.WriteString(" AND p.slug = $2 AND c.id = $3")

	row := repository.DB.QueryRowContext(ctx, builder.String(), programCode, productSlug, certificationID)
	record, err := scanProgramCertificate(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &record, nil
}

func (repository *ProductCertificationRepository) GetProductProgramCode(ctx context.Context, productSlug string) (string, error) {
	const query = `
		SELECT prog.code
		FROM public.products p
		JOIN public.lkp_product_program prog ON prog.id = p.program_id
		WHERE p.slug = $1`

	var code string
	if err := repository.DB.QueryRowContext(ctx, query, productSlug).Scan(&code); err != nil {
		return "", err
	}
	return code, nil
}

func (repository *ProductCertificationRepository) GetCertificationProgramCode(ctx context.Context, certificationID int64) (string, error) {
	const query = `
		SELECT prog.code
		FROM public.certifications c
		JOIN public.lkp_product_program prog ON prog.id = c.program_id
		WHERE c.id = $1`

	var code string
	if err := repository.DB.QueryRowContext(ctx, query, certificationID).Scan(&code); err != nil {
		return "", err
	}
	return code, nil
}

func scanProgramCertificate(row programRowScanner) (domain.ProgramCertificate, error) {
	var (
		record        domain.ProgramCertificate
		metaRaw       []byte
		issue         sql.NullTime
		expiry        sql.NullTime
		documentFile  sql.NullString
		certificateNo sql.NullString
		certImage     sql.NullString
		statusID      sql.NullInt64
		statusCode    sql.NullString
		statusName    sql.NullString
	)

	if err := row.Scan(
		&record.ID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&certificateNo,
		&issue,
		&expiry,
		&documentFile,
		&metaRaw,
		&record.Product.ID,
		&record.Product.Name,
		&record.Product.Slug,
		&record.Brand.ID,
		&record.Brand.Name,
		&record.Brand.Slug,
		&record.Brand.Category.ID,
		&record.Brand.Category.Name,
		&record.Brand.Category.Slug,
		&record.Company.ID,
		&record.Company.Name,
		&record.Company.Slug,
		&record.Certification.ID,
		&record.Certification.Name,
		&certImage,
		&record.Program.ID,
		&record.Program.Code,
		&record.Program.Name,
		&statusID,
		&statusCode,
		&statusName,
	); err != nil {
		return domain.ProgramCertificate{}, err
	}

	if certificateNo.Valid {
		record.CertificateNo = certificateNo.String
	}
	if issue.Valid {
		t := issue.Time
		record.IssueDate = &t
	}
	if expiry.Valid {
		t := expiry.Time
		record.ExpiryDate = &t
	}
	if documentFile.Valid {
		record.DocumentFile = documentFile.String
	}
	if certImage.Valid {
		record.Certification.Image = certImage.String
	}
	if len(metaRaw) > 0 {
		if err := json.Unmarshal(metaRaw, &record.Meta); err != nil {
			return domain.ProgramCertificate{}, err
		}
	} else {
		record.Meta = map[string]any{}
	}
	if statusID.Valid && statusCode.Valid && statusName.Valid {
		status := domain.CertificationStatus{
			ID:   int16(statusID.Int64),
			Code: statusCode.String,
			Name: statusName.String,
		}
		record.Status = &status
	}
	return record, nil
}
