package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementMongo struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID       string             `bson:"studentId" json:"studentId"`
	AchievementType string             `bson:"achievementType" json:"achievementType"`
	Title           string             `bson:"title" json:"title"`
	Description     string             `bson:"description" json:"description"`
	Details         AchievementDetails `bson:"details" json:"details"`
	Attachments     []Attachment       `bson:"attachments" json:"attachments"`
	Tags            []string           `bson:"tags" json:"tags"`
	Points          int                `bson:"points" json:"points"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type AchievementDetails struct {
	// Competition
	CompetitionName  string `bson:"competitionName,omitempty" json:"competitionName,omitempty"`
	CompetitionLevel string `bson:"competitionLevel,omitempty" json:"competitionLevel,omitempty"`
	Rank             int    `bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType        string `bson:"medalType,omitempty" json:"medalType,omitempty"`

	// Publication
	PublicationType  string   `bson:"publicationType,omitempty" json:"publicationType,omitempty"`
	PublicationTitle string   `bson:"publicationTitle,omitempty" json:"publicationTitle,omitempty"`
	Authors          []string `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher        string   `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN             string   `bson:"issn,omitempty" json:"issn,omitempty"`

	// Organization
	OrganizationName string              `bson:"organizationName,omitempty" json:"organizationName,omitempty"`
	Position         string              `bson:"position,omitempty" json:"position,omitempty"`
	Period           *OrganizationPeriod `bson:"period,omitempty" json:"period,omitempty"`

	// Certification
	CertificationName   string    `bson:"certificationName,omitempty" json:"certificationName,omitempty"`
	IssuedBy            string    `bson:"issuedBy,omitempty" json:"issuedBy,omitempty"`
	CertificationNumber string    `bson:"certificationNumber,omitempty" json:"certificationNumber,omitempty"`
	ValidUntil          time.Time `bson:"validUntil,omitempty" json:"validUntil,omitempty"`

	// General
	EventDate time.Time `bson:"eventDate,omitempty" json:"eventDate,omitempty"`
	Location  string    `bson:"location,omitempty" json:"location,omitempty"`
	Organizer string    `bson:"organizer,omitempty" json:"organizer,omitempty"`
	Score     float64   `bson:"score,omitempty" json:"score,omitempty"`
}

type OrganizationPeriod struct {
	Start time.Time `bson:"start" json:"start"`
	End   time.Time `bson:"end" json:"end"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"fileName"`
	FileURL    string    `bson:"fileUrl" json:"fileUrl"`
	FileType   string    `bson:"fileType" json:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}

type CreateAchievementRequest struct {
	Title               string               `json:"title" binding:"required"`
	AchievementType     string               `json:"achievement_type" binding:"required,oneof=academic competition organization publication certification other"`
	Description         string               `json:"description"`
	CompetitionDetails  *CompetitionRequest  `json:"competition_details,omitempty"`
	PublicationDetails  *PublicationRequest  `json:"publication_details,omitempty"`
	OrganizationDetails *OrganizationRequest `json:"organization_details,omitempty"`
	EventDate           string               `json:"event_date"`
	Tags                []string             `json:"tags"`
}

type UpdateAchievementRequest struct {
	Title               *string              `json:"title,omitempty"`
	AchievementType     *string              `json:"achievement_type,omitempty" binding:"omitempty,oneof=academic competition organization publication certification other"`
	Description         *string              `json:"description,omitempty"`
	CompetitionDetails  *CompetitionRequest  `json:"competition_details,omitempty"`
	PublicationDetails  *PublicationRequest  `json:"publication_details,omitempty"`
	OrganizationDetails *OrganizationRequest `json:"organization_details,omitempty"`
	EventDate           *string              `json:"event_date,omitempty"`
	Tags                *[]string            `json:"tags,omitempty"`
}

type AchievementResponse struct {
	ID              uuid.UUID          `json:"id"`
	MongoID         string             `json:"mongo_achievement_id"`
	StudentID       uuid.UUID          `json:"student_id"`
	StudentName     string             `json:"student_name,omitempty"`
	Status          string             `json:"status"`
	AchievementType string             `json:"achievement_type"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Details         AchievementDetails `json:"details"`
	Attachments     []Attachment       `json:"attachments"`
	Tags            []string           `json:"tags"`
	Points          int                `json:"points"`
	RejectionNote   string             `json:"rejection_note,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type CompetitionRequest struct {
	CompetitionName  string `json:"competition_name"`
	CompetitionLevel string `json:"competition_level"`
	Rank             int    `json:"rank"`
	MedalType        string `json:"medal_type"`
}

type PublicationRequest struct {
	PublicationTitle string   `json:"publication_title"`
	Authors          []string `json:"authors"`
	Publisher        string   `json:"publisher"`
	ISSN             string   `json:"issn"`
}

type OrganizationRequest struct {
	OrganizationName string `json:"organization_name"`
	Position         string `json:"position"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
}

type VerifyRequest struct {
	Points int `json:"points" form:"points" binding:"required,gt=0"`
}

type RejectRequest struct {
	RejectionNote string `json:"rejection_note" binding:"required"`
}

type AchievementHistoryResponse struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	VerifierName  string     `json:"verifier_name,omitempty"`
	RejectionNote string     `json:"rejection_note,omitempty"`
	Points        int        `json:"points,omitempty"`
}

type StatItem struct {
	Label string `json:"label" bson:"_id"`
	Count int    `json:"count" bson:"count"`
}
type TopStudent struct {
	StudentID          string `json:"student_id"`
	StudentName        string `json:"student_name"`
	Program            string `json:"program_study"`
	TotalAchievements  int    `json:"total_achievements"`
	TotalPoints        int    `json:"total_points"`
}

type StatsResponse struct {
	TotalAchievements int64        `json:"total_achievements"`
	ByType            []StatItem   `json:"by_type"`
	ByLevel           []StatItem   `json:"by_competition_level"`
	ByPeriod          []StatItem   `json:"by_period"`
	TopStudents       []TopStudent `json:"top_students"`
}

type StudentStatsResponse struct {
	StudentProfile TopStudent `json:"student_profile"`
	Stats          StatsResponse `json:"stats"`
}