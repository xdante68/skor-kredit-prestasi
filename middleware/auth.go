package middleware

import (
	"strings"

	"fiber/skp/app/model"
	"fiber/skp/db"
	"fiber/skp/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearer := strings.TrimSpace(c.Get("Authorization"))
		if bearer == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Token not found",
			})
		}

		if len(bearer) < 7 || !strings.EqualFold(bearer[:7], "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Bearer format wrong",
			})
		}
		token := strings.TrimSpace(bearer[7:])

		claims, err := helper.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Token invalid",
			})
		}

		if claims.Type != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Token type invalid",
			})
		}

		var blacklistedToken model.BlacklistedToken

		if err := db.GetDB().Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Where("token = ?", token).First(&blacklistedToken).Error; err == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Token blacklisted",
			})
		}

		if claims == nil || claims.UserID == uuid.Nil || claims.Username == "" || claims.Role == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Token claim incomplete",
			})
		}

		role := strings.ToLower(claims.Role)

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", role)
		c.Locals("user", claims)

		return c.Next()
	}
}

func PermissionsRequired(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user")
		if claims == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "User claims not found",
			})
		}

		jwtClaims, ok := claims.(*model.JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Invalid claims format",
			})
		}

		userPermissions := jwtClaims.Permissions

		for _, required := range requiredPermissions {
			for _, userPerm := range userPermissions {
				if required == userPerm {
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Success: false,
			Message: "Forbidden access",
		})
	}
}
