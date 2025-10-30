package company

import (
	"context"
	"errors"

	"github.com/ktsiligkos/xm_project/internal/domain"
)

// ErrNotFound indicates that the requested company does not exist in the repository.
var ErrNotFound = errors.New("company not found")
var ErrUniquenessViolation = errors.New("name already exists")

// Repository defines the contract the service layer relies on for company data access.
type Repository interface {
	GetCompanyByID(ctx context.Context, companyID string) (domain.Company, error)
	CreateCompany(ctx context.Context, company domain.Company) (domain.Company, error)
	DeleteCompanyByID(ctx context.Context, companyID string) error
	PatchCompanyByID(ctx context.Context, patchCompanyRequest domain.PatchCompanyRequest, uuid string, maxNumOfFields int) error
}
