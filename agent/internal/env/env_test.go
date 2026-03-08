package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsSensitive(t *testing.T) {
	cases := []struct {
		key      string
		expected bool
	}{
		{"JWT_SECRET", true},
		{"DB_PASSWORD", true},
		{"API_KEY", true},
		{"ACCESS_TOKEN", true},
		{"AUTH_HEADER", true},
		{"PRIVATE_KEY", true},
		{"DB_PWD", true},
		{"DATABASE_URL", false},
		{"PORT", false},
		{"NODE_ENV", false},
		{"APP_NAME", false},
		// case-insensitive
		{"jwt_secret", true},
		{"db_password", true},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			got := IsSensitive(tc.key)
			if got != tc.expected {
				t.Errorf("IsSensitive(%q) = %v, want %v", tc.key, got, tc.expected)
			}
		})
	}
}

func TestTokensMatch(t *testing.T) {
	if !TokensMatch("abc123", "abc123") {
		t.Error("identical tokens should match")
	}
	if TokensMatch("abc123", "abc124") {
		t.Error("different tokens should not match")
	}
	if TokensMatch("", "notempty") {
		t.Error("empty vs non-empty should not match")
	}
	// Two empty strings are considered equal by ConstantTimeCompare — that's correct.
	// Rejecting empty tokens is the middleware's responsibility, not TokensMatch.
}

func TestParse(t *testing.T) {
	content := `
# comment
DATABASE_URL=postgres://localhost/mydb
JWT_SECRET=supersecret
PORT=3000
EMPTY=
`
	f, err := os.CreateTemp(t.TempDir(), "*.env")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()

	vars, err := Parse(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v, ok := vars["DATABASE_URL"]; !ok {
		t.Error("DATABASE_URL missing")
	} else if v.Masked {
		t.Error("DATABASE_URL should not be masked")
	} else if v.Value != "postgres://localhost/mydb" {
		t.Errorf("DATABASE_URL value = %q", v.Value)
	}

	if v, ok := vars["JWT_SECRET"]; !ok {
		t.Error("JWT_SECRET missing")
	} else if !v.Masked {
		t.Error("JWT_SECRET should be masked")
	} else if v.Value != "••••••••" {
		t.Errorf("masked value = %q, want ••••••••", v.Value)
	}

	if v, ok := vars["PORT"]; !ok {
		t.Error("PORT missing")
	} else if v.Value != "3000" {
		t.Errorf("PORT = %q, want 3000", v.Value)
	}

	if _, ok := vars["#"]; ok {
		t.Error("comment lines should not be parsed")
	}
}

func TestWrite_AtomicAndReadback(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.env"

	vars := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}

	if err := Write(path, vars); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	// Verify no .tmp file left behind
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Error(".tmp file should not exist after successful write")
	}

	// Read back and verify
	got, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse after Write error: %v", err)
	}
	if got["FOO"].Value != "bar" {
		t.Errorf("FOO = %q, want bar", got["FOO"].Value)
	}
}

func TestWrite_PreservesCommentsAndOrder(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.env"

	original := "# Top comment\nFOO=old\n\n# Section\nBAR=old\nBAZ=keep\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	if err := Write(path, map[string]string{"FOO": "new", "BAR": "new", "BAZ": "keep"}); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Comments and blank lines must be present.
	if !strings.Contains(content, "# Top comment") {
		t.Error("top comment missing")
	}
	if !strings.Contains(content, "# Section") {
		t.Error("section comment missing")
	}

	// Updated values must be written.
	if !strings.Contains(content, "FOO=new") {
		t.Error("FOO value not updated")
	}
	if !strings.Contains(content, "BAR=new") {
		t.Error("BAR value not updated")
	}

	// Original order must be preserved: FOO before BAR.
	fooIdx := strings.Index(content, "FOO=")
	barIdx := strings.Index(content, "BAR=")
	if fooIdx > barIdx {
		t.Error("key order not preserved")
	}
}

func TestWrite_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.env"

	// Write initial file
	os.WriteFile(path, []byte("ORIGINAL=yes\n"), 0600)

	if err := Write(path, map[string]string{"NEW": "value"}); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	// A backup file should exist inside .env-backups/
	entries, _ := os.ReadDir(filepath.Join(dir, ".env-backups"))
	if len(entries) == 0 {
		t.Error("expected backup file to be created in .env-backups/")
	}
}
