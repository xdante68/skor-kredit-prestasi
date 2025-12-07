package model

import (
	"time"

	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	LecturerID string    `json:"lecturer_id"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`

	// Relasi
	User     User      `json:"user,omitempty"`
	Advisees []Student `json:"advisees,omitempty"`
}

type LecturerListResponse struct {
	ID         uuid.UUID `json:"id"`
	LecturerID string    `json:"lecturer_id"`
	FullName   string    `json:"full_name"`
	Department string    `json:"department"`
}
