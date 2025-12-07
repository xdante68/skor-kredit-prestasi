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
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	// Relasi
	Users       []User       `json:"users,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`

	// Relasi
	Roles []Role `json:"roles,omitempty"`
}
