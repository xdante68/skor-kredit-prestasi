package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FullName     string     `json:"full_name"`
	RoleID       *uuid.UUID `json:"role_id"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RefreshToken string     `json:"-"`

	// Relasi
	Role     Role      `json:"role,omitempty"`
	Student  *Student  `json:"student,omitempty"`
	Lecturer *Lecturer `json:"lecturer,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateUserRequest struct {
	Username string              `json:"username" validate:"required"`
	Email    string              `json:"email" validate:"required,email"`
	Password string              `json:"password" validate:"required,min=6"`
	FullName string              `json:"full_name" validate:"required"`
	Role     string              `json:"role" validate:"required,oneof=admin mahasiswa dosen_wali"`
	Student  *CreateStudentData  `json:"student,omitempty"`
	Lecturer *CreateLecturerData `json:"lecturer,omitempty"`
}

type CreateStudentData struct {
	ProgramStudy string `json:"program_study" validate:"required"`
	AcademicYear string `json:"academic_year" validate:"required"`
}

type CreateLecturerData struct {
	Department string `json:"department" validate:"required"`
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
	ID        uint
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RefreshTokenResponse struct {
	Token string `json:"token" binding:"required"`
}

type ChangeRoleRequest struct {
	Role     string              `json:"role" validate:"required,oneof=admin mahasiswa dosen_wali"`
	Student  *CreateStudentData  `json:"student,omitempty"`
	Lecturer *CreateLecturerData `json:"lecturer,omitempty"`
}
