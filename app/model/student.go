package model

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	StudentID    string     `json:"student_id"`
	ProgramStudy string     `json:"program_study"`
	AcademicYear string     `json:"academic_year"`
	AdvisorID    *uuid.UUID `json:"advisor_id"`
	CreatedAt    time.Time  `json:"created_at"`

	// Relasi
	User         User                   `json:"user,omitempty"`
	Advisor      *Lecturer              `json:"advisor,omitempty"`
	Achievements []AchievementReference `json:"achievements,omitempty"`
}

type StudentListResponse struct {
	ID           uuid.UUID `json:"id"`
	StudentID    string    `json:"student_id"`
	FullName     string    `json:"full_name"`
	ProgramStudy string    `json:"program_study"`
	AdvisorName  string    `json:"advisor_name"`
}

type StudentDetailResponse struct {
	ID           uuid.UUID `json:"id"`
	StudentID    string    `json:"student_id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	ProgramStudy string    `json:"program_study"`
	AdvisorName  string    `json:"advisor_name"`
}
