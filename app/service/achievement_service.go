package service

import (
	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AchievementService struct {
	repo        *repo.AchievementRepo
	studentRepo *repo.StudentRepo
}

func NewAchievementService(repo *repo.AchievementRepo, studentRepo *repo.StudentRepo) *AchievementService {
	return &AchievementService{
		repo:        repo,
		studentRepo: studentRepo,
	}
}

// /api/v1/achievements
func (s *AchievementService) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	data, err := s.repo.FindAll(role, userID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}

	return c.JSON(model.SuccessResponse[[]model.AchievementResponse]{Success: true, Data: data})
}

// /api/v1/achievements/:id
func (s *AchievementService) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid AchievementId"})
	}

	data, err := s.repo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{Success: false, Message: "Achievement not found"})
	}

	return c.JSON(model.SuccessResponse[*model.AchievementResponse]{Success: true, Data: data})
}

// /api/v1/achievements
func (s *AchievementService) Create(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "mahasiswa" {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "Mahasiswa Only"})
	}

	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid input"})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{Success: false, Message: "Student profile not found"})
	}

	res, err := s.repo.Create(student.ID, req)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}

	return c.Status(201).JSON(model.SuccessResponse[*model.AchievementResponse]{Success: true, Data: res})
}

// /api/v1/achievements/:id
func (s *AchievementService) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid AchievementId"})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "You are not authorised to update this achievement"})
	}

	var req model.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid input"})
	}

	if err := s.repo.Update(id, req); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(model.SuccessMessageResponse{Success: true, Message: "Updated"})
}

// /api/v1/achievements/:id
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid AchievementId"})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "You are not authorised to delete this achievement"})
	}

	if err := s.repo.Delete(id); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(model.SuccessMessageResponse{Success: true, Message: "Deleted"})
}

// /api/v1/achievements/:id/submit
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid AchievementId"})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "You are not authorised to submit this achievement"})
	}

	if err := s.repo.UpdateStatus(id, "submitted", nil, "", 0); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(model.SuccessMessageResponse{Success: true, Message: "Submitted for verification"})
}

// /api/v1/achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	if c.Locals("role") != "dosen wali" {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "Advising lecturer only"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid AchievementId"})
	}
	verifierID := c.Locals("user_id").(uuid.UUID)

	// Cek apakah dosen ini berhak (Advisornnya mahasiswa ini)
	if !s.repo.IsAdvisor(verifierID, id) {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "You are not the advisor for this student"})
	}

	var req model.VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid input"})
	}

	if err := s.repo.UpdateStatus(id, "verified", &verifierID, "", req.Points); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(model.SuccessMessageResponse{Success: true, Message: "Verified"})
}

// /api/v1/achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	if c.Locals("role") != "dosen wali" {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "Advising lecturer only"})
	}

	var req model.RejectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid input"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid AchievementId"})
	}
	verifierID := c.Locals("user_id").(uuid.UUID)

	if !s.repo.IsAdvisor(verifierID, id) {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "You are not the advisor for this student"})
	}

	if err := s.repo.UpdateStatus(id, "rejected", &verifierID, req.RejectionNote, 0); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(model.SuccessMessageResponse{Success: true, Message: "Rejected"})
}

// /api/v1/achievements/:id/history
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	return s.Get(c)
}

// /api/v1/achievements/:id/attachments
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	if c.Locals("role") != "mahasiswa" {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "Mahasiswa only"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "Invalid AchievementId"})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: err.Error()})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{Success: false, Message: "You are not authorised to upload attachments for this achievement"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Success: false, Message: "File required"})
	}

	storedFilename := fmt.Sprintf("%s_%s", id.String(), file.Filename)
	uploadDir := "./uploads"
	path := filepath.Join(uploadDir, storedFilename)

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: "Failed to create upload directory"})
	}
	if err := c.SaveFile(file, path); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: "Failed saving file"})
	}

	attachment := model.Attachment{
		FileName:   file.Filename,
		FileURL:    "/uploads/" + storedFilename,
		FileType:   filepath.Ext(file.Filename),
		UploadedAt: time.Now(),
	}

	if err := s.repo.AddAttachment(id, attachment); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Success: false, Message: "Failed updating db"})
	}

	return c.JSON(model.SuccessResponse[*model.Attachment]{Success: true, Data: &attachment})
}
