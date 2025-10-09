# smtpdiscord

A lightweight SMTP → Discord bridge. Receive emails on a local SMTP server and forward them to Discord channels via configured webhooks.

**Quick summary**

- Runs an SMTP server on `:25` and a simple web UI on `:8080`.
- Stores domain → webhook mappings in a local SQLite DB (`config.db`).
- Intended for local / internal use and experimentation—not production-ready as-is.

**Why this exists**

This project is a small, pragmatic tool to pipe inbound email into Discord channels for notifications, debugging, or just having fun. It was iterated quickly and "full vibe coded" — heavy LLM assistance informed structure, wording, and implementation choices. The README leans into that vibe while keeping instructions practical.

**Features**

- Accepts SMTP messages and forwards them to Discord webhooks selected by recipient domain.
- Simple web UI for adding/removing domain → webhook mappings, powered by `htmx` and server templates.
- Single-file, easy-to-read Go components so you can hack on it quickly.

**Ports & files**

- **SMTP:** `:25`
- **Web UI:** `:8080`
- **DB file:** `config.db` (created in the working directory, configurable via DB_PATH env var)

Getting started

- Build and run locally:

```bash
go build -o smtpdiscord ./cmd/smtpdiscord
./smtpdiscord
```

- Or with Docker:

```bash
docker build -t smtpdiscord .
docker run -p 25:25 -p 8080:8080 smtpdiscord
```

Configuration

- Create a `.env` file based on `.env.example` to set the database path (optional, defaults to `config.db`).
- Open the web UI at `http://localhost:8080` and add domain/webhook pairs.
- Each incoming recipient address is split on `@`; the domain part is used to look up the webhook.
- Example: an email to `alerts@example.com` will use the webhook configured for the `example.com` domain.

Important notes & security

- The SMTP server binds to port `25` by default. Do not expose this to the public internet without adding authentication, TLS, and rate limiting.
- Discord webhooks are secrets. Keep your `config.db` safe and do not commit webhook URLs to public repos.
- Payloads are posted directly to the configured webhook URL. There is no retry/backoff logic beyond the HTTP POST, and incoming message validation is minimal.

Implementation details (short)

- SMTP handling: `internal/smtp` parses the raw email, extracts Subject and body, builds a Discord webhook payload, and POSTs to the matching webhook(s).
- Web UI: `internal/web` serves a tiny HTML page and endpoints to list/add/delete domain mappings (uses `htmx` for convenience).
- DB: `internal/db` initializes an SQLite DB and ensures a `domains` table exists with `domain` and `webhook_url` columns.

LLM / "full vibe coded" tell-tales (what to expect)

- **Inline templates:** HTML served as raw string literals in Go source rather than separate template files — quick scaffolding choice.
- **Minimal validation:** code focuses on the happy path and returns early on missing config rather than trying many recovery paths.
- **Simple, explicit wiring:** clearly separated packages (`internal/smtp`, `internal/web`, `internal/db`) with a small `main` to start both servers — typical of a generated starter structure.
- **Opinionated defaults:** ports `25` and `8080`, DB filename `config.db` are hard-coded for convenience.
- **Compact error handling:** errors are logged and returned; there is no elaborate metrics / retry / backoff system.

These are not criticisms — they’re honest signals that the project was built quickly and assisted by an LLM. They make the code easy to read and improve.

Roadmap & suggestions

- Add TLS and authentication for SMTP (do NOT expose plain SMTP to the open internet).
- Add webhook validation (verify URLs look like Discord webhooks and optionally test them on add).
- Add retry/backoff for webhook delivery and logging of failed deliveries.
- Move HTML templates to separate files if you plan to expand the UI.

Contributing

License

MIT
