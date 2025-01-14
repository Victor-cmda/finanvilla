// internal/domain/repositories/user_repository.go
package repositories

import (
	"context"
	"finanvilla/internal/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	List(ctx context.Context, page, limit int) ([]entities.User, int, error)
	UpdateSettings(ctx context.Context, settings *entities.UserSettings) error
	AddPermissions(ctx context.Context, userID string, permissions []string) error
	RemovePermissions(ctx context.Context, userID string, permissions []string) error
}
