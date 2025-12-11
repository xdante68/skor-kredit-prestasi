package main

import (
	"log"

	"fiber/skp/config"
	"fiber/skp/db"
	"fiber/skp/route"
)

func main() {
	db.ConnectDB()
	config.Logger()

	app := config.NewApp()

	route.SetupRoutes(app, db.GetDB(), db.GetMongo())

	port := ":" + config.GetAppPort()
	log.Fatal(app.Listen(port))
}
