package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func LoadEnv(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf(".env file not found, using system environment variables")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}
}
