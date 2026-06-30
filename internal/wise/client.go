package wise

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://wise.com/rates/history"

// rawRate mirrors the JSON structure from the Wise API.
// time is a Unix millisecond timestamp.
type rawRate struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Value  float64 `json:"value"`
	TimeMs int64   `json:"time"`
}

type Rate struct {
	Source string
	Target string
	Value  float64
	Time   time.Time
}

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 15 * time.Second}, baseURL: baseURL}
}

// NewClientWithBaseURL creates a Client pointing at a custom base URL (useful in tests).
func NewClientWithBaseURL(base string) *Client {
	return &Client{http: &http.Client{Timeout: 15 * time.Second}, baseURL: base}
}

// FetchHistory retrieves hourly rates for the given currency pair over the last `length` days.
func (c *Client) FetchHistory(source, target string, length int) ([]Rate, error) {
	url := fmt.Sprintf("%s?source=%s&target=%s&length=%d&unit=day&resolution=hourly", c.baseURL, source, target, length)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("wise: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wise: unexpected status %d", resp.StatusCode)
	}

	var raw []rawRate
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("wise: decode failed: %w", err)
	}

	rates := make([]Rate, len(raw))
	for i, r := range raw {
		rates[i] = Rate{
			Source: r.Source,
			Target: r.Target,
			Value:  r.Value,
			Time:   time.UnixMilli(r.TimeMs).UTC(),
		}
	}
	return rates, nil
}
