package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Footer      *footer `json:"footer,omitempty"`
}

type footer struct {
	Text string `json:"text"`
}

type payload struct {
	Embeds []embed `json:"embeds"`
}

// Send posts an embed message to the given Discord webhook URL.
func Send(webhookURL, title, description string, color int) error {
	p := payload{
		Embeds: []embed{
			{
				Title:       title,
				Description: description,
				Color:       color,
				Footer:      &footer{Text: fmt.Sprintf("currency-bot • %s", time.Now().UTC().Format(time.RFC1123))},
			},
		},
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("discord: marshal failed: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return nil
}
