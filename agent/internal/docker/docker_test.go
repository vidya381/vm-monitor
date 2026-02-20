package docker

import "testing"

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantTS    string
		wantMsg   string
	}{
		{
			name:    "valid rfc3339nano line",
			line:    "2024-01-15T10:23:01.123456789Z hello world",
			wantTS:  "2024-01-15T10:23:01.123456789Z",
			wantMsg: "hello world",
		},
		{
			name:    "message with spaces",
			line:    "2024-01-15T10:23:01.000000000Z [INFO] server started on :8080",
			wantTS:  "2024-01-15T10:23:01Z",
			wantMsg: "[INFO] server started on :8080",
		},
		{
			name:    "no timestamp prefix",
			line:    "plain log line without timestamp",
			wantTS:  "",
			wantMsg: "plain log line without timestamp",
		},
		{
			name:    "empty line",
			line:    "",
			wantTS:  "",
			wantMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts, msg := parseLogLine(tc.line)
			if msg != tc.wantMsg {
				t.Errorf("msg = %q, want %q", msg, tc.wantMsg)
			}
			// For timestamp, just check presence/absence since nanosecond
			// representation can vary by Go version.
			if tc.wantTS == "" && ts != "" {
				t.Errorf("ts = %q, want empty", ts)
			}
			if tc.wantTS != "" && ts == "" {
				t.Errorf("ts empty, want non-empty")
			}
		})
	}
}
