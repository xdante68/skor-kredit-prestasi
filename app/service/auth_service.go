package service

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"fiber/skp/helper"
)

type AuthService struct {
	userRepo *repo.UserRepo
}

func NewAuthService(userRepo *repo.UserRepo) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (ctrl *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid input",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Username and Password are required",
		})
	}

	user, err := ctrl.userRepo.FindByUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	if !helper.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	token, err := helper.GenerateToken(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to generate token",
		})
	}

	refreshToken, err := helper.GenerateRefreshToken(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to generate refresh token",
		})
	}

	// Save refresh token to user
	user.RefreshToken = refreshToken
	if err := ctrl.userRepo.DB.Save(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to save refresh token",
		})
	}

	return c.JSON(model.LoginSuccessResponse{
		Success: true,
		Message: "Login successful",
		Data: model.LoginResponse{
			User: model.LoginUser{
				ID:        user.ID.String(),
				Username:  user.Username,
				Email:     user.Email,
				Role:      user.Role.Name,
				CreatedAt: user.CreatedAt.Format(time.RFC3339),
			},
			Token:        token,
			RefreshToken: refreshToken,
		},
	})
}

func (ctrl *AuthService) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "Refresh token required",
		})
	}

	claims, err := helper.ValidateToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid refresh token",
		})
	}

	if claims.Type != "refresh" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid token type",
		})
	}

	// Find user to check stored refresh token
	user, err := ctrl.userRepo.FindByID(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "User not found",
		})
	}

	if user.RefreshToken != req.RefreshToken {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid refresh token",
		})
	}

	newToken, err := helper.GenerateToken(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Token refreshed",
		"data": fiber.Map{
			"token": newToken,
		},
	})
}

func (ctrl *AuthService) Logout(c *fiber.Ctx) error {
	bearer := strings.TrimSpace(c.Get("Authorization"))
	if bearer == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Token required",
		})
	}

	tokenString := strings.TrimSpace(bearer[7:])

	// Get expiration time from token
	claims, err := helper.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Success: false,
			Message: "Invalid token",
		})
	}

	// Save access token to blacklist
	blacklistedToken := model.BlacklistedToken{
		Token:     tokenString,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	if err := ctrl.userRepo.DB.Create(&blacklistedToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "Failed to logout",
		})
	}

	// Handle refresh token if provided
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		// Validate refresh token to get expiration
		refreshClaims, err := helper.ValidateToken(req.RefreshToken)
		if err == nil {
			// Add refresh token to blacklist
			blacklistedRefreshToken := model.BlacklistedToken{
				Token:     req.RefreshToken,
				ExpiresAt: refreshClaims.ExpiresAt.Time,
			}
			ctrl.userRepo.DB.Create(&blacklistedRefreshToken)
		}
	}

	// Clear refresh token from user record
	if err := ctrl.userRepo.DB.Model(&model.User{}).Where("id = ?", claims.UserID).Update("refresh_token", "").Error; err != nil {
		// Log error but don't fail the request since token is already blacklisted
		// log.Println("Failed to clear refresh token from user:", err)
	}

	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "Successfully logged out",
	})
}

func (ctrl *AuthService) Profile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		// Try to get from user object if direct string fails (depends on middleware)
		// For now, assume middleware sets it correctly or handle error
		// If middleware sets uuid.UUID, we need to handle that
	}

	if uid, ok := c.Locals("user_id").(string); ok {
		userID = uid
	} else if uid, ok := c.Locals("user_id").(interface{ String() string }); ok {
		userID = uid.String()
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
