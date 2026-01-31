# Go Backend Boilerplate

Production-ready Go backend with **Clean Architecture** and **SOLID principles**.

Built for developers who value clarity, security, and maintainability.

---

## ✨ Features

- ✅ Clean Architecture & SOLID principles
- 🔒 JWT Authentication (access + refresh tokens)
- 🚀 WebSocket support
- 🛡️ Rate limiting & CORS
- 📊 Structured JSON logging
- 🔐 Security-first design
- 📦 Modular & easily extensible

---

## 🚀 Quick Start

### 1. Prerequisites

- Go 1.23+ ([Download](https://golang.org/dl))

### 2. Install

```bash
git clone <your-repo-url>
cd go-boilerplate
go mod download
```

### 3. Run

```bash
# Set secrets (required)
export ACCESS_TOKEN_SECRET="your-secret-here"
export REFRESH_TOKEN_SECRET="your-secret-here"

# Start server
go run cmd/api/main.go
```

Server starts on `http://localhost:8080` 🎉

---

## 🧪 Test

Open another terminal:

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'

# Login
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

---

## 📁 Project Structure

```
go-boilerplate/
├── cmd/api/              # Application entry point
├── internal/
│   ├── server/           # Server setup & routing
│   ├── middleware/       # Auth, logging, rate limit, CORS
│   ├── modules/          # Domain modules (auth, health)
│   ├── realtime/         # WebSocket hub
│   ├── security/         # JWT & crypto
│   ├── config/           # Configuration
│   └── ...
├── go.mod
└── README.md
```

---

## 🔌 API Endpoints

### Public
- `GET /health` - Health check
- `POST /v1/auth/register` - Register user
- `POST /v1/auth/login` - Login
- `POST /v1/auth/refresh` - Refresh token
- `POST /v1/auth/logout` - Logout

### Protected (requires `Authorization: Bearer <token>`)
- `GET /v1/auth/me` - Get current user
- `GET /v1/ws` - WebSocket connection

---

## ⚙️ Configuration

Via environment variables:

```bash
# Required
ACCESS_TOKEN_SECRET=your-secret
REFRESH_TOKEN_SECRET=your-secret

# Optional (with defaults)
PORT=8080
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
RATE_LIMIT_RPM=60
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

Or create `.env` file from `env.example`.

---

## 🛠️ Development

```bash
# Run server
make run

# Build binary
make build

# Run tests
make test

# Format & lint
make lint
```

---

## 📦 Adding New Features

### New Module

```bash
mkdir -p internal/modules/posts
touch internal/modules/posts/{model,service,handler}.go
```

Wire it in `internal/server/server.go`

### Database Integration

```go
// Implement UserStore interface for your DB
type PostgresUserStore struct {
    db *sql.DB
}

func (s *PostgresUserStore) FindByEmail(email string) (*User, error) {
    // Your PostgreSQL query
}

// Inject in server.go
userStore := NewPostgresUserStore(db)
authService := auth.NewService(tokenManager, userStore)
```

---

## 🏗️ Architecture Principles

### Clean Architecture Layers

```
┌─────────────────────┐
│  HTTP/WebSocket     │  Handlers (translate requests/responses)
├─────────────────────┤
│  Business Logic     │  Services (domain rules)
├─────────────────────┤
│  Infrastructure     │  Security, logging, storage
└─────────────────────┘
```

### Key Rules

- **Handlers**: NO business logic, only HTTP translation
- **Services**: NO HTTP knowledge, only domain rules
- **Middleware**: NO business decisions, only cross-cutting concerns

### SOLID Principles

- **Single Responsibility**: Each file has one clear purpose
- **Open/Closed**: Extend via composition, not modification
- **Liskov Substitution**: Interfaces are fully replaceable
- **Interface Segregation**: Small, focused interfaces
- **Dependency Inversion**: Depend on abstractions, not implementations

---

## 🔒 Security

- JWT tokens (separate access & refresh)
- bcrypt password hashing
- IP-based rate limiting
- Strict CORS policy
- Safe error messages (no internal leaks)
- Token rotation on refresh

---

## 📊 WebSocket

```javascript
// Client-side
const ws = new WebSocket('ws://localhost:8080/v1/ws');
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log(message);
};
```

```go
// Server-side broadcasting
hub.Broadcast("notification", payload)
hub.SendToUser(userID, "message", payload)
```

---

## 🚀 Production Checklist

- [ ] Generate strong secrets (`openssl rand -base64 32`)
- [ ] Configure production CORS origins
- [ ] Adjust rate limits
- [ ] Set up HTTPS (nginx/caddy)
- [ ] Replace in-memory stores with database
- [ ] Configure logging aggregation
- [ ] Set up monitoring

---

## 📄 License

MIT License - use freely in your projects.

---

**Built with ❤️ for Go developers who value clean code.**
