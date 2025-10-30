package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ktsiligkos/xm_project/internal/domain"
	companyservice "github.com/ktsiligkos/xm_project/internal/service/company"
)

// CompanyService captures the service capabilities needed by the HTTP layer.
type CompanyService interface {
	GetCompanyByID(ctx context.Context, companyID string) (domain.Company, error)
	CreateCompany(ctx context.Context, company domain.Company) (domain.Company, error)
	DeleteCompanyByID(ctx context.Context, companyID string) error
	PatchCompanyByID(ctx context.Context, patchCompanyRequest domain.PatchCompanyRequest, uuid string) error
}

// CompaniesHandler exposes company endpoints.
type CompaniesHandler struct {
	service CompanyService
	logger  *zap.Logger
}

// NewCompaniesHandler wires a service into the HTTP handler.
func NewCompaniesHandler(service CompanyService, logger *zap.Logger) *CompaniesHandler {
	return &CompaniesHandler{service: service, logger: logger}
}

// Get returns a single company identified by the route param.
func (h *CompaniesHandler) Get(c *gin.Context) {
	logger := h.requestLogger(c)
	companyID := c.Param("uuid")
	if logger != nil {
		logger = logger.With(zap.String("company_id", companyID))
	}

	company, err := h.service.GetCompanyByID(c.Request.Context(), companyID)
	if err != nil {
		switch err {
		case companyservice.ErrNotFound:
			if logger != nil {
				logger.Info("company not found", zap.Error(err))
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		default:
			if logger != nil {
				logger.Error("failed to fetch company", zap.Error(err), zap.Stack("stack"))
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch company"})
		}
		return
	}

	if logger != nil {
		logger.Info("company fetched")
	}

	c.JSON(http.StatusOK, company)
}

// Create persists a new company received from the request payload.
func (h *CompaniesHandler) Create(c *gin.Context) {
	logger := h.requestLogger(c)

	var payload createCompanyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		if logger != nil {
			logger.Info("invalid request body", zap.Error(err))
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if logger != nil {
		hasDescription := payload.Description != nil && strings.TrimSpace(*payload.Description) != ""
		logger.Debug("create company payload",
			zap.String("name", payload.Name),
			zap.Int("employees", payload.AmountOfEmployees),
			zap.Bool("registered", payload.Registered),
			zap.String("type", string(payload.Type)),
			zap.Bool("has_description", hasDescription),
		)
	}

	company, err := h.service.CreateCompany(c.Request.Context(), payload.toDomain())
	if err != nil {
		if logger != nil {
			logger = logger.With(zap.String("company_name", payload.Name))
		}

		switch {
		case errors.Is(err, companyservice.ErrValidationError):
			msg := validationMessage(err)
			if logger != nil {
				logger.Info("validation failed", zap.Error(err))
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		case errors.Is(err, companyservice.ErrUniquenessViolation):
			if logger != nil {
				logger.Warn("uniqueness violation", zap.Error(err))
			}
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, companyservice.ErrInvalidInput):
			if logger != nil {
				logger.Warn("invalid input", zap.Error(err))
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			if logger != nil {
				logger.Error("failed to create company", zap.Error(err), zap.Stack("stack"))
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create company"})
		}
		return
	}

	if logger != nil {
		logger.Info("company created", zap.String("company_id", company.ID))
	}

	c.JSON(http.StatusCreated, company)
}

// Create persists a new company received from the request payload.
func (h *CompaniesHandler) Patch(c *gin.Context) {
	logger := h.requestLogger(c)
	companyID := c.Param("uuid")
	if logger != nil {
		logger = logger.With(zap.String("company_id", companyID))
	}

	var payload domain.PatchCompanyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		if logger != nil {
			logger.Info("invalid patch request body", zap.Error(err))
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.service.PatchCompanyByID(c.Request.Context(), payload, companyID)
	if err != nil {
		switch {
		case errors.Is(err, companyservice.ErrNotFound):
			if logger != nil {
				logger.Info("company not found for patch", zap.Error(err))
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		case errors.Is(err, companyservice.ErrValidationError):
			msg := validationMessage(err)
			if logger != nil {
				logger.Info("validation failed on patch", zap.String("reason", msg))
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		case errors.Is(err, companyservice.ErrUniquenessViolation):
			if logger != nil {
				logger.Warn("uniqueness violation on patch", zap.Error(err))
			}
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, companyservice.ErrInvalidInput):
			if logger != nil {
				logger.Warn("invalid input on patch", zap.Error(err))
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			if logger != nil {
				logger.Error("failed to patch company", zap.Error(err), zap.Stack("stack"))
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update company"})
		}
		return
	}

	if logger != nil {
		logger.Info("company patched")
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success"})
}

// Delete removes a company received from the request payload.
func (h *CompaniesHandler) Delete(c *gin.Context) {
	logger := h.requestLogger(c)
	companyID := c.Param("uuid")
	if logger != nil {
		logger = logger.With(zap.String("company_id", companyID))
	}

	err := h.service.DeleteCompanyByID(c.Request.Context(), companyID)

	if err != nil {
		switch err {
		case companyservice.ErrNotFound:
			if logger != nil {
				logger.Info("company not found for delete", zap.Error(err))
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		default:
			if logger != nil {
				logger.Error("failed to delete company", zap.Error(err), zap.Stack("stack"))
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete the company"})
		}
		return
	}

	if logger != nil {
		logger.Info("company deleted")
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

type createCompanyRequest struct {
	Name              string             `json:"name" binding:"required"`
	Description       *string            `json:"description,omitempty"`
	AmountOfEmployees int                `json:"amount_of_employees" binding:"required"`
	Registered        bool               `json:"registered" binding:"required"`
	Type              domain.CompanyType `json:"type" binding:"required"`
}

// converts the request to the company domain value
// it generates also the UUID
func (r createCompanyRequest) toDomain() domain.Company {
	return domain.Company{
		ID:                uuid.New().String(),
		Name:              r.Name,
		Description:       r.Description,
		AmountOfEmployees: r.AmountOfEmployees,
		Registered:        r.Registered,
		Type:              r.Type,
	}
}

func (h *CompaniesHandler) requestLogger(c *gin.Context) *zap.Logger {
	if h.logger == nil {
		return nil
	}

	logger := h.logger
	if reqID := c.GetString("request_id"); reqID != "" {
		logger = logger.With(zap.String("request_id", reqID))
	}

	return logger
}

func validationMessage(err error) string {
	prefix := companyservice.ErrValidationError.Error() + ": "
	msg := err.Error()
	if strings.HasPrefix(msg, prefix) {
		return strings.TrimPrefix(msg, prefix)
	}

	return msg
}
