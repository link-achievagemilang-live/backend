# Backend CI/CD

Automated testing and deployment pipeline for the URL Shortener backend.

## Workflows

### Main CI/CD Pipeline (`.github/workflows/ci.yml`)

Runs on every push to `main` or `develop` branches and on pull requests.

**Jobs:**

1. **Test** - Run unit and integration tests with PostgreSQL and Redis
   - Sets up Go 1.21
   - Runs tests with race detection
   - Generates coverage reports
   - Uploads coverage to Codecov

2. **Lint** - Code quality checks
   - Runs golangci-lint with configured rules
   - Checks for common issues and code smells

3. **Build** - Compile the application
   - Builds the server binary
   - Uploads binary as artifact

4. **Docker** - Build Docker image
   - Creates optimized Docker image
   - Uses build cache for faster builds

5. **Integration Tests** - End-to-end testing
   - Starts PostgreSQL and Redis
   - Runs database migrations
   - Starts the application
   - Executes integration test script

## Running Locally

### Unit Tests

```bash
go test -v ./...
```

### With Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Start services
docker compose up -d postgres redis

# Run migrations
PGPASSWORD=urlshortener psql -h localhost -U urlshortener -d urlshortener -f internal/database/migrations/001_init.sql

# Start application
go run cmd/server/main.go &

# Run tests
./test.sh
```

### Linting

```bash
golangci-lint run --timeout=5m
```

## Environment Variables

Required for CI/CD:

- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `REDIS_HOST` - Redis host
- `REDIS_PORT` - Redis port

## Secrets

Configure these in GitHub repository settings:

- `CODECOV_TOKEN` (optional) - For coverage reporting

## Deployment

The pipeline builds and tests the application but does not automatically deploy. For deployment:

1. Tag a release: `git tag v1.0.0 && git push --tags`
2. Use the built Docker image from the registry
3. Deploy to your infrastructure (Kubernetes, Docker Swarm, etc.)

## Badge

Add to your README:

```markdown
![CI](https://github.com/yourusername/url-shortener-backend/workflows/Backend%20CI%2FCD/badge.svg)
```
