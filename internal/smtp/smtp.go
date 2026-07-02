package smtp

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"log"
	"net"
	"net/mail"
	"strings"

	"github.com/mhale/smtpd"
	"github.com/tris203/smtpdiscord/internal/discord"
)

const (
	discordEmbedTitleLimit       = 256
	discordEmbedDescriptionLimit = 4096
	discordEmbedFieldValueLimit  = 1024
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

		bodyText := strings.TrimSpace(string(body))
		if bodyText == "" {
			bodyText = "(empty email body)"
		}

		if strings.TrimSpace(subject) == "" {
			subject = "(no subject)"
		}
		subject = truncate(subject, discordEmbedTitleLimit)
		description := truncate(bodyText, discordEmbedDescriptionLimit)
		toField := truncate(strings.Join(to, "\n"), discordEmbedFieldValueLimit)
		fromField := truncate(from, discordEmbedFieldValueLimit)

		payload := discord.WebhookContent{
			Username: from,
			Embeds: []discord.WebhookEmbed{
				{
					Title:       subject,
					Description: description,
					Color:       0x5865F2,
					Fields: []discord.WebhookEmbedField{
						{Name: "From", Value: fromField, Inline: true},
						{Name: "To", Value: toField, Inline: true},
					},
					Footer: &discord.WebhookEmbedFooter{Text: "smtpdiscord"},
				},
			},
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

func truncate(value string, limit int) string {
	const suffix = "... truncated"

	if limit <= 0 {
		return ""
	}

	if limit <= len(suffix) {
		return suffix[:limit]
	}

	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit-len(suffix)-1]) + " " + suffix
}

func Start(db *sql.DB, addr string) error {
	log.Printf("Starting SMTP server on %s", addr)
	return smtpd.ListenAndServe(addr, MailHandler(db), "smtpdiscord", "")
}
