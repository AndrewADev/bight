package env

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func BackupFile(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return os.WriteFile(path+".bak", data, fi.Mode().Perm())
}

func Patch(path, key, value string) error {
	return PatchAll(path, map[string]string{key: value}, nil)
}

// CommentBlock is a contiguous run of full-line comments from an env file.
type CommentBlock struct {
	Lines  []string
	Anchor string // key the block is attached to; empty for detached blocks
	After  bool   // block goes directly below Anchor instead of above it
	Header bool   // detached block located before the first key
}

// ScanComments reads path and returns its full-line comment blocks per mode.
// mode: "all" | "blocks-only" | anything else → nil, nil
//
// A block directly followed by a key line is anchored above that key. A block
// directly preceded by a key line and followed by a blank line or EOF is
// anchored below the preceding key. Any other block is detached; detached
// blocks before the first key are marked as Header.
func ScanComments(path, mode string) ([]CommentBlock, error) {
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

	var collected []CommentBlock
	var currentBlock []string
	seenKey := false
	prevKey := ""    // key on the line directly above the current line, if any
	blockAfter := "" // key directly above the current block, if any

	flush := func(anchor string) {
		if mode == "blocks-only" && len(currentBlock) < 2 {
			currentBlock = nil
			return
		}
		block := CommentBlock{Lines: currentBlock, Anchor: anchor}
		switch {
		case anchor == "" && blockAfter != "":
			block.Anchor = blockAfter
			block.After = true
		case anchor == "":
			block.Header = !seenKey
		}
		collected = append(collected, block)
		currentBlock = nil
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#"):
			if len(currentBlock) == 0 {
				blockAfter = prevKey
			}
			currentBlock = append(currentBlock, line)
			prevKey = ""
		case trimmed == "":
			if len(currentBlock) > 0 {
				flush("")
			}
			prevKey = ""
		default:
			key := parseKey(line)
			if len(currentBlock) > 0 {
				flush(key)
			}
			seenKey = true
			prevKey = key
		}
	}
	if len(currentBlock) > 0 {
		flush("")
	}

	return collected, scanner.Err()
}

// parseKey returns the key of an env line, accepting an optional "export "
// prefix and either "=" or ":" as the separator.
func parseKey(line string) string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "export ")
	if i := strings.IndexAny(trimmed, "=:"); i >= 0 {
		trimmed = trimmed[:i]
	}
	return strings.TrimSpace(trimmed)
}

// mergeComments inserts comment blocks into marshaled env content. Anchored
// blocks go directly above or below their key, header blocks at the top
// followed by a blank line, and all remaining blocks at the end preceded by a
// blank line.
func mergeComments(content string, blocks []CommentBlock) string {
	if len(blocks) == 0 {
		return content
	}

	var header, trailing []string
	above := make(map[string][]string)
	below := make(map[string][]string)
	for _, b := range blocks {
		switch {
		case b.Anchor != "" && b.After:
			below[b.Anchor] = append(below[b.Anchor], b.Lines...)
		case b.Anchor != "":
			above[b.Anchor] = append(above[b.Anchor], b.Lines...)
		case b.Header:
			header = append(header, b.Lines...)
		default:
			trailing = append(trailing, b.Lines...)
		}
	}

	var out []string
	if len(header) > 0 {
		out = append(out, header...)
		out = append(out, "")
	}
	if content != "" {
		for line := range strings.SplitSeq(content, "\n") {
			key := parseKey(line)
			out = append(out, above[key]...)
			out = append(out, line)
			out = append(out, below[key]...)
			delete(above, key)
			delete(below, key)
		}
	}

	// Blocks whose anchor key is missing from the output follow detached blocks.
	for _, b := range blocks {
		anchored := above
		if b.After {
			anchored = below
		}
		if lines, ok := anchored[b.Anchor]; ok {
			trailing = append(trailing, lines...)
			delete(anchored, b.Anchor)
		}
	}
	if len(trailing) > 0 {
		out = append(out, "")
		out = append(out, trailing...)
	}
	return strings.Join(out, "\n")
}

// PatchAll reads existing env from path, merges patches, and writes atomically via a
// temp file + rename. If path does not exist, starts from an empty env.
// comments are placed per mergeComments and may be nil (no-op). The original
// file's permissions are preserved.
func PatchAll(path string, patches map[string]string, comments []CommentBlock) error {
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

	maps.Copy(existing, patches)

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

	if _, err := tmp.WriteString(mergeComments(content, comments) + "\n"); err != nil {
		return err
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
