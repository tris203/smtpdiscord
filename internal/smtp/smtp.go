package smtp

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/mail"
	"strings"

	"github.com/mhale/smtpd"
	"github.com/tris203/smtpdiscord/internal/discord"
)

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
		payload := discord.WebhookContent{
			Content:  content,
			Username: from,
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

				err = discord.SendToWebhook(webhook, payload)
				if err != nil {
					return err
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

func Start(db *sql.DB) error {
	log.Println("Starting SMTP server on :25")
	return smtpd.ListenAndServe(":25", MailHandler(db), "smtpdiscord", "")
}
