package env

import (
	"os"
	"path/filepath"
	"reflect"
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

func assertBlocks(t *testing.T, got, want []CommentBlock) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v\nwant %#v", got, want)
	}
}

func TestScanComments_All(t *testing.T) {
	path := writeTempEnv(t, "# comment one\nDB_NAME=foo\n# comment two\n")

	got, err := ScanComments(path, "all")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}
	assertBlocks(t, got, []CommentBlock{
		{Lines: []string{"# comment one"}, Anchor: "DB_NAME"},
		{Lines: []string{"# comment two"}, Anchor: "DB_NAME", After: true},
	})
}

func TestScanComments_BlocksOnly(t *testing.T) {
	path := writeTempEnv(t, "# single line\nDB_NAME=foo\n# block line 1\n# block line 2\n")

	got, err := ScanComments(path, "blocks-only")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}
	assertBlocks(t, got, []CommentBlock{
		{Lines: []string{"# block line 1", "# block line 2"}, Anchor: "DB_NAME", After: true},
	})
}

func TestScanComments_Header(t *testing.T) {
	path := writeTempEnv(t, "# header\n\n# about A\nA=1\n")

	got, err := ScanComments(path, "all")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}
	assertBlocks(t, got, []CommentBlock{
		{Lines: []string{"# header"}, Header: true},
		{Lines: []string{"# about A"}, Anchor: "A"},
	})
}

func TestScanComments_AnchorSyntax(t *testing.T) {
	path := writeTempEnv(t, "# a\nexport A=1\n# b\nB: 2\n")

	got, err := ScanComments(path, "all")
	if err != nil {
		t.Fatalf("ScanComments: %v", err)
	}
	assertBlocks(t, got, []CommentBlock{
		{Lines: []string{"# a"}, Anchor: "A"},
		{Lines: []string{"# b"}, Anchor: "B"},
	})
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

func TestPatchAll_CommentPlacement(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		input string
		want  string
	}{
		{
			name:  "anchored comment follows its key on sort",
			mode:  "all",
			input: "# z note\nZ=1\nA=2\n",
			want:  "A=2\n# z note\nZ=1\n",
		},
		{
			name:  "header stays at top",
			mode:  "all",
			input: "# header\n# more\n\nZ=1\nA=2\n",
			want:  "# header\n# more\n\nA=2\nZ=1\n",
		},
		{
			name:  "trailing comment follows preceding key on sort",
			mode:  "all",
			input: "DB_NAME=demo\n\nFIREBASE_VAL_1=some-val\n#FIREBASE_VAL_2=other-val\n\nA=2\n",
			want:  "A=2\nDB_NAME=\"demo\"\nFIREBASE_VAL_1=\"some-val\"\n#FIREBASE_VAL_2=other-val\n",
		},
		{
			name:  "block between two keys attaches to the key below",
			mode:  "all",
			input: "B=1\n# about C\nC=3\nA=2\n",
			want:  "A=2\nB=1\n# about C\nC=3\n",
		},
		{
			name:  "detached block goes to end",
			mode:  "all",
			input: "B=1\n\n# detached\n\nA=2\n",
			want:  "A=2\nB=1\n\n# detached\n",
		},
		{
			name:  "blocks-only drops single-line anchored comment",
			mode:  "blocks-only",
			input: "# single\nB=1\n# one\n# two\nA=2\n",
			want:  "# one\n# two\nA=2\nB=1\n",
		},
		{
			name:  "export and colon anchors resolve",
			mode:  "all",
			input: "# b\nexport B=1\n# a\nA: 2\n",
			want:  "# a\nA=2\n# b\nB=1\n",
		},
		{
			name:  "multiple blocks for one key keep order",
			mode:  "all",
			input: "# first\nA=1\n# second\nA=2\n",
			want:  "# first\n# second\nA=2\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempEnv(t, tc.input)
			comments, err := ScanComments(path, tc.mode)
			if err != nil {
				t.Fatalf("ScanComments: %v", err)
			}
			if err := PatchAll(path, map[string]string{"A": "2"}, comments); err != nil {
				t.Fatalf("PatchAll: %v", err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", data, tc.want)
			}
		})
	}
}

func TestMergeComments_OrphanedAnchorGoesToEnd(t *testing.T) {
	got := mergeComments("A=1", []CommentBlock{
		{Lines: []string{"# gone"}, Anchor: "GONE"},
		{Lines: []string{"# gone after"}, Anchor: "GONE_TOO", After: true},
	})
	want := "A=1\n\n# gone\n# gone after"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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

func TestBackupFile_CreatesBackup(t *testing.T) {
	path := writeTempEnv(t, "KEY=original\n")
	if err := os.Chmod(path, 0640); err != nil {
		t.Fatal(err)
	}

	if err := BackupFile(path); err != nil {
		t.Fatalf("BackupFile: %v", err)
	}

	data, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	t.Cleanup(func() { os.Remove(path + ".bak") })

	if string(data) != "KEY=original\n" {
		t.Errorf("backup content = %q, want %q", string(data), "KEY=original\n")
	}

	fi, err := os.Stat(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0640 {
		t.Errorf("backup mode = %04o, want 0640", fi.Mode().Perm())
	}
}

func TestBackupFile_NonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.env")

	if err := BackupFile(path); err != nil {
		t.Errorf("BackupFile on missing file: %v, want nil", err)
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Errorf("expected no backup file, but it exists")
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
