package web

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"
)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		s.indexHandler(w, r)
	case "/domains":
		switch r.Method {
		case "GET":
			s.listDomainsHandler(w, r)
		case "POST":
			s.addDomainHandler(w, r)
		case "DELETE":
			s.deleteDomainHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>SMTP Discord Config</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    <h1>Domain Configurations</h1>
    <div hx-get="/domains" hx-trigger="load, every 5s" hx-target="this" hx-swap="innerHTML">
        Loading...
    </div>
    <h2>Add Domain</h2>
    <form hx-post="/domains" hx-target="#domains-list" hx-swap="beforeend">
        <input type="text" name="domain" placeholder="Domain" required>
        <input type="url" name="webhook" placeholder="Webhook URL" required>
        <button type="submit">Add</button>
    </form>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

func (s *Server) listDomainsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT domain, webhook_url FROM domains")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var domains []struct {
		Domain  string
		Webhook string
	}
	for rows.Next() {
		var d struct {
			Domain  string
			Webhook string
		}
		err := rows.Scan(&d.Domain, &d.Webhook)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		domains = append(domains, d)
	}

	tmpl := `
<div id="domains-list">
{{range .}}
    <div>
        <strong>{{.Domain}}</strong>: {{.Webhook}}
        <button hx-delete="/domains?domain={{.Domain}}" hx-target="closest div" hx-swap="outerHTML">Delete</button>
    </div>
{{end}}
</div>
`
	t := template.Must(template.New("list").Parse(tmpl))
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, domains)
}

func (s *Server) addDomainHandler(w http.ResponseWriter, r *http.Request) {
	domain := strings.TrimSpace(r.FormValue("domain"))
	webhook := strings.TrimSpace(r.FormValue("webhook"))
	if domain == "" || webhook == "" {
		http.Error(w, "Domain and webhook required", http.StatusBadRequest)
		return
	}

	_, err := s.db.Exec("INSERT OR REPLACE INTO domains (domain, webhook_url) VALUES (?, ?)", domain, webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the new domain entry
	tmpl := `
<div>
    <strong>{{.Domain}}</strong>: {{.Webhook}}
    <button hx-delete="/domains?domain={{.Domain}}" hx-target="closest div" hx-swap="outerHTML">Delete</button>
</div>
`
	t := template.Must(template.New("item").Parse(tmpl))
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, struct {
		Domain  string
		Webhook string
	}{Domain: domain, Webhook: webhook})
}

func (s *Server) deleteDomainHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "Domain required", http.StatusBadRequest)
		return
	}

	_, err := s.db.Exec("DELETE FROM domains WHERE domain = ?", domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
