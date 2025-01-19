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

func (r *postgresUserRepository) List(ctx context.Context, page, limit int) ([]entities.User, int, error) {
	var users []entities.User
	var total int64

	offset := (page - 1) * limit

	if err := r.db.WithContext(ctx).Model(&entities.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

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

func (r *postgresUserRepository) UpdateSettings(ctx context.Context, settings *entities.UserSettings) error {
	return r.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			var existingSettings entities.UserSettings
			err := tx.Where("user_id = ?", settings.UserID).First(&existingSettings).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return tx.Create(settings).Error
				}
				return err
			}

			return tx.Model(&existingSettings).Updates(settings).Error
		})
}

func (r *postgresUserRepository) AddPermissions(ctx context.Context, userID string, permissions []string) error {
	return r.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			var user entities.User
			if err := tx.First(&user, "id = ?", userID).Error; err != nil {
				return err
			}

			var permsToAdd []entities.Permission
			if err := tx.Where("name IN ?", permissions).Find(&permsToAdd).Error; err != nil {
				return err
			}

			return tx.Model(&user).Association("Permissions").Append(permsToAdd)
		})
}

func (r *postgresUserRepository) RemovePermissions(ctx context.Context, userID string, permissions []string) error {
	return r.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			var user entities.User
			if err := tx.First(&user, "id = ?", userID).Error; err != nil {
				return err
			}

			var permsToRemove []entities.Permission
			if err := tx.Where("name IN ?", permissions).Find(&permsToRemove).Error; err != nil {
				return err
			}

			return tx.Model(&user).Association("Permissions").Delete(permsToRemove)
		})
}
