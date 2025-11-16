package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/domain"
)

var (
	ErrProductNotFound       = errors.New("product_not_found")
	ErrCertificationNotFound = errors.New("certification_not_found")
	ErrProgramMismatch       = errors.New("program_mismatch")
	ErrCertificateNotFound   = errors.New("program_certificate_not_found")
)

type ProgramCertificateRepository interface {
	GetProductProgramCode(ctx context.Context, productSlug string) (string, error)
	GetCertificationProgramCode(ctx context.Context, certificationID int64) (string, error)
	ListProgramCertificates(ctx context.Context, programCode string, filter domain.ProgramCertificateFilter) ([]domain.ProgramCertificate, int, error)
	GetProgramCertificate(ctx context.Context, programCode, productSlug string, certificationID int64) (*domain.ProgramCertificate, error)
}

type ProgramCertificateService struct {
	repo               ProgramCertificateRepository
	productCertService *ProductCertificationService
}

func NewProgramCertificateService(
	repo ProgramCertificateRepository,
	productCertService *ProductCertificationService,
) *ProgramCertificateService {
	return &ProgramCertificateService{
		repo:               repo,
		productCertService: productCertService,
	}
}

func (s *ProgramCertificateService) List(ctx context.Context, programCode string, filter domain.ProgramCertificateFilter) (domain.ProgramCertificateListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	items, total, err := s.repo.ListProgramCertificates(ctx, programCode, filter)
	if err != nil {
		return domain.ProgramCertificateListResponse{}, err
	}

	return domain.ProgramCertificateListResponse{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *ProgramCertificateService) Create(ctx context.Context, programCode string, payload domain.ProgramCertificatePayload) (domain.ProgramCertificate, error) {
	if err := s.ensureProgramConsistency(ctx, programCode, payload.ProductSlug, payload.CertificationID); err != nil {
		return domain.ProgramCertificate{}, err
	}

	certPayload := domain.ProductCertificationPayload{
		CertificationID: payload.CertificationID,
		CertificateNo:   payload.CertificateNo,
		IssueDate:       payload.IssueDate,
		ExpiryDate:      payload.ExpiryDate,
		StatusID:        payload.StatusID,
		DocumentFile:    payload.DocumentFile,
		Meta:            payload.Meta,
	}

	if _, err := s.productCertService.CreateProductCertification(ctx, payload.ProductSlug, certPayload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProgramCertificate{}, ErrCertificateNotFound
		}
		return domain.ProgramCertificate{}, err
	}

	record, err := s.repo.GetProgramCertificate(ctx, programCode, payload.ProductSlug, payload.CertificationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProgramCertificate{}, ErrCertificateNotFound
		}
		return domain.ProgramCertificate{}, err
	}
	if record == nil {
		return domain.ProgramCertificate{}, fmt.Errorf("created certificate not found")
	}
	return *record, nil
}

func (s *ProgramCertificateService) Update(ctx context.Context, programCode, productSlug string, certificationID int64, payload domain.ProductCertificationPayload) (domain.ProgramCertificate, error) {
	if err := s.ensureProgramConsistency(ctx, programCode, productSlug, certificationID); err != nil {
		return domain.ProgramCertificate{}, err
	}

	if _, err := s.productCertService.UpdateProductCertification(ctx, productSlug, certificationID, payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProgramCertificate{}, ErrCertificateNotFound
		}
		return domain.ProgramCertificate{}, err
	}

	record, err := s.repo.GetProgramCertificate(ctx, programCode, productSlug, certificationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProgramCertificate{}, ErrCertificateNotFound
		}
		return domain.ProgramCertificate{}, err
	}
	if record == nil {
		return domain.ProgramCertificate{}, ErrCertificateNotFound
	}
	return *record, nil
}

func (s *ProgramCertificateService) Delete(ctx context.Context, programCode, productSlug string, certificationID int64) error {
	if err := s.ensureProgramConsistency(ctx, programCode, productSlug, certificationID); err != nil {
		return err
	}
	if _, err := s.repo.GetProgramCertificate(ctx, programCode, productSlug, certificationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCertificateNotFound
		}
		return err
	}
	if err := s.productCertService.DeleteProductCertification(ctx, productSlug, certificationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCertificateNotFound
		}
		return err
	}
	return nil
}

func (s *ProgramCertificateService) ensureProgramConsistency(ctx context.Context, programCode, productSlug string, certificationID int64) error {
	productProgram, err := s.repo.GetProductProgramCode(ctx, productSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProductNotFound
		}
		return err
	}
	if productProgram == "" {
		return ErrProductNotFound
	}

	certProgram, err := s.repo.GetCertificationProgramCode(ctx, certificationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCertificationNotFound
		}
		return err
	}
	if certProgram == "" {
		return ErrCertificationNotFound
	}

	if productProgram != programCode || certProgram != programCode {
		return ErrProgramMismatch
	}

	return nil
}
