package main

import (
	"log"

	"fiber/skp/config"
	"fiber/skp/db"
	"fiber/skp/route"

	_ "fiber/skp/docs"
)

// @title SKP API
// @version 1.0
// @description API untuk Sistem Kredit Prestasi Mahasiswa
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@skp.ac.id

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name 5.1 Authentication
// @tag.description Endpoint untuk autentikasi (login, logout, refresh token, profile)

// @tag.name 5.2 Users (Admin)
// @tag.description Endpoint manajemen user (Admin only) - CRUD dan change role

// @tag.name 5.3 Achievements
// @tag.description Endpoint manajemen prestasi mahasiswa - CRUD, submit, verify, reject, history, upload

// @tag.name 5.4 Students & Lecturers
// @tag.description Endpoint manajemen mahasiswa, dosen, dan assign dosen wali (Admin only)

// @tag.name 5.5 Reports & Analytics
// @tag.description Endpoint statistik dan laporan prestasi

func main() {
	db.ConnectDB()
	config.Logger()

	app := config.NewApp()

	route.SetupRoutes(app, db.GetDB(), db.GetMongo())

	port := ":" + config.GetAppPort()
	log.Fatal(app.Listen(port))
}
