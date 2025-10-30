package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	driver "github.com/go-sql-driver/mysql"
	"github.com/ktsiligkos/xm_project/internal/domain"
	companyrepository "github.com/ktsiligkos/xm_project/internal/repository/company"
)

// Making sure to prevent string errors when building queries
const (
	columnID                = "id"
	columnName              = "name"
	columnDescription       = "description"
	columnAmountOfEmployees = "amount_of_employees"
	columnRegistered        = "registered"
	columnType              = "type"
)

// MySQLRepository persists companies using a MySQL-compatible database
type MySQLRepository struct {
	db *sql.DB
}

// NewMySQL creates a repository backed by the supplied database handle
func NewMySQL(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// Get returns a single company by ID
func (r *MySQLRepository) GetCompanyByID(ctx context.Context, companyID string) (domain.Company, error) {

	query := fmt.Sprintf(
		`SELECT %s, %s, %s, %s, %s, %s FROM companies WHERE %s = ?`,
		columnID,
		columnName,
		columnDescription,
		columnAmountOfEmployees,
		columnRegistered,
		columnType,
		columnID,
	)

	var (
		company domain.Company
	)

	if err := r.db.QueryRowContext(ctx, query, companyID).Scan(&company.ID, &company.Name, &company.Description, &company.AmountOfEmployees, &company.Registered, &company.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Company{}, companyrepository.ErrNotFound
		}

		return domain.Company{}, fmt.Errorf("query company: %w", err)
	}

	return company, nil
}

// Delete removes a company record
func (r *MySQLRepository) DeleteCompanyByID(ctx context.Context, companyID string) error {
	query := fmt.Sprintf(
		`DELETE FROM companies WHERE %s = ?`,
		columnID,
	)

	result, err := r.db.ExecContext(ctx, query, companyID)
	if err != nil {
		return fmt.Errorf("delete company: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete company, rows affected: %w", err)
	}
	if rows == 0 {
		return companyrepository.ErrNotFound
	}

	return nil
}

// Create writes a new company record and returns the stored record
func (r *MySQLRepository) CreateCompany(ctx context.Context, company domain.Company) (domain.Company, error) {

	queryWithID := fmt.Sprintf(
		`INSERT INTO companies (%s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?)`,
		columnID,
		columnName,
		columnDescription,
		columnAmountOfEmployees,
		columnRegistered,
		columnType,
	)

	if _, err := r.db.ExecContext(ctx, queryWithID, company.ID, company.Name, company.Description, company.AmountOfEmployees, company.Registered, company.Type); err != nil {
		if uniquenessViolation(err) {
			return domain.Company{}, companyrepository.ErrUniquenessViolation
		}
		return domain.Company{}, fmt.Errorf("insert company with id: %w", err)
	}

	return company, nil
}

// Returns true if the supplied name field already exists
func uniquenessViolation(err error) bool {
	var mysqlErr *driver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

// Patch partially updates an existing record
func (r *MySQLRepository) PatchCompanyByID(ctx context.Context, patchCompanyRequest domain.PatchCompanyRequest, uuid string, maxNumOfFields int) error {

	fields, field_values := createPatchRequestFields(patchCompanyRequest, uuid, maxNumOfFields)
	query := fmt.Sprintf(
		`UPDATE  companies SET %s WHERE id = ?`,
		fields,
	)
	field_values = append(field_values, uuid)

	result, err := r.db.ExecContext(ctx, query, field_values...)
	if err != nil {
		if uniquenessViolation(err) {
			return companyrepository.ErrUniquenessViolation
		}
		return fmt.Errorf("patch company: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete company, rows affected: %w", err)
	}
	if rows == 0 {
		return companyrepository.ErrNotFound
	}

	return nil
}

func createPatchRequestFields(patchCompanyRequest domain.PatchCompanyRequest, uuid string, companyColumnSize int) (string, []any) {
	fields := make([]string, 0, companyColumnSize)
	field_values := make([]any, 0, companyColumnSize)

	if patchCompanyRequest.Name != nil {
		fields = append(fields, fmt.Sprintf("%s = ?", columnName))
		field_values = append(field_values, *patchCompanyRequest.Name)
	}
	if patchCompanyRequest.Description != nil {
		fields = append(fields, fmt.Sprintf("%s = ?", columnDescription))
		field_values = append(field_values, *patchCompanyRequest.Description)
	}
	if patchCompanyRequest.AmountOfEmployees != nil {
		fields = append(fields, fmt.Sprintf("%s = ?", columnAmountOfEmployees))
		field_values = append(field_values, *patchCompanyRequest.AmountOfEmployees)
	}
	if patchCompanyRequest.Registered != nil {
		fields = append(fields, fmt.Sprintf("%s = ?", columnRegistered))
		field_values = append(field_values, *patchCompanyRequest.Registered)
	}
	if patchCompanyRequest.Type != nil {
		fields = append(fields, fmt.Sprintf("%s = ?", columnType))
		field_values = append(field_values, string(*patchCompanyRequest.Type))
	}
	return strings.Join(fields, ", "), field_values
}
