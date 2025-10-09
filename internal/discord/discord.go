package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type WebhookContent struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

func SendToWebhook(webhookURL string, payload WebhookContent) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending to Discord webhook %s: %v", webhookURL, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		log.Printf("Discord webhook %s returned status %d", webhookURL, resp.StatusCode)
		return errors.New("discord webhook returned non-success status")
	}

	log.Printf("Successfully sent to Discord webhook")
	return nil
}
