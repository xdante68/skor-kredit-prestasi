package model

import (
	"time"

	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	LecturerID string    `gorm:"size:20;unique;not null" json:"lecturer_id"` // NIP/NIDN
	Department string    `gorm:"size:100" json:"department"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relasi
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Advisees []Student `gorm:"foreignKey:AdvisorID" json:"advisees,omitempty"` // Mahasiswa bimbingan
}