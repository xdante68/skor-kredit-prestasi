package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	AppPort   string
	DBDSN     string
	MongoURI  string
	MongoDB   string
	JWTSecret string
}

var Env EnvConfig

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	Env.AppPort = os.Getenv("APP_PORT")
	Env.DBDSN = os.Getenv("DB_DSN")
	Env.MongoURI = os.Getenv("MONGO_URI")
	Env.MongoDB = os.Getenv("MONGO_DB_NAME")
	Env.JWTSecret = os.Getenv("JWT_SECRET")
}

func GetDBDSN() string {
	return Env.DBDSN
}
func GetMongoURI() string {
	return Env.MongoURI
}

func GetMongoDB() string {
	return Env.MongoDB
}
func GetJWTSecret() string {
	return Env.JWTSecret
}

func GetAppPort() string {
	if Env.AppPort == "" {
		return "3000"
	}
	return Env.AppPort
}
