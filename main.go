package main

import (
	"log"

	"fiber/skp/config"
	"fiber/skp/db"

	"fiber/skp/route"

	"github.com/gofiber/fiber/v2"
)

func main() {
	db.ConnectDB()

	app := config.NewApp()

	route.SetupRoutes(app, db.GetDB(), db.GetMongo())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Listen(":3000"))
}
