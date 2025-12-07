package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Username     string     `gorm:"size:50;unique;not null" json:"username"`
	Email        string     `gorm:"size:100;unique;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	FullName     string     `gorm:"size:100;not null" json:"full_name"`
	RoleID       *uuid.UUID `gorm:"type:uuid" json:"role_id"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	RefreshToken string     `gorm:"type:text" json:"-"`

	// Relasi
	Role     Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Student  *Student  `gorm:"foreignKey:UserID" json:"student,omitempty"`
	Lecturer *Lecturer `gorm:"foreignKey:UserID" json:"lecturer,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=admin mahasiswa dosen_wali"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name"`
	Password string `json:"password" validate:"omitempty,min=6"`
}

type LoginUser struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Role     string    `json:"role"`
}

type ProfileData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type JWTClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	Type        string    `json:"type"`
	jwt.RegisteredClaims
}

type BlacklistedToken struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RefreshTokenResponse struct {
	Token string `json:"token" binding:"required"`
}

type ChangeRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin mahasiswa dosen_wali"`
}
