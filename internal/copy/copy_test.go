package copy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFile_CopiesContents(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	content := []byte("KEY=val\nOTHER=thing\n")
	if err := os.WriteFile(src, content, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := File(src, dst); err != nil {
		t.Fatalf("File: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("dst content = %q, want %q", got, content)
	}
}

func TestFile_PreservesPerms(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := File(src, dst); err != nil {
		t.Fatalf("File: %v", err)
	}

	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if got := fi.Mode().Perm(); got != 0o640 {
		t.Errorf("dst perm = %#o, want %#o", got, 0o640)
	}
}

func TestFile_OverwritesExistingDest(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := File(src, dst); err != nil {
		t.Fatalf("File: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "new" {
		t.Errorf("dst = %q, want %q", got, "new")
	}
}

func TestFile_SourceMissing(t *testing.T) {
	dir := t.TempDir()
	err := File(filepath.Join(dir, "no-such"), filepath.Join(dir, "dst"))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestResolveSource_Absolute(t *testing.T) {
	got, err := ResolveSource("/abs/path/.env", "/base", "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/abs/path/.env" {
		t.Errorf("got %q", got)
	}
}

func TestResolveSource_Relative(t *testing.T) {
	got, err := ResolveSource("../main/.env", "/base/sub", "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/base/main/.env" {
		t.Errorf("got %q, want %q", got, "/base/main/.env")
	}
}

func TestResolveSource_TildeExpands(t *testing.T) {
	got, err := ResolveSource("~/envs/myapp.env", "/base", "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/home/u/envs/myapp.env" {
		t.Errorf("got %q, want %q", got, "/home/u/envs/myapp.env")
	}
}

func TestResolveSource_TildeAlone(t *testing.T) {
	got, err := ResolveSource("~", "/base", "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/home/u" {
		t.Errorf("got %q, want %q", got, "/home/u")
	}
}

func TestResolveSource_TildeWithoutHome(t *testing.T) {
	_, err := ResolveSource("~/x", "/base", "")
	if err == nil {
		t.Fatal("expected error when home is empty")
	}
}

func TestResolveSource_UserTildeUnsupported(t *testing.T) {
	_, err := ResolveSource("~someone/x", "/base", "/home/u")
	if err == nil {
		t.Fatal("expected error for ~user form")
	}
}

func TestResolveSource_Empty(t *testing.T) {
	_, err := ResolveSource("", "/base", "/home/u")
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}
