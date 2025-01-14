package repositories

import (
	"context"
	"errors"
	"finanvilla/internal/domain/entities"

	"gorm.io/gorm"
)

type postgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) *postgresUserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *postgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *postgresUserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.User{}, id).Error
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).
		Preload("Settings").
		Preload("Permissions").
		First(&user, "id = ?", id).Error
	return &user, err
}

// GetByEmail busca um usuário pelo email
func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).
		Preload("Settings").
		Preload("Permissions").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List retorna uma lista paginada de usuários
func (r *postgresUserRepository) List(ctx context.Context, page, limit int) ([]entities.User, int, error) {
	var users []entities.User
	var total int64

	// Calcula o offset
	offset := (page - 1) * limit

	// Conta o total de registros
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Busca os usuários com paginação
	err := r.db.WithContext(ctx).
		Preload("Settings").
		Preload("Permissions").
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

// UpdateSettings atualiza as configurações de um usuário
func (r *postgresUserRepository) UpdateSettings(ctx context.Context, settings *entities.UserSettings) error {
	return r.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// Verifica se as configurações existem
			var existingSettings entities.UserSettings
			err := tx.Where("user_id = ?", settings.UserID).First(&existingSettings).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Se não existir, cria novo
					return tx.Create(settings).Error
				}
				return err
			}

			// Se existir, atualiza
			return tx.Model(&existingSettings).Updates(settings).Error
		})
}

// AddPermissions adiciona permissões a um usuário
func (r *postgresUserRepository) AddPermissions(ctx context.Context, userID string, permissions []string) error {
	return r.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// Busca o usuário
			var user entities.User
			if err := tx.First(&user, "id = ?", userID).Error; err != nil {
				return err
			}

			// Busca as permissões a serem adicionadas
			var permsToAdd []entities.Permission
			if err := tx.Where("name IN ?", permissions).Find(&permsToAdd).Error; err != nil {
				return err
			}

			// Adiciona as permissões ao usuário
			return tx.Model(&user).Association("Permissions").Append(permsToAdd)
		})
}

// RemovePermissions remove permissões de um usuário
func (r *postgresUserRepository) RemovePermissions(ctx context.Context, userID string, permissions []string) error {
	return r.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// Busca o usuário
			var user entities.User
			if err := tx.First(&user, "id = ?", userID).Error; err != nil {
				return err
			}

			// Busca as permissões a serem removidas
			var permsToRemove []entities.Permission
			if err := tx.Where("name IN ?", permissions).Find(&permsToRemove).Error; err != nil {
				return err
			}

			// Remove as permissões do usuário
			return tx.Model(&user).Association("Permissions").Delete(permsToRemove)
		})
}
