package middleware

import (
	"strings"

	"fiber/skp/app/model"
	"fiber/skp/db"
	"fiber/skp/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

		if err := db.GetDB().Where("token = ?", token).First(&blacklistedToken).Error; err == nil {
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

func AdminOnly(c *fiber.Ctx) error {
	role := c.Locals("role")

	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Success: false,
			Message: "Access denied. Admin only.",
		})
	}

	return c.Next()
}
