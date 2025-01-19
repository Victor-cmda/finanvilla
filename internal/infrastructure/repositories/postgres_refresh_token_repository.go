package repositories

import (
	"context"
	"errors"
	"finanvilla/internal/domain/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresRefreshTokenRepository struct {
	db *gorm.DB
}

func NewPostgresRefreshTokenRepository(db *gorm.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{
		db: db,
	}
}

func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	var refreshToken entities.RefreshToken
	result := r.db.WithContext(ctx).
		Where("token = ? AND NOT revoked AND expires_at > ?", token, time.Now()).
		First(&refreshToken)

	if result.Error != nil {
		return nil, result.Error
	}

	return &refreshToken, nil
}

func (r *PostgresRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entities.RefreshToken{}).
		Where("user_id = ? AND NOT revoked", userID).
		Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": now,
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at < ? OR (revoked = true AND revoked_at < ?)",
			time.Now(),
			time.Now().Add(-30*24*time.Hour)).
		Delete(&entities.RefreshToken{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	var refreshToken entities.RefreshToken
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return refreshToken.Revoked, nil
}

func (r *PostgresRefreshTokenRepository) RevokeToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Model(&entities.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("token not found")
	}

	return nil
}
