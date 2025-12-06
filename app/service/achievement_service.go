package service

import (
	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AchievementService struct {
	repo         repo.AchievementRepository
	studentRepo  repo.StudentRepository
	lecturerRepo repo.LecturerRepository
}

func NewAchievementService(repo repo.AchievementRepository, studentRepo repo.StudentRepository, lecturerRepo repo.LecturerRepository) *AchievementService {
	return &AchievementService{
		repo:         repo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// /api/v1/achievements
func (s *AchievementService) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")
	validSorts := map[string]bool{"created_at": true, "updated_at": true, "status": true, "date": true}
	
	if !validSorts[sortBy] {
		sortBy = "created_at"
	}

	data, total, err := s.repo.FindAll(role, userID, page, limit, search, sortBy, order)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(model.SuccessResponse[model.PaginationData[model.AchievementResponse]]{
		Success: true,
		Data: model.PaginationData[model.AchievementResponse]{
			Items: data,
			Meta: model.MetaInfo{
				Page:   page,
				Limit:  limit,
				Total:  total,
				Pages:  totalPages,
				SortBy: sortBy,
				Order:  order,
				Search: search,
			},
		},
	})
}

// /api/v1/achievements/:id
func (s *AchievementService) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid achievement_id",
		})
	}

	data, err := s.repo.FindByAchievementID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Achievement not found",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	if role == model.RoleMahasiswa {
		ownerID, err := s.repo.GetOwnerID(id)
		if err != nil || ownerID != userID {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "You are not authorised to view other people's achievements.",
			})
		}
	} else if role == model.RoleDosenWali {
		isAdvisor, err := s.repo.IsAdvisor(userID, id)
		if err != nil {
			return c.Status(500).JSON(model.ErrorResponse{
				Success: false,
				Message: err.Error(),
			})
		}
		if !isAdvisor {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "You are not the advisor for this student",
			})
		}
	}

	return c.JSON(model.SuccessResponse[*model.AchievementResponse]{
		Success: true,
		Data:    data,
	})
}

// /api/v1/achievements
func (s *AchievementService) Create(c *fiber.Ctx) error {

	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid input",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Student profile not found",
		})
	}

	res, err := s.repo.Create(student.ID, req)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(201).JSON(model.SuccessResponse[*model.AchievementResponse]{
		Success: true,
		Data:    res,
	})
}

// /api/v1/achievements/:id
func (s *AchievementService) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid achievement_id",
		})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{
			Success: false,
			Message: "You are not authorised to update this achievement",
		})
	}

	// Check status - can only update draft or rejected
	currentStatus, err := s.repo.GetStatus(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false, Message: err.Error(),
		})
	}
	if currentStatus != "draft" && currentStatus != "rejected" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Can only update achievements with status 'draft' or 'rejected'",
		})
	}

	var req model.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid input",
		})
	}

	// If rejected, reset to draft when updating
	if currentStatus == "rejected" {
		if err := s.repo.UpdateStatus(id, "draft", nil, "", 0); err != nil {
			return c.Status(500).JSON(model.ErrorResponse{
				Success: false,
				Message: err.Error(),
			})
		}
	}

	_, err = s.repo.Update(id, req)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Updated Successfully",
	})
}

// /api/v1/achievements/:id
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid achievement_id",
		})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{
			Success: false,
			Message: "You are not authorised to delete this achievement",
		})
	}

	// Check status - can only delete draft
	currentStatus, err := s.repo.GetStatus(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false, Message: err.Error(),
		})
	}
	if currentStatus != "draft" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Can only delete achievements with status 'draft'",
		})
	}

	if err := s.repo.Delete(id); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Deleted Successfully",
	})
}

// /api/v1/achievements/:id/submit
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false, Message: "Invalid achievement_id",
		})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{
			Success: false,
			Message: "You are not authorised to submit this achievement",
		})
	}

	currentStatus, err := s.repo.GetStatus(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false, Message: err.Error(),
		})
	}
	if currentStatus != "draft" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Can only submit achievements with status 'draft'",
		})
	}

	if err := s.repo.UpdateStatus(id, "submitted", nil, "", 0); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Submitted Successfully",
	})
}

// /api/v1/achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid achievement_id",
		})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	isAdvisor, err := s.repo.IsAdvisor(userID, id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	if !isAdvisor {
		return c.Status(403).JSON(model.ErrorResponse{
			Success: false,
			Message: "You are not the advisor for this student",
		})
	}

	currentStatus, err := s.repo.GetStatus(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false, Message: err.Error(),
		})
	}
	if currentStatus != "submitted" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Achievement must be submitted before verification",
		})
	}

	var req model.VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid input",
		})
	}

	if req.Points <= 0 {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Points must be greater than 0",
		})
	}

	if err := s.repo.UpdateStatus(id, "verified", &userID, "", req.Points); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Verified Successfully",
	})
}

// /api/v1/achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {

	var req model.RejectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid input",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid achievement_id",
		})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	isAdvisor, err := s.repo.IsAdvisor(userID, id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if !isAdvisor {
		return c.Status(403).JSON(model.ErrorResponse{
			Success: false,
			Message: "You are not the advisor for this student",
		})
	}

	currentStatus, err := s.repo.GetStatus(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if currentStatus != "submitted" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Achievement must be submitted before rejection",
		})
	}

	if err := s.repo.UpdateStatus(id, "rejected", &userID, req.RejectionNote, 0); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Rejected Successfully",
	})
}

// /api/v1/achievements/:id/history
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid achievement_id",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	if role == model.RoleMahasiswa {
		ownerID, err := s.repo.GetOwnerID(id)
		if err != nil || ownerID != userID {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "You are not authorised to view this history.",
			})
		}
	} else if role == model.RoleDosenWali {
		isAdvisor, err := s.repo.IsAdvisor(userID, id)
		if err != nil || !isAdvisor {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "You are not the advisor for this student",
			})
		}
	}

	history, err := s.repo.GetHistory(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Achievement not found",
		})
	}

	return c.JSON(model.SuccessResponse[*model.AchievementHistoryResponse]{
		Success: true,
		Data:    history,
	})
}

// /api/v1/achievements/:id/attachments
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid achievement_id",
		})
	}
	userID := c.Locals("user_id").(uuid.UUID)

	ownerID, err := s.repo.GetOwnerID(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if ownerID != userID {
		return c.Status(403).JSON(model.ErrorResponse{
			Success: false,
			Message: "You are not authorised to upload attachments for this achievement",
		})
	}

	currentStatus, err := s.repo.GetStatus(id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if currentStatus != "draft" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Can only upload attachments for achievements with status 'draft'",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "File required",
		})
	}

	storedFilename := fmt.Sprintf("%s_%s", id.String(), file.Filename)
	uploadDir := "./uploads"
	path := filepath.Join(uploadDir, storedFilename)

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to create upload directory",
		})
	}
	if err := c.SaveFile(file, path); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed saving file",
		})
	}

	attachment := model.Attachment{
		FileName:   file.Filename,
		FileURL:    "/uploads/" + storedFilename,
		FileType:   filepath.Ext(file.Filename),
		UploadedAt: time.Now(),
	}

	if err := s.repo.AddAttachment(id, attachment); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed updating db",
		})
	}

	return c.JSON(model.SuccessResponse[*model.Attachment]{
		Success: true,
		Message: "Attachment uploaded successfully",
		Data:    &attachment,
	})
}
