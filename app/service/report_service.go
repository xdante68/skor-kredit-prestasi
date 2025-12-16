package service

import (
	"fiber/skp/app/model"
	"fiber/skp/app/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReportService struct {
	repo        repo.ReportRepository
	studentRepo repo.StudentRepository
}

func NewReportService(repo repo.ReportRepository, studentRepo repo.StudentRepository) *ReportService {
	return &ReportService{
		repo:        repo,
		studentRepo: studentRepo,
	}
}

// GetStatistics godoc
// @Summary Get statistics
// @Description Get achievement statistics based on user role
// @Tags 5.5 Reports & Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.SwaggerStatsResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /reports/statistics [get]
func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	mongoIDs, err := s.repo.GetMongoIDs(role, userID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal mengambil data achievement",
			Error:   err.Error(),
		})
	}

	byType, _ := s.repo.GetStatsByType(mongoIDs)
	byLevel, _ := s.repo.GetStatsByLevel(mongoIDs)
	byPeriod, _ := s.repo.GetStatsByPeriod(role, userID)

	var topStudents []model.TopStudent
	if role != model.RoleMahasiswa {
		topStudents, _ = s.repo.GetTopStudents()
	}

	totalVerified := int64(len(mongoIDs))

	return c.JSON(model.SwaggerStatsResponse{
		Success: true,
		Data: model.StatsResponse{
			TotalAchievements: totalVerified,
			ByType:            byType,
			ByLevel:           byLevel,
			ByPeriod:          byPeriod,
			TopStudents:       topStudents,
		},
	})
}

// GetStudentStats godoc
// @Summary Get student statistics
// @Description Get detailed statistics for a specific student
// @Tags 5.5 Reports & Analytics
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {object} model.SwaggerStudentStatsResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Router /reports/student/{id} [get]
func (s *ReportService) GetStudentStats(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "student_id tidak valid",
		})
	}

	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(uuid.UUID)

	if role == model.RoleMahasiswa {
		student, err := s.studentRepo.FindByUserID(userID)
		if err != nil || student.ID != studentID {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "Anda tidak memiliki akses ke statistik mahasiswa ini",
			})
		}
	} else if role == model.RoleDosenWali {
		student, err := s.studentRepo.FindByID(studentID)
		if err != nil {
			return c.Status(404).JSON(model.ErrorResponse{
				Success: false,
				Message: "Mahasiswa tidak ditemukan",
			})
		}
		if student.AdvisorID == nil {
			return c.Status(403).JSON(model.ErrorResponse{
				Success: false,
				Message: "Mahasiswa ini tidak memiliki dosen pembimbing",
			})
		}
	}

	profile, err := s.repo.GetStudentProfile(studentID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Mahasiswa tidak ditemukan",
		})
	}

	mongoIDs, err := s.repo.GetMongoIDsByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal mengambil data statistik",
		})
	}

	byType, _ := s.repo.GetStatsByType(mongoIDs)
	byLevel, _ := s.repo.GetStatsByLevel(mongoIDs)

	return c.JSON(model.SwaggerStudentStatsResponse{
		Success: true,
		Data: model.StudentStatsResponse{
			StudentProfile: *profile,
			Stats: model.StatsResponse{
				TotalAchievements: int64(len(mongoIDs)),
				ByType:            byType,
				ByLevel:           byLevel,
			},
		},
	})
}
