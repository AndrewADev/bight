package env

import "github.com/joho/godotenv"

func Patch(path, key, value string) error {
	existing, err := godotenv.Read(path)
	if err != nil {
		// If file doesn't exist, start fresh
		existing = make(map[string]string)
	}
	existing[key] = value
	return godotenv.Write(existing, path)
}
