package album

import (
	"context"
	"errors"

	"github.com/ktsiligkos/xm_project/internal/domain"
)

// ErrNotFound indicates that the requested company does not exist in the repository.
var ErrNotFound = errors.New("user not found")

type Repository interface {
	GetUserByEmail(tx context.Context, email string) (user domain.User, err error)
}
