package env

import (
	"bufio"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func Patch(path, key, value string) error {
	existing, err := godotenv.Read(path)
	if err != nil {
		// If file doesn't exist, start fresh
		existing = make(map[string]string)
	}
	existing[key] = value
	return godotenv.Write(existing, path)
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

// PatchAll reads existing env from path, merges patches, writes once, then appends comments.
// If path does not exist, starts from an empty env. comments may be nil (no-op).
func PatchAll(path string, patches map[string]string, comments []string) error {
	existing, err := godotenv.Read(path)
	if err != nil {
		existing = make(map[string]string)
	}
	for k, v := range patches {
		existing[k] = v
	}
	if err := godotenv.Write(existing, path); err != nil {
		return err
	}
	return appendComments(path, comments)
}

// appendComments appends collected comment lines to path, preceded by a blank line.
// No-op if comments is empty.
func appendComments(path string, comments []string) error {
	if len(comments) == 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("\n" + strings.Join(comments, "\n") + "\n")
	return err
}
