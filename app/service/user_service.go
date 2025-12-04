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
	userRepo *repo.UserRepo
}

func NewUserService(userRepo *repo.UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
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
			Message: "Failed to fetch users",
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
			Message: "Invalid userId",
			Error:   err.Error(),
		})
	}

	user, err := s.userRepo.FindByID(userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Success: false,
			Message: "User not found",
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
			Message: "Invalid input",
			Error:   err.Error(),
		})
	}

	hashedPwd, err := helper.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
	}

	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid roleId",
			Error:   err.Error(),
		})
	}

	newUser := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPwd,
		FullName:     req.FullName,
		RoleID:       &roleUUID,
	}

	if err := s.userRepo.Create(&newUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "User created successfully",
	})
}

// PUT /api/v1/users/:id
func (s *UserService) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid userId",
			Error:   err.Error(),
		})
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid input",
			Error:   err.Error(),
		})
	}

	user, err := s.userRepo.FindByID(userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Success: false,
			Message: "User not found",
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
				Message: "Failed to hash password",
				Error:   err.Error(),
			})
		}
		user.PasswordHash = hashedPwd
	}

	if err := s.userRepo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "User updated",
	})
}

// DELETE /api/v1/users/:id
func (s *UserService) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid userId",
			Error:   err.Error(),
		})
	}

	if err := s.userRepo.Delete(userUUID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "User deleted",
	})
}

// PUT /api/v1/users/:id/role
func (s *UserService) ChangeRole(c *fiber.Ctx) error {
	id := c.Params("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid userId",
			Error:   err.Error(),
		})
	}

	var req model.ChangeRoleRequest
	if err := c.BodyParser(&req); err != nil || req.RoleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "RoleId required",
		})
	}

	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid roleId",
			Error:   err.Error(),
		})
	}

	if err := s.userRepo.UpdateRole(userUUID, roleUUID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to update role",
			Error:   err.Error(),
		})
	}

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Role updated",
	})
}
