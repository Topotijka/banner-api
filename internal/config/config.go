package config

import (
	"github.com/joho/godotenv"
	"os"
)

func DBConnectionString() string {
	if err := godotenv.Load("config.env"); err != nil {
		panic(err)
	}
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbHost := os.Getenv("DATABASE_HOST")
	dbName := os.Getenv("DATABASE_DBNAME")
	dbPort := os.Getenv("DATABASE_PORT")
	connectionString := "user=" + dbUser + " dbname=" + dbName + " sslmode=disable password=" + dbPassword + " host=" + dbHost + " port=" + dbPort
	return connectionString
}
