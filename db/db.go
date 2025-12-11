package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"fiber/skp/config"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB    *sql.DB
	Mongo *mongo.Database
)

func ConnectDB() {
	config.LoadEnv()

	connectPostgres()
	connectMongo()
}

func connectPostgres() {
	dsn := config.GetDBDSN()

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Gagal terhubung ke PostgreSQL:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Gagal ping PostgreSQL:", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Berhasil terhubung ke PostgreSQL")
}

func connectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := config.GetMongoURI()
	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Gagal terhubung ke MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Gagal ping MongoDB:", err)
	}

	dbName := config.GetMongoDB()
	Mongo = client.Database(dbName)
	log.Println("Berhasil terhubung ke MongoDB")
}

func GetDB() *sql.DB {
	return DB
}

func GetMongo() *mongo.Database {
	return Mongo
}
