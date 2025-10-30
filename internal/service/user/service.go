package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ktsiligkos/xm_project/internal/auth"
	"github.com/ktsiligkos/xm_project/internal/domain"
	userrepository "github.com/ktsiligkos/xm_project/internal/repository/user"
	"golang.org/x/crypto/bcrypt"
)

// Exported errors map internal failures to business-level concerns
var (
	ErrNotFound              = errors.New("company not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrAuthFailed            = errors.New("authentication failed")
	ErrTokenGenerationFailed = errors.New("token generation failed")
)

// Service orchestrates the application's business logic for user
type Service struct {
	repo        userrepository.Repository
	tokenSecret []byte
	tokenTTL    time.Duration
}

// NewService creates a new user service bound to the provided repository
func NewService(repo userrepository.Repository, tokenSecret []byte, tokenTTL time.Duration) *Service {
	if tokenTTL <= 0 {
		tokenTTL = time.Hour
	}

	return &Service{repo: repo, tokenSecret: tokenSecret, tokenTTL: tokenTTL}
}

// Get retrieves a single user in order to authenticate the user
func (s *Service) AuthenticateUser(ctx context.Context, user domain.UserLoginRequest) (string, error) {
	//TODO: add validatin validation for the contents of the UserLoginRequest
	userFromDB, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		if errors.Is(err, userrepository.ErrNotFound) {
			return "", ErrNotFound
		}

		return "", err
	}

	verified := checkPasswordHash(user.Password, userFromDB.Password)
	if !verified {
		return "", fmt.Errorf("invalid password for email %s: %w", user.Email, ErrAuthFailed)
	}

	if len(s.tokenSecret) == 0 {
		return "", fmt.Errorf("token secret not configured")
	}

	token, err := auth.GenerateJWT(userFromDB.ID, s.tokenSecret, s.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGenerationFailed, err)
	}

	return token, nil
}

func checkPasswordHash(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}
