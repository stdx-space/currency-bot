package wise_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stdx-space/currency-bot/internal/wise"
)

func TestFetchHistory(t *testing.T) {
	fixture, err := os.ReadFile("testdata/rates_hkd_cad_30d.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query params forwarded correctly.
		q := r.URL.Query()
		if got := q.Get("source"); got != "HKD" {
			t.Errorf("source = %q, want HKD", got)
		}
		if got := q.Get("target"); got != "CAD" {
			t.Errorf("target = %q, want CAD", got)
		}
		if got := q.Get("length"); got != "30" {
			t.Errorf("length = %q, want 30", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(fixture)
	}))
	defer srv.Close()

	client := wise.NewClientWithBaseURL(srv.URL)
	rates, err := client.FetchHistory("HKD", "CAD", 30)
	if err != nil {
		t.Fatalf("FetchHistory: %v", err)
	}

	t.Run("entry count", func(t *testing.T) {
		if got, want := len(rates), 720; got != want {
			t.Errorf("len(rates) = %d, want %d", got, want)
		}
	})

	t.Run("first entry", func(t *testing.T) {
		r := rates[0]
		if r.Source != "HKD" {
			t.Errorf("Source = %q, want HKD", r.Source)
		}
		if r.Target != "CAD" {
			t.Errorf("Target = %q, want CAD", r.Target)
		}
		if r.Value != 0.176031 {
			t.Errorf("Value = %v, want 0.176031", r.Value)
		}
		wantTime := time.UnixMilli(1780203600000).UTC()
		if !r.Time.Equal(wantTime) {
			t.Errorf("Time = %v, want %v", r.Time, wantTime)
		}
	})

	t.Run("last entry", func(t *testing.T) {
		r := rates[len(rates)-1]
		if r.Value != 0.181471 {
			t.Errorf("Value = %v, want 0.181471", r.Value)
		}
		wantTime := time.UnixMilli(1782792000000).UTC()
		if !r.Time.Equal(wantTime) {
			t.Errorf("Time = %v, want %v", r.Time, wantTime)
		}
	})
}

func TestFetchHistory_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := wise.NewClientWithBaseURL(srv.URL)
	_, err := client.FetchHistory("HKD", "CAD", 30)
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestFetchHistory_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := wise.NewClientWithBaseURL(srv.URL)
	_, err := client.FetchHistory("HKD", "CAD", 30)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
