// Package docker provides status, log, and restart operations for Docker containers.
// It mirrors the interface of the systemd package so handlers can dispatch uniformly.
package docker

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/vidya381/vm-monitor/agent/internal/systemd"
)

// AppStatus returns the running status of a Docker container.
func AppStatus(container string) systemd.Status {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", container).Output()
	if err != nil {
		// Container not found or Docker not reachable.
		return systemd.StatusUnknown
	}
	switch strings.TrimSpace(string(out)) {
	case "running":
		return systemd.StatusRunning
	case "exited", "created":
		return systemd.StatusStopped
	case "restarting", "paused", "dead", "removing":
		return systemd.StatusUnhealthy
	default:
		return systemd.StatusUnknown
	}
}

// Logs fetches log lines for a Docker container.
//
// On the initial call (cursor == ""), the last n lines are returned along with
// an RFC3339Nano timestamp cursor.  On subsequent calls the cursor is passed as
// --since to retrieve only new lines.
func Logs(container string, tail int, cursor string) (*systemd.LogResult, error) {
	args := []string{"logs", "--timestamps"}
	if cursor != "" {
		// Add one nanosecond to avoid re-returning the last line we already sent.
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err == nil {
			args = append(args, "--since", t.Add(time.Nanosecond).UTC().Format(time.RFC3339Nano))
		} else {
			args = append(args, "--since", cursor)
		}
	} else {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	args = append(args, container)

	// docker logs writes to stderr by default; combine both streams.
	cmd := exec.Command("docker", args...)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("docker logs: %w", err)
	}

	var lines []string
	var lastCursor string

	for _, raw := range strings.Split(combined.String(), "\n") {
		raw = strings.TrimRight(raw, "\r")
		if raw == "" {
			continue
		}
		ts, msg := parseLogLine(raw)
		if msg != "" {
			lines = append(lines, msg)
		}
		if ts != "" {
			lastCursor = ts
		}
	}

	return &systemd.LogResult{
		Lines:   lines,
		Cursor:  lastCursor,
		HasMore: false,
	}, nil
}

// StreamLogs starts docker logs -f for a container and returns a ReadCloser
// that streams new log lines. Closing it kills the subprocess.
func StreamLogs(container string) (io.ReadCloser, error) {
	cmd := exec.Command("docker", "logs", "-f", "--timestamps", container)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return nil, fmt.Errorf("starting docker logs: %w", err)
	}
	go func() {
		cmd.Wait()
		pw.Close()
	}()
	return &cmdReadCloser{ReadCloser: pr, cmd: cmd}, nil
}

type cmdReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (c *cmdReadCloser) Close() error {
	err := c.ReadCloser.Close()
	if c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	c.cmd.Wait()
	return err
}

// GetPID returns the main PID of a running Docker container.
func GetPID(container string) (int, error) {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Pid}}", container).Output()
	if err != nil {
		return 0, fmt.Errorf("docker inspect: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || pid == 0 {
		return 0, fmt.Errorf("no running PID for container %s", container)
	}
	return pid, nil
}

// Restart restarts a Docker container.
func Restart(container string) error {
	cmd := exec.Command("docker", "restart", container)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restarting %s: %s", container, stderr.String())
	}
	return nil
}

// parseLogLine splits a docker --timestamps log line into timestamp and message.
// Docker emits lines like: "2024-01-15T10:23:01.123456789Z message here"
func parseLogLine(line string) (timestamp, message string) {
	idx := strings.IndexByte(line, ' ')
	if idx < 0 {
		return "", line
	}
	ts := line[:idx]
	msg := line[idx+1:]
	// Validate the timestamp is parseable; if not, treat the whole line as message.
	if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
		return "", line
	}
	return ts, msg
}
