package alert

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert holds information about a triggered alert.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	Timestamp time.Time
}

// Notifier defines the interface for sending alerts.
type Notifier interface {
	Send(a Alert) error
}

// LogNotifier writes alerts to the standard logger.
type LogNotifier struct{}

func (l *LogNotifier) Send(a Alert) error {
	log.Printf("[%s] %s — %s (at %s)", a.Level, a.JobName, a.Message, a.Timestamp.Format(time.RFC3339))
	return nil
}

// WebhookNotifier posts alerts to an HTTP endpoint.
type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL:    url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WebhookNotifier) Send(a Alert) error {
	body := fmt.Sprintf(`{"job":%q,"level":%q,"message":%q,"timestamp":%q}`,
		a.JobName, a.Level, a.Message, a.Timestamp.Format(time.RFC3339))
	resp, err := w.Client.Post(w.URL, "application/json", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook post failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}

// Multi fans out an alert to multiple notifiers.
type Multi struct {
	Notifiers []Notifier
}

func NewMulti(notifiers ...Notifier) *Multi {
	return &Multi{Notifiers: notifiers}
}

func (m *Multi) Send(a Alert) error {
	var errs []string
	for _, n := range m.Notifiers {
		if err := n.Send(a); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("alert errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
