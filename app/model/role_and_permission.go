package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleAdmin     = "admin"
	RoleMahasiswa = "mahasiswa"
	RoleDosenWali = "dosen_wali"
)

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name        string    `gorm:"size:50;unique;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relasi
	Users       []User       `gorm:"foreignKey:RoleID" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name        string    `gorm:"size:100;unique;not null" json:"name"`
	Resource    string    `gorm:"size:50;not null" json:"resource"`
	Action      string    `gorm:"size:50;not null" json:"action"`
	Description string    `gorm:"type:text" json:"description"`

	// Relasi
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}
