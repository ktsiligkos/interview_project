package domain

type CompanyType string

func (t CompanyType) IsValid() bool {
	switch t {
	case Corporations, NonProfit, Cooperative, SoleProprietor:
		return true
	}
	return false
}

const (
	Corporations   CompanyType = "Corporations"
	NonProfit      CompanyType = "NonProfit"
	Cooperative    CompanyType = "Cooperative"
	SoleProprietor CompanyType = "Sole Proprietorship"
)

type Company struct {
	ID                string      `json:"id" binding:"required"`
	Name              string      `json:"name" binding:"required"`
	Description       *string     `json:"description,omitempty"`
	AmountOfEmployees int         `json:"amount_of_employees" binding:"required"`
	Registered        bool        `json:"registered" binding:"required"`
	Type              CompanyType `json:"type" binding:"required"`
}

type PatchCompanyRequest struct {
	Name              *string      `json:"name,omitempty"`
	Description       *string      `json:"description,omitempty"`
	AmountOfEmployees *int         `json:"amount_of_employees,omitempty"`
	Registered        *bool        `json:"registered,omitempty"`
	Type              *CompanyType `json:"type,omitempty"`
}
