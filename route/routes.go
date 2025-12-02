package route

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"fiber/skp/app/repo"
	"fiber/skp/app/service"
	"fiber/skp/middleware"
)

func SetupRoutes(app *fiber.App, pgDB *gorm.DB, mongoDB *mongo.Database) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	userRepo := repo.NewUserRepo()
	authService := service.NewAuthService(userRepo)

	auth := v1.Group("/auth")

	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)

	protected := v1.Group("", middleware.AuthRequired())
	
	protected.Get("/auth/profile", authService.Profile)
}
