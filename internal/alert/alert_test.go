package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/alert"
)

func makeAlert(level alert.Level) alert.Alert {
	return alert.Alert{
		JobName:   "backup",
		Level:     level,
		Message:   "job missed its window",
		Timestamp: time.Now(),
	}
}

func TestLogNotifier_Send(t *testing.T) {
	n := &alert.LogNotifier{}
	if err := n.Send(makeAlert(alert.LevelWarn)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wn := alert.NewWebhookNotifier(ts.URL)
	a := makeAlert(alert.LevelError)
	if err := wn.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["job"] != "backup" {
		t.Errorf("expected job=backup, got %q", received["job"])
	}
	if received["level"] != string(alert.LevelError) {
		t.Errorf("expected level=ERROR, got %q", received["level"])
	}
}

func TestWebhookNotifier_Send_NonOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	wn := alert.NewWebhookNotifier(ts.URL)
	if err := wn.Send(makeAlert(alert.LevelWarn)); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestMultiNotifier_AllSucceed(t *testing.T) {
	m := alert.NewMulti(&alert.LogNotifier{}, &alert.LogNotifier{})
	if err := m.Send(makeAlert(alert.LevelWarn)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultiNotifier_PartialFailure(t *testing.T) {
	badURL := "http://127.0.0.1:0/no-such-server"
	m := alert.NewMulti(&alert.LogNotifier{}, alert.NewWebhookNotifier(badURL))
	if err := m.Send(makeAlert(alert.LevelError)); err == nil {
		t.Fatal("expected error when one notifier fails")
	}
}
