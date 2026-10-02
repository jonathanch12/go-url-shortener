# Go URL Shortener

A small URL shortener written in Go. It accepts a long URL, generates a unique short code, and redirects visitors from the short URL back to the original. Each redirect is logged for basic analytics.

Built with [Gin](https://github.com/gin-gonic/gin) for HTTP routing, [GORM](https://gorm.io) with a PostgreSQL driver for persistence, and [godotenv](https://github.com/joho/godotenv) for configuration.

## Features

- **Create short URL** — `POST /shorten` with a long URL returns a unique 6-character base62 short code. If the long URL was already shortened, the existing code is returned instead of a new one.
- **Redirect** — `GET /:code` looks up the code and issues an HTTP 302 redirect to the original URL. Expired or unknown codes return 404.
- **Basic analytics** — every redirect is logged with the mapping ID, client IP, user agent, and timestamp.

## Project Structure

```
go-url-shortener/
├── main.go          # Entire application: models, handlers, routing, startup
├── go.mod           # Module definition and dependencies
├── go.sum           # Dependency checksums
├── .env             # Local config (not committed)
├── .env.example     # Template for the .env file
├── .gitignore
```

This is a learning project, so all application logic lives in a single `main.go`. It defines two data models:

- `URLMapping` — stores the `short_code` ↔ `long_url` mapping with creation and optional expiry timestamps.
- `URLClick` — one row per redirect, capturing IP, user agent, and time.

## Prerequisites

- Go 1.22+ (see `go.mod`)
- A running PostgreSQL instance with a database named `urlshortener`

## Setup

1. **Clone and enter the project**

   ```powershell
   git clone <repo-url>
   cd go-url-shortener
   ```

2. **Create your `.env`** by copying the example and filling in your Postgres password:

   ```powershell
   Copy-Item .env.example .env
   ```

   ```env
   DATABASE_URL=host=localhost user=postgres password=YOUR_PASSWORD dbname=urlshortener port=5432 sslmode=disable TimeZone=UTC
   BASE_URL=http://localhost:8080
   PORT=8080
   ```

3. **Install dependencies**

   ```powershell
   go mod download
   ```

## Running

```powershell
go run main.go
```

On startup the app connects to Postgres, auto-migrates the `url_mappings` and `url_clicks` tables, and listens on the port from `PORT` (default `8080`). You should see:

```
Connected to the database
Database migrated
Server listening on http://localhost:8080
```

## API

### Shorten a URL

```
POST /shorten
Content-Type: application/json

{ "long_url": "https://www.example.com/path/to/page" }
```

Response (`201 Created`, or `200 OK` if the URL was already shortened):

```json
{
  "short_url": "http://localhost:8080/Ab3xYz",
  "short_code": "Ab3xYz"
}
```

Example with curl:

```powershell
curl -X POST http://localhost:8080/shorten -H "Content-Type: application/json" -d '{"long_url":"https://www.example.com/path/to/page"}'
```

### Follow a short URL

```
GET /:code
```

Returns a `302 Found` redirect to the original URL, or `404 Not Found` if the code is unknown or expired. Open `http://localhost:8080/Ab3xYz` in a browser to be redirected.
