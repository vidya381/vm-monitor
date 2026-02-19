package systemd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Status string

const (
	StatusRunning   Status = "running"
	StatusStopped   Status = "stopped"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// AppStatus returns the current status of a systemd service.
func AppStatus(service string) Status {
	out, err := exec.Command("systemctl", "is-active", service).Output()
	if err != nil {
		return StatusStopped
	}
	switch strings.TrimSpace(string(out)) {
	case "active":
		return StatusRunning
	case "failed":
		return StatusUnhealthy
	default:
		return StatusStopped
	}
}

// LogEntry is a single journalctl JSON log line.
type LogEntry struct {
	Message string `json:"MESSAGE"`
	Cursor  string `json:"__CURSOR"`
	// RealtimeTimestamp is microseconds since epoch as a string.
	RealtimeTimestamp string `json:"__REALTIME_TIMESTAMP"`
}

// LogResult holds log lines and the cursor for the next poll.
type LogResult struct {
	Lines   []string `json:"lines"`
	Cursor  string   `json:"cursor"`
	HasMore bool     `json:"has_more"`
}

// Logs fetches log lines for a systemd service.
// If cursor is empty, fetches the last n lines.
// If cursor is set, fetches only lines after that cursor.
func Logs(service string, tail int, cursor string) (*LogResult, error) {
	args := []string{"-u", service, "-o", "json", "--no-pager"}
	if cursor != "" {
		args = append(args, "--after-cursor="+cursor)
	} else {
		args = append(args, fmt.Sprintf("-n%d", tail))
	}

	out, err := exec.Command("journalctl", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("journalctl: %w", err)
	}

	var lines []string
	var lastCursor string

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry.Message != "" {
			lines = append(lines, entry.Message)
		}
		if entry.Cursor != "" {
			lastCursor = entry.Cursor
		}
	}

	return &LogResult{
		Lines:   lines,
		Cursor:  lastCursor,
		HasMore: false,
	}, nil
}

// Restart restarts a systemd service.
func Restart(service string) error {
	cmd := exec.Command("systemctl", "restart", service)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restarting %s: %s", service, stderr.String())
	}
	return nil
}
