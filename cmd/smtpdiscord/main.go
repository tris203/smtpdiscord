package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/mhale/smtpd"
	"github.com/tris203/smtpdiscord/internal/db"
	"github.com/tris203/smtpdiscord/internal/smtp"
	"github.com/tris203/smtpdiscord/internal/web"
)

func main() {
	database, err := db.InitDB("config.db")
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer database.Close()

	var wg sync.WaitGroup

	// Start SMTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting SMTP server on :25")
		err := smtpd.ListenAndServe(":25", smtp.MailHandler(database), "smtpdiscord", "")
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Start web server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting web server on :8080")
		server := web.NewServer(database)
		err := http.ListenAndServe(":8080", server)
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}
