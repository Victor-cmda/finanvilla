package entities

import "time"

type UserSettings struct {
	ID                   string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID               string    `json:"userId" gorm:"not null"`
	Theme                string    `json:"theme" gorm:"default:'light'"`
	Language             string    `json:"language" gorm:"default:'pt-BR'"`
	NotificationsEnabled bool      `json:"notificationsEnabled" gorm:"default:true"`
	Currency             string    `json:"currency" gorm:"default:'BRL'"`
	DateFormat           string    `json:"dateFormat" gorm:"default:'DD/MM/YYYY'"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}
