package main

import (
	"log"
	"login-auth-template/config"
)

func main() {
	// Connect to Postgres
	config.ConnectDB()

	// Quick test: ping and print success
	if config.DB != nil {
		log.Println("DB connection is live!")
	} else {
		log.Println("DB connection is nil")
	}
}
