package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var loadEnvOnce sync.Once

// LoadEnv attempts to load environment variables from a predictable set of .env locations.
// Preference order:
//  1. ENV_PATH (if provided)
//  2. .env in the current working directory
//  3. backend/.env relative to the current working directory
//  4. ../.env (useful when running from backend/)
//  5. Paths relative to the compiled binary's directory
//
// If no files are found the process environment is left untouched.
func LoadEnv() {
	loadEnvOnce.Do(func() {
		candidates := gatherEnvCandidates()
		for _, path := range candidates {
			if path == "" {
				continue
			}
			if _, err := os.Stat(path); err != nil {
				continue
			}

			if err := godotenv.Overload(path); err != nil {
				log.Printf("failed to load env file %s: %v", path, err)
				continue
			}

			log.Printf("loaded environment variables from %s", path)
			return
		}

		log.Println("no .env file loaded; relying on existing environment variables")
	})
}

func gatherEnvCandidates() []string {
	var paths []string

	if custom := os.Getenv("ENV_PATH"); custom != "" {
		paths = append(paths, custom)
	}

	if wd, err := os.Getwd(); err == nil {
		paths = append(paths,
			filepath.Join(wd, ".env"),
			filepath.Join(wd, "backend", ".env"),
			filepath.Join(wd, "..", ".env"),
		)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(exeDir, ".env"),
			filepath.Join(exeDir, "..", ".env"),
		)
	}

	unique := make([]string, 0, len(paths))
	seen := make(map[string]struct{})
	for _, p := range paths {
		if p == "" {
			continue
		}
		abs := p
		if !filepath.IsAbs(abs) {
			if resolved, err := filepath.Abs(abs); err == nil {
				abs = resolved
			}
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		unique = append(unique, abs)
	}

	return unique
}
