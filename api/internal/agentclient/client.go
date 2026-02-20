package agentclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client makes authenticated requests to agent HTTP servers.
type Client struct {
	http *http.Client
}

func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// AgentApp is the status of one app as returned by GET /apps on the agent.
type AgentApp struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// GetApps fetches all app statuses from an agent.
func (c *Client) GetApps(address, token string) ([]AgentApp, error) {
	var apps []AgentApp
	if err := c.get(address+"/apps", token, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// ProxyRequest forwards a request from the control plane to an agent and writes
// the agent's response directly to w. Returns the agent's HTTP status code so
// callers can decide whether to perform post-action work (e.g. audit logging).
// Returns 0 if the request could not be sent at all.
func (c *Client) ProxyRequest(w http.ResponseWriter, r *http.Request, agentURL, token string) int {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, agentURL, r.Body)
	if err != nil {
		http.Error(w, "failed to build proxy request", http.StatusInternalServerError)
		return 0
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	req.URL.RawQuery = r.URL.RawQuery

	resp, err := c.http.Do(req)
	if err != nil {
		http.Error(w, "agent unreachable", http.StatusBadGateway)
		return 0
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	w.Write(buf.Bytes())
	return resp.StatusCode
}

func (c *Client) get(url, token string, out any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent returned %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
