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
	ID                 uuid.UUID         `json:"id"`
	StudentID          uuid.UUID         `json:"student_id"`
	MongoAchievementID string            `json:"mongo_achievement_id"`
	Status             AchievementStatus `json:"status"`
	SubmittedAt        *time.Time        `json:"submitted_at"`
	VerifiedAt         *time.Time        `json:"verified_at"`
	VerifiedBy         *uuid.UUID        `json:"verified_by"`
	RejectionNote      string            `json:"rejection_note"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`

	// Relasi
	Student  Student `json:"student,omitempty"`
	Verifier *User   `json:"verifier,omitempty"`
}
