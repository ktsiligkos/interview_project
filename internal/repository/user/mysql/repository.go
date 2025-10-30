package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ktsiligkos/xm_project/internal/domain"
	userrepository "github.com/ktsiligkos/xm_project/internal/repository/user"
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
func (r *MySQLRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {

	query := "SELECT id, name, email, password_hash FROM users WHERE email = ?"

	var (
		user domain.User
	)

	if err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, userrepository.ErrNotFound
		}

		return domain.User{}, fmt.Errorf("query company: %w", err)
	}

	return user, nil
}
