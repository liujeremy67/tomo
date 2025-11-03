package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// DB is the global database connection
var DB *sql.DB

func ConnectDB() {
	LoadEnv()

	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	user := strings.TrimSpace(os.Getenv("DB_USER"))
	password := strings.TrimSpace(os.Getenv("DB_PASS"))
	dbname := strings.TrimSpace(os.Getenv("DB_NAME"))

	ensureEnv("DB_HOST", host)
	ensureEnv("DB_PORT", port)
	ensureEnv("DB_USER", user)
	ensureEnv("DB_PASS", password)
	ensureEnv("DB_NAME", dbname)

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	log.Println("Database connected successfully")
}

func ensureEnv(key, value string) {
	if value == "" {
		log.Fatalf("environment variable %s is required but not set", key)
	}
}
