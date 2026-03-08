package env

import (
	"bufio"
	"crypto/subtle"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type EnvVar struct {
	Value  string `json:"value"`
	Masked bool   `json:"masked"`
}

var sensitiveKeyPatterns = []string{
	"SECRET", "PASSWORD", "PASS", "KEY", "TOKEN",
	"AUTH", "CREDENTIAL", "PRIVATE", "PWD",
}

func IsSensitive(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range sensitiveKeyPatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

// TokensMatch compares two tokens in constant time to prevent timing attacks.
func TokensMatch(incoming, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(incoming), []byte(expected)) == 1
}

// ParseRaw reads an .env file and returns all key=value pairs without masking.
// Used internally by PUT /env to merge new values on top of existing ones.
func ParseRaw(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening env file: %w", err)
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		result[key] = value
	}
	return result, scanner.Err()
}

// Parse reads an .env file and returns a map of key → EnvVar.
// Sensitive values are replaced with "••••••••".
func Parse(path string) (map[string]EnvVar, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening env file: %w", err)
	}
	defer f.Close()

	result := make(map[string]EnvVar)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		masked := IsSensitive(key)
		displayValue := value
		if masked {
			displayValue = "••••••••"
		}
		result[key] = EnvVar{Value: displayValue, Masked: masked}
	}
	return result, scanner.Err()
}

// Write atomically writes key=value pairs to path, preserving the original
// file's comments, blank lines, and key order.
// Steps: capture metadata → backup → build output → write .tmp → rename .tmp to path → restore ownership.
func Write(path string, vars map[string]string) error {
	// 1. Capture original file permissions and ownership before touching anything.
	var mode os.FileMode = 0600
	var uid, gid int = -1, -1
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			uid = int(st.Uid)
			gid = int(st.Gid)
		}
	}

	// 2. Backup into a dedicated .env-backups/ subdirectory next to the file.
	backupDir := filepath.Join(filepath.Dir(path), ".env-backups")
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("creating backup dir: %w", err)
	}
	backup := filepath.Join(backupDir, fmt.Sprintf("%s.%d", filepath.Base(path), time.Now().Unix()))
	if err := copyFile(path, backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("backing up env file: %w", err)
	}

	// 3. Build output lines, preserving the structure of the original file
	// (comments, blank lines, key order). Only the values are updated.
	var outLines []string
	written := make(map[string]bool)

	if orig, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(orig)
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				outLines = append(outLines, line)
				continue
			}
			key, _, ok := strings.Cut(trimmed, "=")
			if !ok {
				outLines = append(outLines, line)
				continue
			}
			key = strings.TrimSpace(key)
			written[key] = true
			if newVal, exists := vars[key]; exists {
				outLines = append(outLines, fmt.Sprintf("%s=%s", key, newVal))
			} else {
				outLines = append(outLines, line)
			}
		}
		orig.Close()
	}

	// Append any brand-new keys not present in the original file.
	for k, v := range vars {
		if !written[k] {
			outLines = append(outLines, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 4. Write to temp file using original permissions.
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	w := bufio.NewWriter(f)
	for _, line := range outLines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			f.Close()
			os.Remove(tmp)
			return fmt.Errorf("writing line: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()

	// 5. Atomic rename
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	// 6. Restore original ownership (best-effort; only fails if agent lacks CAP_CHOWN).
	if uid >= 0 {
		_ = os.Chown(path, uid, gid)
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
