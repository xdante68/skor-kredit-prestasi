package route

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"fiber/skp/app/model"
	"fiber/skp/app/repo"
	"fiber/skp/app/service"
	"fiber/skp/middleware"
)

func SetupRoutes(app *fiber.App, pgDB *gorm.DB, mongoDB *mongo.Database) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	userRepo := repo.NewUserRepo()
	studentRepo := repo.NewStudentRepo()
	achievementRepo := repo.NewAchievementRepo(pgDB, mongoDB)

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	achievementSvc := service.NewAchievementService(achievementRepo, studentRepo)

	auth := v1.Group("/auth")

	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)

	protected := v1.Group("", middleware.AuthRequired())

	protected.Get("/auth/profile", authService.Profile)

	users := protected.Group("/users", middleware.RolesRequired(model.RoleAdmin))

	users.Get("/", userService.GetAllUsers)
	users.Get("/:id", userService.GetUser)
	users.Post("/", userService.CreateUser)
	users.Put("/:id", userService.UpdateUser)
	users.Delete("/:id", userService.DeleteUser)
	users.Put("/:id/role", userService.ChangeRole)

	achievements := protected.Group("/achievements")

	achievements.Get("/", achievementSvc.List)
	achievements.Get("/:id", achievementSvc.Get)
	achievements.Post("/", middleware.RolesRequired(model.RoleMahasiswa), achievementSvc.Create)
	achievements.Put("/:id", middleware.RolesRequired(model.RoleMahasiswa), achievementSvc.Update)
	achievements.Delete("/:id", middleware.RolesRequired(model.RoleMahasiswa), achievementSvc.Delete)
	achievements.Post("/:id/submit", middleware.RolesRequired(model.RoleMahasiswa), achievementSvc.Submit)
	achievements.Post("/:id/verify", middleware.RolesRequired(model.RoleDosenWali), achievementSvc.Verify)
	achievements.Post("/:id/reject", middleware.RolesRequired(model.RoleDosenWali), achievementSvc.Reject)
	achievements.Get("/:id/history", achievementSvc.GetHistory)
	achievements.Post("/:id/attachments", middleware.RolesRequired(model.RoleMahasiswa), achievementSvc.UploadAttachment)
}
