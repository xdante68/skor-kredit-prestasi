package service

import (
	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"fiber/skp/helper"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo     repo.UserRepository
	studentRepo  repo.StudentRepository
	lecturerRepo repo.LecturerRepository
}

func NewUserService(userRepo repo.UserRepository, studentRepo repo.StudentRepository, lecturerRepo repo.LecturerRepository) *UserService {
	return &UserService{
		userRepo:     userRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// GET /api/v1/users
func (s *UserService) GetAllUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")

	validSorts := map[string]bool{"username": true, "email": true, "full_name": true, "created_at": true}
	if !validSorts[sortBy] {
		sortBy = "created_at"
	}

	users, total, err := s.userRepo.FindAll(page, limit, search, sortBy, order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal memuat data user",
			Error:   err.Error(),
		})
	}

	var userResponses []model.UserResponse
	for _, u := range users {
		userResponses = append(userResponses, model.UserResponse{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			FullName: u.FullName,
			Role:     u.Role.Name,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(model.SuccessResponse[model.PaginationData[model.UserResponse]]{
		Success: true,
		Data: model.PaginationData[model.UserResponse]{
			Items: userResponses,
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

// GET /api/v1/users/:id
func (s *UserService) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "user_id tidak valid",
			Error:   err.Error(),
		})
	}

	user, err := s.userRepo.FindByUserIDSimple(userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Success: false,
			Message: "User tidak ditemukan",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse[model.UserResponse]{
		Success: true,
		Data: model.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role.Name,
		},
	})
}

// POST /api/v1/users
func (s *UserService) CreateUser(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
			Error:   err.Error(),
		})
	}

	if err := helper.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Validasi gagal",
			Error:   helper.FormatValidationErrors(err),
		})
	}

	if req.Role == model.RoleMahasiswa {
		if req.Student == nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Data khusus mahasiswa diperlukan untuk role mahasiswa",
			})
		}
		if err := helper.ValidateStruct(*req.Student); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Validasi data mahasiswa gagal",
				Error:   helper.FormatValidationErrors(err),
			})
		}
	} else if req.Role == model.RoleDosenWali {
		if req.Lecturer == nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Data khusus dosen wali diperlukan untuk role dosen_wali",
			})
		}
		if err := helper.ValidateStruct(*req.Lecturer); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Validasi data dosen wali gagal",
				Error:   helper.FormatValidationErrors(err),
			})
		}
	}

	roleData, err := s.userRepo.FindRoleByName(req.Role)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Role tidak valid: " + req.Role,
			Error:   err.Error(),
		})
	}

	if roleData.Name == model.RoleMahasiswa {
		if exists, _ := s.studentRepo.ExistsByStudentID(req.Username); exists {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Username (student_id) sudah terdaftar",
				Error:   "Data duplicated",
			})
		}
	} else if roleData.Name == model.RoleDosenWali {
		if exists, _ := s.lecturerRepo.ExistsByLecturerID(req.Username); exists {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Username (lecturer_id) sudah terdaftar",
				Error:   "Data duplicated",
			})
		}
	}

	hashedPwd, err := helper.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal menghash password",
			Error:   err.Error(),
		})
	}

	newUser := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPwd,
		FullName:     req.FullName,
		RoleID:       &roleData.ID,
	}

	if err := s.userRepo.Create(&newUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal membuat user",
			Error:   err.Error(),
		})
	}

	if roleData.Name == model.RoleMahasiswa {
		student := model.Student{
			UserID:       newUser.ID,
			StudentID:    req.Username,
			ProgramStudy: req.Student.ProgramStudy,
			AcademicYear: req.Student.AcademicYear,
		}
		if err := s.studentRepo.Create(&student); err != nil {
			_ = s.userRepo.Delete(newUser.ID)
			return c.Status(500).JSON(model.ErrorResponse{
				Success: false,
				Message: "User berhasil dibuat tetapi gagal membuat profile mahasiswa. Rolled back.",
				Error:   err.Error(),
			})
		}
	} else if roleData.Name == model.RoleDosenWali {
		lecturer := model.Lecturer{
			UserID:     newUser.ID,
			LecturerID: req.Username,
			Department: req.Lecturer.Department,
		}
		if err := s.lecturerRepo.Create(&lecturer); err != nil {
			_ = s.userRepo.Delete(newUser.ID)
			return c.Status(500).JSON(model.ErrorResponse{
				Success: false,
				Message: "User berhasil dibuat tetapi gagal membuat profile dosen wali. Rolled back.",
				Error:   err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "User berhasil dibuat",
	})
}

// PUT /api/v1/users/:id
func (s *UserService) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "user_id tidak valid",
			Error:   err.Error(),
		})
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
			Error:   err.Error(),
		})
	}

	if err := helper.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Validasi gagal",
			Error:   helper.FormatValidationErrors(err),
		})
	}

	user, err := s.userRepo.FindByUserID(userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Success: false,
			Message: "User tidak ditemukan",
			Error:   err.Error(),
		})
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Password != "" {
		hashedPwd, err := helper.HashPassword(req.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Success: false,
				Message: "Gagal menghash password",
				Error:   err.Error(),
			})
		}
		user.PasswordHash = hashedPwd
	}

	if err := s.userRepo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal mengupdate user",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "User berhasil diupdate",
	})
}

// DELETE /api/v1/users/:id
func (s *UserService) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "user_id tidak valid",
			Error:   err.Error(),
		})
	}

	if err := s.userRepo.Delete(userUUID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal menghapus user",
			Error:   err.Error(),
		})
	}

	_ = s.studentRepo.DeleteByUserID(userUUID)
	_ = s.lecturerRepo.DeleteByUserID(userUUID)

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "User berhasil dihapus",
	})
}

// PUT /api/v1/users/:id/role
func (s *UserService) ChangeRole(c *fiber.Ctx) error {
	id := c.Params("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "user_id tidak valid",
			Error:   err.Error(),
		})
	}

	var req model.ChangeRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
			Error:   err.Error(),
		})
	}

	if err := helper.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Validasi gagal",
			Error:   helper.FormatValidationErrors(err),
		})
	}

	if req.Role == model.RoleMahasiswa {
		if req.Student == nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Data khusus mahasiswa diperlukan untuk role mahasiswa",
			})
		}
		if err := helper.ValidateStruct(*req.Student); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Validasi data mahasiswa gagal",
				Error:   helper.FormatValidationErrors(err),
			})
		}
	} else if req.Role == model.RoleDosenWali {
		if req.Lecturer == nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Data khusus dosen wali diperlukan untuk role dosen_wali",
			})
		}
		if err := helper.ValidateStruct(*req.Lecturer); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Success: false,
				Message: "Validasi data dosen wali gagal",
				Error:   helper.FormatValidationErrors(err),
			})
		}
	}

	roleData, err := s.userRepo.FindRoleByName(req.Role)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Role tidak valid: " + req.Role,
			Error:   err.Error(),
		})
	}

	user, err := s.userRepo.FindByUserID(userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Success: false,
			Message: "User tidak ditemukan",
			Error:   err.Error(),
		})
	}

	if err := s.userRepo.UpdateRole(userUUID, roleData.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal mengupdate role",
			Error:   err.Error(),
		})
	}

	_ = s.studentRepo.DeleteByUserID(user.ID)
	_ = s.lecturerRepo.DeleteByUserID(user.ID)

	if roleData.Name == model.RoleMahasiswa {
		student := model.Student{
			UserID:       user.ID,
			StudentID:    user.Username,
			ProgramStudy: req.Student.ProgramStudy,
			AcademicYear: req.Student.AcademicYear,
		}
		if err := s.studentRepo.Create(&student); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Success: false,
				Message: "Role berhasil diupdate tetapi gagal membuat profile mahasiswa",
				Error:   err.Error(),
			})
		}
	} else if roleData.Name == model.RoleDosenWali {
		lecturer := model.Lecturer{
			UserID:     user.ID,
			LecturerID: user.Username,
			Department: req.Lecturer.Department,
		}
		if err := s.lecturerRepo.Create(&lecturer); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Success: false,
				Message: "Role berhasil diupdate tetapi gagal membuat profile dosen wali",
				Error:   err.Error(),
			})
		}
	}

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Role berhasil diupdate",
	})
}
