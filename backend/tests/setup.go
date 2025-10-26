package tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func SetupTestDB(t *testing.T) *sql.DB {
	LoadTestEnv()
	db := ConnectTestDB()
	RunSchema(db, "../db/schema.sql")
	TruncateAll(db)
	return db
}

func LoadTestEnv() {
	if err := godotenv.Load(".env.test"); err != nil {
		log.Println("No .env.test file found (skipping)")
	}
}

// ConnectTestDB opens a connection to the test database
func ConnectTestDB() *sql.DB {
	host := os.Getenv("TEST_DB_HOST")
	port := os.Getenv("TEST_DB_PORT")
	user := os.Getenv("TEST_DB_USER")
	pass := os.Getenv("TEST_DB_PASSWORD")
	name := os.Getenv("TEST_DB_NAME")
	ssl := os.Getenv("TEST_DB_SSLMODE")

	if !strings.Contains(name, "test") { // sanity check we're on test db
		log.Fatal("refusing to connect to non-test database")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, ssl,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open test db: %v", err)
	}

	// wait until DB is ready
	deadline := time.Now().Add(15 * time.Second)
	var err2 error
	for {
		err2 = db.Ping()
		if err2 == nil {
			break
		}
		if time.Now().After(deadline) {
			log.Fatalf("db not ready: %v", err2)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return db
}

// RunSchema runs schema SQL file
func RunSchema(db *sql.DB, path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read schema file: %v", err)
	}
	if _, err := db.Exec(string(b)); err != nil {
		log.Fatalf("failed to run schema: %v", err)
	}
}

// TruncateAll cleans up tables between tests
func TruncateAll(db *sql.DB) {
	if _, err := db.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE;`); err != nil {
		log.Fatalf("failed to truncate tables: %v", err)
	}
}
