package repositories

import (
	"context"
	"finanvilla/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*entities.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	RevokeToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type PostgresRefreshTokenRepository struct {
	db *sqlx.DB
}

func NewPostgresRefreshTokenRepository(db *sqlx.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)
}

func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	var rt entities.RefreshToken
	query := `
        SELECT id, user_id, token, expires_at, revoked, created_at, updated_at, revoked_at
        FROM refresh_tokens
        WHERE token = $1 AND NOT revoked`

	err := r.db.GetContext(ctx, &rt, query, token)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *PostgresRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
        UPDATE refresh_tokens
        SET revoked = true, revoked_at = CURRENT_TIMESTAMP
        WHERE user_id = $1 AND NOT revoked`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *PostgresRefreshTokenRepository) RevokeToken(ctx context.Context, token string) error {
	query := `
        UPDATE refresh_tokens
        SET revoked = true, revoked_at = CURRENT_TIMESTAMP
        WHERE token = $1 AND NOT revoked`

	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
