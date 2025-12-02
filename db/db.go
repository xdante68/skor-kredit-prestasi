package db

import (
	"context"
	"fiber/skp/app/model"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	Mongo *mongo.Database
)

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	connectPostgres()
	connectMongo()
}

func connectPostgres() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set in .env")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	// AutoMigrate models
	err = DB.AutoMigrate(
		&model.User{},
		&model.Student{},              // Butuh User & Lecturer (Advisor)
		&model.AchievementReference{}, // Butuh Student & User
		&model.BlacklistedToken{},
	)
	if err != nil {
		log.Fatal("Failed to auto migrate database schema:", err)
	}

	log.Println(" Connected to PostgreSQL successfully")
}

func connectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	Mongo = client.Database(dbName)

	log.Println("Connected to MongoDB successfully")
}

func GetDB() *gorm.DB {
	return DB
}

func GetMongo() *mongo.Database {
	return Mongo
}
