package service

import (
	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AcademicService struct {
	studentRepo  repo.StudentRepository
	lecturerRepo repo.LecturerRepository
	achieveRepo  repo.AchievementRepository
}

func NewAcademicService(sRepo repo.StudentRepository, lRepo repo.LecturerRepository, aRepo repo.AchievementRepository) *AcademicService {
	return &AcademicService{
		studentRepo:  sRepo,
		lecturerRepo: lRepo,
		achieveRepo:  aRepo,
	}
}

// GET /api/v1/students
func (s *AcademicService) GetAllStudents(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")

	// Validate sortBy
	validSorts := map[string]bool{"created_at": true, "full_name": true, "nim": true, "program_study": true}
	if !validSorts[sortBy] {
		sortBy = "created_at"
	}

	students, total, err := s.studentRepo.FindAll(page, limit, search, sortBy, order)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	var response []model.StudentListResponse
	for _, st := range students {
		advisorName := "-"
		if st.Advisor != nil {
			advisorName = st.Advisor.User.FullName
		}

		response = append(response, model.StudentListResponse{
			ID:           st.ID,
			StudentID:    st.StudentID,
			FullName:     st.User.FullName,
			ProgramStudy: st.ProgramStudy,
			AdvisorName:  advisorName,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(model.SuccessResponse[model.PaginationData[model.StudentListResponse]]{
		Success: true,
		Data: model.PaginationData[model.StudentListResponse]{
			Items: response,
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

// GET /api/v1/students/:id
func (s *AcademicService) GetStudentDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	studentUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "student_id tidak valid",
			Error:   err.Error(),
		})
	}

	st, err := s.studentRepo.FindByID(studentUUID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Mahasiswa tidak ditemukan",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse[model.StudentDetailResponse]{
		Success: true,
		Data: model.StudentDetailResponse{
			ID:           st.ID,
			StudentID:    st.StudentID,
			FullName:     st.User.FullName,
			Email:        st.User.Email,
			ProgramStudy: st.ProgramStudy,
			Advisor:      st.Advisor,
		},
	})
}

// PUT /api/v1/students/:id/advisor
func (s *AcademicService) AssignAdvisor(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "student_id tidak valid",
		})
	}

	var req struct {
		AdvisorID string `json:"lecturer_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid input",
		})
	}

	advisorUUID, err := uuid.Parse(req.AdvisorID)
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "lecturer_id tidak valid",
		})
	}

	_, err = s.lecturerRepo.FindByID(advisorUUID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Lecturer tidak ditemukan",
		})
	}

	if err := s.studentRepo.UpdateAdvisor(studentID, advisorUUID); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal mengassign advisor",
		})
	}

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Advisor berhasil diassign",
	})
}

// GET /api/v1/students/:id/achievements
func (s *AcademicService) GetStudentAchievements(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "student_id tidak valid",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")

	validSorts := map[string]bool{"created_at": true, "updated_at": true, "status": true, "date": true}
	if !validSorts[sortBy] {
		sortBy = "created_at"
	}

	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Mahasiswa tidak ditemukan",
		})
	}

	achievements, total, err := s.achieveRepo.FindAll(model.RoleMahasiswa, student.UserID, page, limit, search, sortBy, order)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal mengambil data achievement",
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(model.SuccessResponse[model.PaginationData[model.AchievementResponse]]{
		Success: true,
		Data: model.PaginationData[model.AchievementResponse]{
			Items: achievements,
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

// GET /api/v1/lecturers
func (s *AcademicService) GetAllLecturers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")

	validSorts := map[string]bool{"created_at": true, "full_name": true, "nip": true, "department": true}
	if !validSorts[sortBy] {
		sortBy = "created_at"
	}

	lecturers, total, err := s.lecturerRepo.FindAll(page, limit, search, sortBy, order)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	var response []model.LecturerListResponse
	for _, l := range lecturers {
		response = append(response, model.LecturerListResponse{
			ID:         l.ID,
			LecturerID: l.LecturerID,
			FullName:   l.User.FullName,
			Department: l.Department,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(model.SuccessResponse[model.PaginationData[model.LecturerListResponse]]{
		Success: true,
		Data: model.PaginationData[model.LecturerListResponse]{
			Items: response,
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

// GET /api/v1/lecturers/:id/advisees
func (s *AcademicService) GetAdvisees(c *fiber.Ctx) error {
	advisorID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "lecturer_id tidak valid",
		})
	}

	lecturer, err := s.lecturerRepo.FindByID(advisorID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "Lecturer tidak ditemukan",
		})
	}

	students, err := s.lecturerRepo.GetAdvisees(advisorID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal mengambil data mahasiswa",
		})
	}

	var response []model.StudentListResponse
	for _, st := range students {
		response = append(response, model.StudentListResponse{
			ID:           st.ID,
			StudentID:    st.StudentID,
			FullName:     st.User.FullName,
			ProgramStudy: st.ProgramStudy,
			AdvisorName:  lecturer.User.FullName,
		})
	}

	return c.JSON(model.SuccessResponse[model.PaginationData[model.StudentListResponse]]{
		Success: true,
		Data: model.PaginationData[model.StudentListResponse]{
			Items: response,
			Meta: model.MetaInfo{
				Page:   1,
				Limit:  10,
				Total:  int64(len(students)),
				Pages:  1,
				SortBy: "created_at",
				Order:  "desc",
				Search: "",
			},
		},
	})
}
