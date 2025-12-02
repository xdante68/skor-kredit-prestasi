package model

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	NIM          string     `gorm:"column:nim;type:varchar(20);unique" json:"student_id"`
	ProgramStudy string     `gorm:"size:100" json:"program_study"`
	AcademicYear string     `gorm:"size:10" json:"academic_year"`
	AdvisorID    *uuid.UUID `gorm:"type:uuid" json:"advisor_id"`
	CreatedAt    time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relasi
	User         User                   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Advisor      *Lecturer              `gorm:"foreignKey:AdvisorID" json:"advisor,omitempty"`
	Achievements []AchievementReference `gorm:"foreignKey:StudentID" json:"achievements,omitempty"`
}
