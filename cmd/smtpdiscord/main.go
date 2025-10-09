package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/mhale/smtpd"
	"github.com/tris203/smtpdiscord/internal/db"
	"github.com/tris203/smtpdiscord/internal/smtp"
	"github.com/tris203/smtpdiscord/internal/web"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default values")
	}

	// Get DB path from env, default to "config.db"
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "config.db"
	}

	database, err := db.InitDB(dbPath)
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
