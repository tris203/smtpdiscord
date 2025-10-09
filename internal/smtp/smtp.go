package smtp

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"strings"
)

type WebhookContent struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

func MailHandler(db *sql.DB) func(net.Addr, string, []string, []byte) error {
	return func(remoteAddr net.Addr, from string, to []string, data []byte) error {
		log.Printf("Received email from %s to %v", from, to)
		// Parse the email
		msg, err := mail.ReadMessage(bytes.NewReader(data))
		if err != nil {
			log.Printf("Error parsing email: %v", err)
			return err
		}

		subject := msg.Header.Get("Subject")
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			return err
		}

		// Prepare payload
		content := fmt.Sprintf("**Subject:** %s\n\n%s", subject, string(body))
		payload := WebhookContent{
			Content:  content,
			Username: from,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			return err
		}

		// Send to webhooks for each recipient domain
		for _, recipient := range to {
			parts := strings.Split(recipient, "@")
			if len(parts) == 2 {
				domain := parts[1]
				var webhook string
				err := db.QueryRow("SELECT webhook_url FROM domains WHERE domain = ?", domain).Scan(&webhook)
				if err != nil {
					if err == sql.ErrNoRows {
						log.Printf("No webhook configured for domain %s", domain)
						return errors.New("no webhook configured for domain")
					} else {
						log.Printf("Error querying webhook for domain %s: %v", domain, err)
						return err
					}
				}

				resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("Error sending to Discord webhook %s: %v", webhook, err)
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != 204 && resp.StatusCode != 200 {
					log.Printf("Discord webhook %s returned status %d", webhook, resp.StatusCode)
					return errors.New("discord webhook returned non-success status")
				}
				log.Printf("Forwarded email from %s to Discord webhook for domain", from)
			} else {
				log.Printf("Invalid recipient email: %s", recipient)
				return errors.New("invalid recipient email")
			}
		}

		return nil
	}
}
