# SnapURL - URL Shortener Service

A secure, rate-limited URL shortener with SQLite persistence, built in Go.

---

## 🚀 Features

- 🔐 Cryptographically secure short codes (base62)
- ⏳ Expiry support and manual revocation
- 📊 Live `/metrics` endpoint (created, active, redirects)
- 🧠 IP-based rate limiting using token buckets
- 📁 Embedded auto-migrations (SQLite)
- 🐳 Docker-ready & deployable with Caddy

---

## 🛠️ Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/valorm/snapurl.git
cd snapurl
```

### 2. Configure environment variables

Copy `.env.example` to `.env` and adjust if needed:

```env
PORT=:8080
DB_PATH=data/snapurl.db
RATE_LIMIT=100
API_KEYS=default_key_1,default_key_2
```

---

### 3. Run the server locally

```bash
go run ./cmd/server/main.go
```

---

### 4. Test via `curl`

```bash
# Create short URL
curl -X POST http://localhost:8080/shorten   -H "Content-Type: application/json"   -d '{"url": "https://example.com"}'

# Access short URL
curl -v http://localhost:8080/<shortcode>

# Revoke short URL (requires API key)
curl -X DELETE http://localhost:8080/<shortcode>   -H "X-API-Key: default_key_1"

# Get metrics
curl http://localhost:8080/metrics

# Check health
curl http://localhost:8080/health
```

---

## 🌐 API Endpoints

| Method | Path             | Description                  | Auth Required |
|--------|------------------|------------------------------|---------------|
| POST   | `/shorten`       | Create a short URL           | ❌            |
| GET    | `/{shortcode}`   | Redirect to original URL     | ❌            |
| DELETE | `/{shortcode}`   | Revoke an existing short URL | ✅            |
| GET    | `/health`        | Health check                 | ❌            |
| GET    | `/metrics`       | Metrics (JSON)               | ❌            |

---

## 📦 Docker

### Build the image

```bash
docker build -t snapurl .
```

### Run the container

```bash
docker run -p 8080:8080 -v snapurl_data:/data snapurl
```

---

## ⚙️ Caddyfile Example (HTTPS Proxy)

```caddyfile
yourdomain.com {
    reverse_proxy localhost:8080
}
```

---

## 🧪 End-to-End Workflow

1. POST `/shorten`  
2. GET `/shortcode` (should 302)  
3. DELETE `/shortcode` (with API key)  
4. GET `/shortcode` (should return 410 Gone)  
5. GET `/metrics` (shows usage)  

---

## 🧪 Run Tests

```bash
go test ./internal/...
```

---

## 🤝 Contributing

1. Fork this repository
2. Create your feature branch: `git checkout -b feature/feature-name`
3. Commit your changes: `git commit -m 'Add new feature'`
4. Push to the branch: `git push origin feature/feature-name`
5. Open a pull request 🚀

---

## 📄 License

MIT License – see the [`LICENSE`](./LICENSE) file for details.
