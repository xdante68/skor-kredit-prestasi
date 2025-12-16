package service

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"fiber/skp/helper"
)

type AuthService struct {
	repo repo.UserRepository
}

func NewAuthService(repo repo.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags 5.1 Authentication
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials"
// @Success 200 {object} model.LoginSuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /auth/login [post]
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Input tidak valid",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Username dan Password harus diisi",
		})
	}

	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Kredensial tidak valid",
		})
	}

	if !helper.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Kredensial tidak valid",
		})
	}

	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	token, err := helper.GenerateToken(*user, permissions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal menghasilkan token",
		})
	}

	refreshToken, err := helper.GenerateRefreshToken(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal menghasilkan token refresh",
		})
	}

	user.RefreshToken = refreshToken
	if err := s.repo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal menyimpan token refresh",
		})
	}

	return c.JSON(model.LoginSuccessResponse{
		Success: true,
		Message: "Login berhasil",
		Data: model.LoginResponse{
			User: model.LoginUser{
				ID:          user.ID.String(),
				Username:    user.Username,
				FullName:    user.FullName,
				Role:        user.Role.Name,
				Permissions: permissions,
			},
			Token:        token,
			RefreshToken: refreshToken,
		},
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags 5.1 Authentication
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} model.SwaggerRefreshResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /auth/refresh [post]
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Token refresh diperlukan",
		})
	}

	claims, err := helper.ValidateToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Token refresh tidak valid",
		})
	}

	if claims.Type != "refresh" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Token refresh tidak valid",
		})
	}

	user, err := s.repo.FindByUserID(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "User tidak ditemukan",
		})
	}

	if user.RefreshToken != req.RefreshToken {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Token refresh tidak valid",
		})
	}

	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	newToken, err := helper.GenerateToken(*user, permissions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal menghasilkan token",
		})
	}

	return c.JSON(model.SwaggerRefreshResponse{
		Success: true,
		Message: "Token refresh berhasil",
		Data: model.RefreshTokenResponse{
			Token: newToken,
		},
	})
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate JWT tokens
// @Tags 5.1 Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.SuccessMessageResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	bearer := strings.TrimSpace(c.Get("Authorization"))
	if bearer == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Token diperlukan",
		})
	}

	if len(bearer) < 7 || !strings.HasPrefix(strings.ToLower(bearer), "bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Format token tidak valid",
		})
	}

	tokenString := strings.TrimSpace(bearer[7:])

	claims, err := helper.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Token tidak valid",
		})
	}

	blacklistedToken := model.BlacklistedToken{
		Token:     tokenString,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	if err := s.repo.AddBlacklistToken(blacklistedToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal logout",
		})
	}

	var req model.RefreshTokenRequest

	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		refreshClaims, err := helper.ValidateToken(req.RefreshToken)
		if err == nil {
			blacklistedRefreshToken := model.BlacklistedToken{
				Token:     req.RefreshToken,
				ExpiresAt: refreshClaims.ExpiresAt.Time,
			}
			s.repo.AddBlacklistToken(blacklistedRefreshToken)
		}
	}

	if err := s.repo.ClearRefreshToken(claims.UserID); err != nil {
		log.Printf("Gagal menghapus refresh token untuk user %s: %v", claims.UserID, err)

	}

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Logout berhasil",
	})
}

// Profile godoc
// @Summary Get current user profile
// @Description Get profile of authenticated user
// @Tags 5.1 Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.ProfileResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /auth/profile [get]
func (s *AuthService) Profile(c *fiber.Ctx) error {
	var userID string
	switch v := c.Locals("user_id").(type) {
	case string:
		userID = v
	case interface{ String() string }:
		userID = v.String()
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "User tidak ditemukan",
		})
	}

	username, _ := c.Locals("username").(string)
	role, _ := c.Locals("role").(string)

	return c.JSON(model.ProfileResponse{
		Success: true,
		Data: model.ProfileData{
			UserID:   userID,
			Username: username,
			Role:     role,
		},
	})
}
