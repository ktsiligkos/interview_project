package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ktsiligkos/xm_project/internal/domain"
	userservice "github.com/ktsiligkos/xm_project/internal/service/user"
)

// CompanyService captures the service capabilities needed by the HTTP layer.
type UsersService interface {
	AuthenticateUser(ctx context.Context, user domain.UserLoginRequest) (token string, err error)
}

// CompaniesHandler exposes company endpoints.
type UsersHandler struct {
	service UsersService
	logger  *zap.Logger
}

// NewCompaniesHandler wires a service into the HTTP handler.
func NewUsersHandler(service UsersService, logger *zap.Logger) *UsersHandler {
	return &UsersHandler{service: service, logger: logger}
}

// Get returns a single company identified by the route param.
func (h *UsersHandler) Login(c *gin.Context) {
	logger := h.logger
	if logger != nil {
		if reqID := c.GetString("request_id"); reqID != "" {
			logger = logger.With(zap.String("request_id", reqID))
		}
	}

	var payload domain.UserLoginRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		if logger != nil {
			logger.Info("invalid login request body", zap.Error(err))
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, err := h.service.AuthenticateUser(c.Request.Context(), payload)
	if err != nil {
		if logger != nil {
			logger = logger.With(zap.String("email", payload.Email))
		}

		switch {
		case errors.Is(err, userservice.ErrNotFound):
			if logger != nil {
				logger.Warn("user not found", zap.Error(err))
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, userservice.ErrAuthFailed):
			if logger != nil {
				logger.Info("authentication failed", zap.Error(err))
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		case errors.Is(err, userservice.ErrTokenGenerationFailed):
			if logger != nil {
				logger.Error("token generation failed", zap.Error(err), zap.Stack("stack"))
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			if logger != nil {
				logger.Error("login failed", zap.Error(err), zap.Stack("stack"))
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		}
		return
	}

	if logger != nil {
		logger.Info("user authenticated")
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"token":  token,
	})
}
