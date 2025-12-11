package route

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"fiber/skp/app/repo"
	"fiber/skp/app/service"
	"fiber/skp/middleware"
)

func SetupRoutes(app *fiber.App, pgDB *sql.DB, mongoDB *mongo.Database) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	userRepo := repo.NewUserRepo(pgDB)
	studentRepo := repo.NewStudentRepo(pgDB)
	lecturerRepo := repo.NewLecturerRepo(pgDB)
	achievementRepo := repo.NewAchievementRepo(pgDB, mongoDB)

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo, studentRepo, lecturerRepo)
	academicService := service.NewAcademicService(studentRepo, lecturerRepo, achievementRepo)
	achievementSvc := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo)

	auth := v1.Group("/auth")

	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)

	protected := v1.Group("", middleware.AuthRequired())

	protected.Get("/auth/profile", authService.Profile)

	// User endpoint (Admin only)
	users := protected.Group("/users", middleware.PermissionsRequired("user:manage"))

	users.Get("/", userService.GetAllUsers)
	users.Get("/:id", userService.GetUser)
	users.Post("/", userService.CreateUser)
	users.Put("/:id", userService.UpdateUser)
	users.Delete("/:id", userService.DeleteUser)
	users.Put("/:id/role", userService.ChangeRole)

	// Students endpoint (Admin only)
	students := protected.Group("/students", middleware.PermissionsRequired("user:manage"))
	students.Get("/", academicService.GetAllStudents)
	students.Get("/:id", academicService.GetStudentDetail)
	students.Get("/:id/achievements", academicService.GetStudentAchievements)
	students.Put("/:id/advisor", academicService.AssignAdvisor)

	// Lecturers endpoint (Admin only)
	lecturers := protected.Group("/lecturers", middleware.PermissionsRequired("user:manage"))
	lecturers.Get("/", academicService.GetAllLecturers)
	lecturers.Get("/:id/advisees", academicService.GetAdvisees)

	// Achievements endpoint
	achievements := protected.Group("/achievements")

	achievements.Get("/", achievementSvc.List)
	achievements.Get("/:id", achievementSvc.Get)
	achievements.Post("/", middleware.PermissionsRequired("achievement:create"), achievementSvc.Create)
	achievements.Put("/:id", middleware.PermissionsRequired("achievement:update"), achievementSvc.Update)
	achievements.Delete("/:id", middleware.PermissionsRequired("achievement:delete"), achievementSvc.Delete)
	achievements.Post("/:id/submit", middleware.PermissionsRequired("achievement:create"), achievementSvc.Submit)
	achievements.Post("/:id/verify", middleware.PermissionsRequired("achievement:verify"), achievementSvc.Verify)
	achievements.Post("/:id/reject", middleware.PermissionsRequired("achievement:verify"), achievementSvc.Reject)
	achievements.Get("/:id/history", achievementSvc.GetHistory)
	achievements.Post("/:id/attachments", middleware.PermissionsRequired("achievement:create"), achievementSvc.UploadAttachment)
}
