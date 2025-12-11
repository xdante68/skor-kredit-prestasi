package service

import (
	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"fiber/skp/helper"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
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

// GET /api/v1/achievements
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

// GET /api/v1/achievements/:id
func (s *AchievementService) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement_id tidak valid",
		})
	}

	data, err := s.repo.FindByAchievementID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Achievement tidak ditemukan",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	if role == model.RoleMahasiswa {
		ownerID, err := s.repo.GetOwnerID(id)
		if err != nil || ownerID != userID {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "Anda tidak berhak melihat achievement orang lain.",
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
				Message: "Anda tidak berhak melihat achievement mahasiswa yang bukan bimbingan Anda.",
			})
		}
	}

	return c.JSON(model.SuccessResponse[*model.AchievementResponse]{
		Success: true,
		Data:    data,
	})
}

// POST /api/v1/achievements
func (s *AchievementService) Create(c *fiber.Ctx) error {

	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
		})
	}

	if err := helper.ValidateStruct(req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Validasi gagal",
			Error:   helper.FormatValidationErrors(err),
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Profil mahasiswa tidak ditemukan",
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

// PUT /api/v1/achievements/:id
func (s *AchievementService) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement_id tidak valid",
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
			Message: "Anda tidak berhak mengubah achievement yang bukan milik Anda",
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
			Message: "Hanya achievement dengan status 'draft' atau 'rejected' yang dapat diubah",
		})
	}

	var req model.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
		})
	}

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
		Message: "achievement berhasil diubah",
	})
}

// DELETE /api/v1/achievements/:id
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement_id tidak valid",
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
			Message: "Anda tidak berhak menghapus achievement yang bukan milik Anda",
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
			Message: "Hanya achievement dengan status 'draft' yang dapat dihapus",
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
		Message: "achievement berhasil dihapus",
	})
}

// PUT /api/v1/achievements/:id/submit
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false, Message: "achievement_id tidak valid",
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
			Message: "Anda tidak berhak mengajukan achievement yang bukan milik Anda",
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
			Message: "Hanya achievement dengan status 'draft' yang dapat disubmit",
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
		Message: "Achievement berhasil disubmit",
	})
}

// PUT /api/v1/achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement_id tidak valid",
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
			Message: "Anda bukan dosen wali dari mahasiswa ini",
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
			Message: "Achievement harus disubmit sebelum diverifikasi",
		})
	}

	var req model.VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
		})
	}

	if req.Points <= 0 {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Points harus lebih dari 0",
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
		Message: "Achievement berhasil diverifikasi",
	})
}

// PUT /api/v1/achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {

	var req model.RejectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
		})
	}

	if strings.TrimSpace(req.RejectionNote) == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Catatan penolakan (rejection_note) wajib diisi",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement_id tidak valid",
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
			Message: "Anda bukan dosen wali dari mahasiswa ini",
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
			Message: "Achievement harus disubmit sebelum ditolak",
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
		Message: "Achievement berhasil ditolak",
	})
}

// GET /api/v1/achievements/:id/history
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement_id tidak valid",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	if role == model.RoleMahasiswa {
		ownerID, err := s.repo.GetOwnerID(id)
		if err != nil || ownerID != userID {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "Anda tidak berhak melihat history achievement ini",
			})
		}
	} else if role == model.RoleDosenWali {
		isAdvisor, err := s.repo.IsAdvisor(userID, id)
		if err != nil || !isAdvisor {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "Anda bukan dosen wali dari mahasiswa ini",
			})
		}
	}

	history, err := s.repo.GetHistory(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement tidak ditemukan",
		})
	}

	return c.JSON(model.SuccessResponse[*model.AchievementHistoryResponse]{
		Success: true,
		Data:    history,
	})
}

// POST /api/v1/achievements/:id/attachments
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "achievement_id tidak valid",
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
			Message: "Anda tidak berhak mengunggah attachment untuk achievement ini",
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
			Message: "Hanya achievement dengan status 'draft' yang dapat mengunggah attachment",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "File wajib diisi",
		})
	}

	const maxFileSize = 5 * 1024 * 1024
	if file.Size > maxFileSize {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Ukuran file maksimal 5MB",
		})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Hanya file PDF yang diizinkan",
		})
	}

	storedFilename := fmt.Sprintf("%s_%s", id.String(), file.Filename)
	uploadDir := "./uploads"
	path := filepath.Join(uploadDir, storedFilename)

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal membuat direktori uploads",
		})
	}
	if err := c.SaveFile(file, path); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal menyimpan file",
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
			Message: "Gagal memperbarui database",
		})
	}

	return c.JSON(model.SuccessResponse[*model.Attachment]{
		Success: true,
		Message: "Attachment berhasil diunggah",
		Data:    &attachment,
	})
}
