package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EventType string

const (
	EventDown     EventType = "down"
	EventRecovered EventType = "recovered"
)

type Event struct {
	Type     EventType
	VMName   string
	AppName  string
	Duration time.Duration // how long it was down (only set for recovered)
	AutoRestarted bool          // true if auto-restart was triggered (only for EventDown)
}

// Notifier sends notifications when apps go down or recover.
type Notifier struct {
	webhookURL  string
	webhookType string // "slack" or "generic"
	http        *http.Client
}

func New(webhookURL, webhookType string) *Notifier {
	return &Notifier{
		webhookURL:  webhookURL,
		webhookType: webhookType,
		http:        &http.Client{Timeout: 5 * time.Second},
	}
}

// Enabled returns true if a webhook URL is configured.
func (n *Notifier) Enabled() bool {
	return n.webhookURL != ""
}

// Send fires a notification for the given event.
func (n *Notifier) Send(event Event) error {
	if !n.Enabled() {
		return nil
	}
	switch n.webhookType {
	case "slack":
		return n.sendSlack(event)
	default:
		return n.sendGeneric(event)
	}
}

// sendSlack sends a Slack-formatted webhook message.
func (n *Notifier) sendSlack(event Event) error {
	var text string
	switch event.Type {
	case EventDown:
		if event.AutoRestarted {
			text = fmt.Sprintf(":red_circle: *%s / %s* is down — auto-restart triggered", event.VMName, event.AppName)
		} else {
			text = fmt.Sprintf(":red_circle: *%s / %s* is down", event.VMName, event.AppName)
		}
	case EventRecovered:
		if event.Duration > 0 {
			text = fmt.Sprintf(":white_check_mark: *%s / %s* recovered (was down for %s)",
				event.VMName, event.AppName, formatDuration(event.Duration))
		} else {
			text = fmt.Sprintf(":white_check_mark: *%s / %s* recovered", event.VMName, event.AppName)
		}
	}

	payload := map[string]string{"text": text}
	return n.post(payload)
}

// sendGeneric sends a simple JSON payload compatible with most webhook services
// (Discord, ntfy, custom endpoints, etc.).
func (n *Notifier) sendGeneric(event Event) error {
	var title, body string
	switch event.Type {
	case EventDown:
		title = fmt.Sprintf("%s / %s is down", event.VMName, event.AppName)
		if event.AutoRestarted {
			body = fmt.Sprintf("App %s on VM %s has stopped or is unhealthy. Auto-restart triggered.", event.AppName, event.VMName)
		} else {
			body = fmt.Sprintf("App %s on VM %s has stopped or is unhealthy.", event.AppName, event.VMName)
		}
	case EventRecovered:
		title = fmt.Sprintf("%s / %s recovered", event.VMName, event.AppName)
		if event.Duration > 0 {
			body = fmt.Sprintf("App %s on VM %s is back online after %s.",
				event.AppName, event.VMName, formatDuration(event.Duration))
		} else {
			body = fmt.Sprintf("App %s on VM %s is back online.", event.AppName, event.VMName)
		}
	}

	payload := map[string]string{
		"title":   title,
		"message": body,
		"event":   string(event.Type),
		"vm":      event.VMName,
		"app":     event.AppName,
	}
	return n.post(payload)
}

func (n *Notifier) post(payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	resp, err := n.http.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
