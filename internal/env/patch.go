package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func Patch(path, key, value string) error {
	return PatchAll(path, map[string]string{key: value}, nil)
}

// ScanComments reads path and returns full-line comment lines per mode.
// mode: "all" | "blocks-only" | anything else → nil, nil
func ScanComments(path, mode string) ([]string, error) {
	if mode == "" || mode == "none" {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var collected []string
	var currentBlock []string

	flush := func() {
		if mode == "blocks-only" && len(currentBlock) < 2 {
			currentBlock = currentBlock[:0]
			return
		}
		collected = append(collected, currentBlock...)
		currentBlock = currentBlock[:0]
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			currentBlock = append(currentBlock, line)
		} else {
			if len(currentBlock) > 0 {
				flush()
			}
		}
	}
	if len(currentBlock) > 0 {
		flush()
	}

	return collected, scanner.Err()
}

// PatchAll reads existing env from path, merges patches, and writes atomically via a
// temp file + rename. If path does not exist, starts from an empty env.
// comments may be nil (no-op). The original file's permissions are preserved.
func PatchAll(path string, patches map[string]string, comments []string) error {
	var originalMode os.FileMode = 0600
	if fi, err := os.Stat(path); err == nil {
		originalMode = fi.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}

	existing, err := godotenv.Read(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		existing = make(map[string]string)
	}

	for k, v := range patches {
		existing[k] = v
	}

	content, err := godotenv.Marshal(existing)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp, err := os.CreateTemp(dir, fmt.Sprintf(".%s.*.tmp", base))
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.WriteString(content + "\n"); err != nil {
		return err
	}
	if len(comments) > 0 {
		if _, err := tmp.WriteString("\n" + strings.Join(comments, "\n") + "\n"); err != nil {
			return err
		}
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tmpName, originalMode); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	committed = true
	return nil
}
