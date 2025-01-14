package entities

import (
	"finanvilla/internal/domain/enums"
	"time"
)

type User struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string         `json:"name" gorm:"not null"`
	Email       string         `json:"email" gorm:"unique;not null"`
	Password    string         `json:"-" gorm:"not null"` // O "-" oculta o campo nas respostas JSON
	UserType    enums.UserType `json:"userType" gorm:"type:varchar(20);not null"`
	Active      bool           `json:"active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   *time.Time     `json:"deletedAt,omitempty" gorm:"index"`
	Settings    *UserSettings  `json:"settings" gorm:"foreignKey:UserID"`
	Permissions []Permission   `json:"permissions" gorm:"many2many:user_permissions;"`
}
