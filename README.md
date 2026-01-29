# URL Shortener Backend

High-performance URL shortener service built with Go, PostgreSQL, and Redis.

![CI](https://github.com/link-achievagemilang-live/backend/workflows/Backend%20CI%2FCD/badge.svg)
[![codecov](https://codecov.io/gh/link-achievagemilang-live/backend/branch/main/graph/badge.svg)](https://codecov.io/gh/link-achievagemilang-live/backend)

## Features

- ⚡ **Lightning Fast** - Redis caching for sub-10ms response times
- 🔗 **Custom Aliases** - Create branded short links
- 📊 **Analytics** - Track clicks and access times
- 🔒 **Rate Limiting** - IP-based rate limiting (10 req/min)
- 🎯 **Base62 Encoding** - Efficient, collision-free short codes
- 🐳 **Docker Ready** - Complete containerization

## Quick Start

### With Docker

```bash
docker compose up -d
```

### Local Development

```bash
# Install dependencies
go mod download

# Run migrations
PGPASSWORD=urlshortener psql -h localhost -U urlshortener -d urlshortener -f internal/database/migrations/001_init.sql

# Run application
go run cmd/server/main.go
```

## API Endpoints

| Endpoint                         | Method | Description              |
| -------------------------------- | ------ | ------------------------ |
| `/api/v1/urls`                   | POST   | Create short URL         |
| `/{short_code}`                  | GET    | Redirect to original URL |
| `/api/v1/analytics/{short_code}` | GET    | Get analytics            |
| `/health`                        | GET    | Health check             |

### Create Short URL

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{
    "long_url": "https://example.com",
    "custom_alias": "my-link",
    "ttl_days": 30
  }'
```

## Environment Variables

```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
BASE_URL=http://localhost:8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=urlshortener
DB_PASSWORD=urlshortener
DB_NAME=urlshortener
REDIS_HOST=localhost
REDIS_PORT=6379
RATE_LIMIT_RPM=10
```

## Testing

```bash
# Unit tests
go test -v ./...

# With coverage
go test -coverprofile=coverage.out ./...

# Integration tests
./test.sh
```

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Router**: Chi
- **Architecture**: Clean Architecture

## License

MIT
