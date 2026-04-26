package env

import (
	"os"
	"path/filepath"
	"strings"
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

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "bight-env-*")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestScanComments_All(t *testing.T) {
	path := writeTempEnv(t, "# comment one\nDB_NAME=foo\n# comment two\n")

	got, err := ScanComments(path, "all")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}
	want := []string{"# comment one", "# comment two"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScanComments_BlocksOnly(t *testing.T) {
	path := writeTempEnv(t, "# single line\nDB_NAME=foo\n# block line 1\n# block line 2\n")

	got, err := ScanComments(path, "blocks-only")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}
	want := []string{"# block line 1", "# block line 2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScanComments_None(t *testing.T) {
	path := writeTempEnv(t, "# a comment\nDB_NAME=foo\n")

	for _, mode := range []string{"none", ""} {
		got, err := ScanComments(path, mode)
		if err != nil {
			t.Fatalf("ScanComments(%q): %v", mode, err)
		}
		if got != nil {
			t.Errorf("mode %q: got %v, want nil", mode, got)
		}
	}
}

func TestScanComments_NonExistentFile(t *testing.T) {
	got, err := ScanComments("/tmp/bight-does-not-exist-12345.env", "all")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestPatchAll_MultiplePatches(t *testing.T) {
	path := writeTempEnv(t, "DB_NAME=old\nAPP_ENV=local\n")

	patches := map[string]string{"DB_NAME": "new", "APP_ENV": "staging"}
	if err := PatchAll(path, patches, nil); err != nil {
		t.Fatalf("PatchAll: %v", err)
	}

	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if env["DB_NAME"] != "new" {
		t.Errorf("DB_NAME = %q, want %q", env["DB_NAME"], "new")
	}
	if env["APP_ENV"] != "staging" {
		t.Errorf("APP_ENV = %q, want %q", env["APP_ENV"], "staging")
	}
}

func TestPatchAll_WithComments(t *testing.T) {
	path := writeTempEnv(t, "# header comment\n# second line\nDB_NAME=old\n")

	comments, err := ScanComments(path, "all")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}

	if err := PatchAll(path, map[string]string{"DB_NAME": "new"}, comments); err != nil {
		t.Fatalf("PatchAll: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "# header comment") {
		t.Errorf("missing '# header comment' in:\n%s", content)
	}
	if !strings.Contains(content, "# second line") {
		t.Errorf("missing '# second line' in:\n%s", content)
	}
	if !strings.Contains(content, "DB_NAME=") {
		t.Errorf("missing 'DB_NAME=' in:\n%s", content)
	}
	keyIdx := strings.Index(content, "DB_NAME=")
	commentIdx := strings.Index(content, "# header comment")
	if commentIdx < keyIdx {
		t.Errorf("comments appear before key=value content")
	}
}

func TestPatchAll_NilCommentsNoOp(t *testing.T) {
	path := writeTempEnv(t, "DB_NAME=foo\n")

	if err := PatchAll(path, map[string]string{"DB_NAME": "bar"}, nil); err != nil {
		t.Fatalf("PatchAll: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "#") {
		t.Errorf("unexpected comment in output:\n%s", content)
	}
}

func TestPatchAll_NonExistentFile(t *testing.T) {
	path := os.TempDir() + "/bight-test-patchall-new.env"
	defer os.Remove(path)

	if err := PatchAll(path, map[string]string{"SECRET": "abc123"}, nil); err != nil {
		t.Fatalf("PatchAll: %v", err)
	}

	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if env["SECRET"] != "abc123" {
		t.Errorf("SECRET = %q, want %q", env["SECRET"], "abc123")
	}
}

func TestPatchAll_PermissionsPreserved(t *testing.T) {
	path := writeTempEnv(t, "KEY=old\n")
	if err := os.Chmod(path, 0640); err != nil {
		t.Fatal(err)
	}

	if err := PatchAll(path, map[string]string{"KEY": "new"}, nil); err != nil {
		t.Fatalf("PatchAll: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0640 {
		t.Errorf("mode = %04o, want 0640", fi.Mode().Perm())
	}
}

func TestPatchAll_DefaultPermissionsForNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.new")

	if err := PatchAll(path, map[string]string{"X": "1"}, nil); err != nil {
		t.Fatalf("PatchAll: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0600 {
		t.Errorf("mode = %04o, want 0600", fi.Mode().Perm())
	}
}

func TestPatchAll_TempFileRemovedOnError(t *testing.T) {
	dir := t.TempDir()
	// Point at a non-existent subdirectory so CreateTemp fails.
	path := filepath.Join(dir, "nonexistent-subdir", ".env")

	err := PatchAll(path, map[string]string{"KEY": "val"}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("stale temp file found: %s", e.Name())
		}
	}
}

func TestPatch_IsTransactional(t *testing.T) {
	path := writeTempEnv(t, "A=1\nB=2\n")

	if err := Patch(path, "A", "updated"); err != nil {
		t.Fatalf("Patch: %v", err)
	}

	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if env["A"] != "updated" {
		t.Errorf("A = %q, want %q", env["A"], "updated")
	}
	if env["B"] != "2" {
		t.Errorf("B = %q, want %q", env["B"], "2")
	}
}
