package company

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ktsiligkos/xm_project/internal/domain"
	repository "github.com/ktsiligkos/xm_project/internal/repository/company"
)

// Exported errors map internal failures to business-level concerns
var (
	ErrNotFound            = errors.New("company not found")
	ErrInvalidInput        = errors.New("invalid input")
	ErrUniquenessViolation = errors.New("uniqueness violation")
	ErrValidationError     = errors.New("validation error")
)

// TODO: use reflection to find out
// The max number of fields of the PatchCompanyRequest
const maxNumOfFields = 5

// Service orchestrates the application's business logic for company
type Service struct {
	repo      repository.Repository
	publisher EventPublisher
}

// NewService creates a new company service bound to the provided repository
func NewService(repo repository.Repository, publisher EventPublisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

// Get retrieves a single company record, wrapping repository errors into business errors
func (s *Service) GetCompanyByID(ctx context.Context, companyID string) (domain.Company, error) {
	company, err := s.repo.GetCompanyByID(ctx, companyID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.Company{}, ErrNotFound
		}

		return domain.Company{}, err
	}

	return company, nil
}

// Delete removes the record from the persistent storage
func (s *Service) DeleteCompanyByID(ctx context.Context, companyID string) error {
	err := s.repo.DeleteCompanyByID(ctx, companyID)
	if err != nil {
		return ErrNotFound
	}

	s.publish(ctx, CompanyEvent{
		Operation: "company.deleted",
		Company: EventCompany{
			ID: companyID,
		},
	})

	return nil
}

// Patch performs a partial update on a record in the persistent storage
func (s *Service) PatchCompanyByID(ctx context.Context, partial_company domain.PatchCompanyRequest, uuid string) error {

	// TODO: add validation logic for PatchCompanyRequest
	if err := validatePatchCompanyRequestFields(partial_company); err != nil {
		return err
	}

	err := s.repo.PatchCompanyByID(ctx, partial_company, uuid, maxNumOfFields)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		if errors.Is(err, repository.ErrUniquenessViolation) {
			return fmt.Errorf("%w: %v", ErrUniquenessViolation, err)
		}

		return err
	}

	s.publish(ctx, CompanyEvent{
		Operation: "company.patched",
		Company: EventCompany{
			ID: uuid,
		},
	})

	return nil
}

// Create validates and persists a new company
func (s *Service) CreateCompany(ctx context.Context, company domain.Company) (domain.Company, error) {
	if err := validateCompanyFields(company); err != nil {
		return domain.Company{}, err
	}

	company, err := s.repo.CreateCompany(ctx, company)
	if err != nil {
		if errors.Is(err, repository.ErrUniquenessViolation) {
			return domain.Company{}, fmt.Errorf("%w: %v", ErrUniquenessViolation, err)
		}

		return domain.Company{}, err
	}

	s.publish(ctx, CompanyEvent{
		Operation: "company.created",
		Company:   toEventCompany(company),
	})

	return company, nil
}

// Validate the fields of the PatchCompanyRequest
func validatePatchCompanyRequestFields(company domain.PatchCompanyRequest) error {

	if company.Name == nil &&
		company.AmountOfEmployees == nil &&
		company.Registered == nil &&
		company.Type == nil {
		return fmt.Errorf("%w: %v", ErrValidationError, "no fields provided for update")
	}

	if company.Name != nil && len(*company.Name) > 15 {
		return fmt.Errorf("%w: %v", ErrValidationError, "name exceeds the limit of 15 characters")
	}

	if company.Description != nil {
		desc := strings.TrimSpace(*company.Description)
		if desc != "" && len(desc) > 3000 {
			return fmt.Errorf("%w: %v", ErrValidationError, "description exceeds the limit of 3000 characters")
		}
	}

	if company.Type != nil && !company.Type.IsValid() {
		return fmt.Errorf("%w: %v", ErrValidationError, "type value is invalid")

	}

	return nil
}

// Validate the fields of the domain company
func validateCompanyFields(company domain.Company) error {

	if len(company.Name) > 15 {
		return fmt.Errorf("%w: %v", ErrValidationError, "name exceeds the limit of 15 characters")
	}

	if company.Description != nil {
		desc := strings.TrimSpace(*company.Description)
		if desc != "" && len(desc) > 3000 {
			return fmt.Errorf("%w: %v", ErrValidationError, "description exceeds the limit of 3000 characters")
		}
	}

	if !company.Type.IsValid() {
		return fmt.Errorf("%w: %v", ErrValidationError, "type value is invalid")

	}

	return nil
}

func (s *Service) publish(ctx context.Context, event CompanyEvent) {
	if s.publisher == nil {
		return
	}

	if err := s.publisher.PublishCompanyEvent(ctx, event); err != nil {
		log.Printf("publish company event failed: %v", err)
	}
}

func toEventCompany(company domain.Company) EventCompany {
	return EventCompany{
		ID:                company.ID,
		Name:              company.Name,
		Description:       company.Description,
		AmountOfEmployees: company.AmountOfEmployees,
		Registered:        company.Registered,
		Type:              string(company.Type),
	}
}
