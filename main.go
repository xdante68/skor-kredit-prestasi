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
	config.LoadEnv()
	
	app := config.NewApp()
	
	route.SetupRoutes(app, db.GetDB(), db.GetMongo())

	log.Fatal(app.Listen(":3000"))
}
