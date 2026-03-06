package env

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestPatch_ExistingFile(t *testing.T) {
	f, err := os.CreateTemp("", "bight-env-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("DB_NAME=old_db\nAPP_ENV=local\n")
	f.Close()

	if err := Patch(f.Name(), "DB_NAME", "new_db"); err != nil {
		t.Fatalf("Patch: %v", err)
	}

	env, err := godotenv.Read(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if env["DB_NAME"] != "new_db" {
		t.Errorf("DB_NAME = %q, want %q", env["DB_NAME"], "new_db")
	}
	if env["APP_ENV"] != "local" {
		t.Errorf("APP_ENV = %q, want %q", env["APP_ENV"], "local")
	}
}

func TestPatch_NewFile(t *testing.T) {
	path := os.TempDir() + "/bight-test-new.env"
	defer os.Remove(path)

	if err := Patch(path, "SECRET", "abc123"); err != nil {
		t.Fatalf("Patch: %v", err)
	}

	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if env["SECRET"] != "abc123" {
		t.Errorf("SECRET = %q, want %q", env["SECRET"], "abc123")
	}
}
