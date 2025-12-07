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
				Message: "Token tidak ditemukan",
			})
		}

		if len(bearer) < 7 || !strings.EqualFold(bearer[:7], "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Format (Bearer) token tidak valid",
			})
		}
		token := strings.TrimSpace(bearer[7:])

		claims, err := helper.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Token tidak valid",
			})
		}

		if claims.Type != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Tipe token tidak valid",
			})
		}

		var exists bool
		blacklistErr := db.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token = $1)", token).Scan(&exists)
		if blacklistErr == nil && exists {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Token telah di blacklist",
			})
		}

		if claims == nil || claims.UserID == uuid.Nil || claims.Username == "" || claims.Role == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Claim token tidak lengkap",
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
				Message: "Claim user tidak ditemukan",
			})
		}

		jwtClaims, ok := claims.(*model.JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Success: false,
				Message: "Format claim tidak valid",
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
			Message: "Akses dilarang",
		})
	}
}
