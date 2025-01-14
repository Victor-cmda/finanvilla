package entities

import (
	"finanvilla/internal/domain/enums"
	"time"
)

type Permission struct {
	ID          string           `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        enums.Permission `json:"name" gorm:"unique;not null"`
	Description string           `json:"description"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}
