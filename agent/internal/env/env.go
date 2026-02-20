package env

import (
	"bufio"
	"crypto/subtle"
	"fmt"
	"os"
	"strings"
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

// Write atomically writes key=value pairs to path.
// Steps: backup → write to .tmp → rename .tmp to path.
func Write(path string, vars map[string]string) error {
	// 1. Backup
	backup := fmt.Sprintf("%s.backup.%d", path, time.Now().Unix())
	if err := copyFile(path, backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("backing up env file: %w", err)
	}

	// 2. Write to temp file
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	w := bufio.NewWriter(f)
	for k, v := range vars {
		if _, err := fmt.Fprintf(w, "%s=%s\n", k, v); err != nil {
			f.Close()
			os.Remove(tmp)
			return fmt.Errorf("writing env var: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()

	// 3. Atomic rename
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
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
