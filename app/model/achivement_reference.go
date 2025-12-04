package model

import (
	"time"

	"github.com/google/uuid"
)

type AchievementStatus string

const (
	StatusDraft     AchievementStatus = "draft"
	StatusSubmitted AchievementStatus = "submitted"
	StatusVerified  AchievementStatus = "verified"
	StatusRejected  AchievementStatus = "rejected"
	StatusDeleted   AchievementStatus = "deleted"
)

type AchievementReference struct {
	ID                 uuid.UUID         `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	StudentID          uuid.UUID         `gorm:"type:uuid;not null" json:"student_id"`
	MongoAchievementID string            `gorm:"size:24;not null" json:"mongo_achievement_id"`
	Status             AchievementStatus `gorm:"type:achievement_status_enum;default:'draft'" json:"status"`
	SubmittedAt        *time.Time        `json:"submitted_at"`
	VerifiedAt         *time.Time        `json:"verified_at"`
	VerifiedBy         *uuid.UUID        `gorm:"type:uuid" json:"verified_by"`
	RejectionNote      string            `gorm:"type:text" json:"rejection_note"`
	CreatedAt          time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relasi
	Student  Student `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Verifier *User   `gorm:"foreignKey:VerifiedBy" json:"verifier,omitempty"`
}
