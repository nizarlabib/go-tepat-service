# Go Starter Service

A secure, production-ready Go backend service built with Chi router, PostgreSQL, Redis, and MinIO. Provides authentication (JWT), user management, file storage, and role-based access control.

## Table of Contents

- [Project Overview](#project-overview)
- [Prerequisites](#prerequisites)
- [Installation & Setup](#installation--setup)
- [Environment Variables](#environment-variables)
- [Project Structure](#project-structure)
- [Authentication](#authentication)
- [API Overview](#api-overview)
- [Running Tests](#running-tests)
- [Development Guidelines](#development-guidelines)
- [Using as Project Starter](#using-as-project-starter)
- [Deployment](#deployment)
- [License](#license)

## Project Overview

**Go Starter Service** is a Go-based backend API service designed for secure user authentication and file management.

### Main Features

- **Authentication**: JWT-based access and refresh tokens with HS256 signing
- **Token Revocation**: Redis-based JTI blocklist for immediate token invalidation
- **Password Security**: bcrypt password hashing with cost factor 12
- **CSRF Protection**: Token-based CSRF validation on all state-changing endpoints
- **Rate Limiting**: IP-based rate limiting to prevent abuse
- **User Management**: Registration, profile updates, role switching, account deletion
- **File Storage**: MinIO integration with AES-256 encryption at rest
- **Health Checks**: `/health` and `/ready` endpoints for Kubernetes probes
- **Security Headers**: HSTS, X-Content-Type-Options, X-Frame-Options, CSP
- **CORS**: Configurable CORS with explicit allowed origins

### Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26 |
| HTTP Router | Chi | v5.3.2 |
| Database | PostgreSQL (GORM) | v1.31.2 |
| Cache | Redis | v9.23.0 |
| Object Storage | MinIO | v7.3.0 |
| JWT | golang-jwt/jwt/v5 | v5.2.0 |
| Config | Viper | v1.21.0 |
| Validation | go-playground/validator/v10 | v10.30.5 |
| Logging | log/slog | stdlib |
| Error Tracking | Sentry | v0.50.0 |

### Architecture

The service follows a Clean Architecture-inspired layered pattern:

```
cmd/                    — Entry points (api, migration)
internal/
  api/                  — API service layer (thin wrappers over core)
  core/                 — Business logic (auth, user, file, storage)
  entity/               — Domain models (GORM entities)
  framework/
    primary/rest/       — Chi HTTP handlers
    secondary/repository/ — GORM data access
    secondary/storage/    — MinIO storage adapter
  infrastructure/       — Postgres, Redis, Viper, Sentry, MinIO clients
  middleware/           — Auth, refresh token middleware
  pkg/                  — Health, cache, rate limiter, RBAC, encryption, security
  spesial/              — JWT, claims, HTTP request info, Sentry user middleware
  value_object/         — DTOs, validators, responses, error handling, pagination
```

Dependency direction: handlers → API → core → repository → infrastructure.

## Prerequisites

- **Go**: 1.26.0 or higher
- **PostgreSQL**: 16 or higher
- **Redis**: 7 or higher
- **MinIO**: Latest (for file storage)
- **Development Tools**: `go`, `docker`, `docker-compose`

## Installation & Setup

### 1. Clone the repository

```bash
git clone <repository-url>
cd go-starter-service
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure environment variables

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` with your local configuration values. **Never commit real credentials.**

### 4. Database setup

Ensure PostgreSQL is running and create the database:

```sql
CREATE DATABASE starter_db;
```

### 5. Database migration

Run the migration to create all tables:

```bash
go run ./cmd/migration
```

Or use Docker Compose (which runs migrations automatically):

```bash
docker-compose up -d
```

### 6. Running the application locally

```bash
go run ./cmd/api
```

The server will start on port 8080 (or the port specified by `HTTP_PORT`).

### Using Docker Compose

```bash
docker-compose up -d
```

This starts PostgreSQL, Redis, MinIO, and the API service.

## Environment Variables

| Variable | Description | Required | Default |
|-----------|-------------|----------|---------|
| `APP_ENV` | Application environment (local, staging, production) | No | `local` |
| `APP_VERSION` | Application version | No | `local` |
| `APP_SERVICE_NAME` | Service name for logging and tracing | No | `starter-service` |
| `HTTP_PORT` | HTTP server port | No | `8080` |
| `POSTGRES_CONNECTION` | PostgreSQL connection string | Yes | — |
| `REDIS_CONNECTION` | Redis connection URL | Yes | — |
| `MINIO_ENDPOINT` | MinIO endpoint | Yes | — |
| `MINIO_ACCESS_KEY` | MinIO access key | Yes | — |
| `MINIO_SECRET_KEY` | MinIO secret key | Yes | — |
| `MINIO_BUCKET` | MinIO bucket name | Yes | — |
| `MINIO_REGION` | MinIO region | No | `ap-southeast-1` |
| `MINIO_USE_SSL` | Use SSL for MinIO | No | `false` |
| `JWT_KEY` | JWT signing key (must be secure and random) | Yes | — |
| `AES_KEY` | AES encryption key (base64-encoded) | Yes | — |
| `JWT_ACCESS_TOKEN_EXPIRY` | Access token expiry in minutes | No | `15` |
| `JWT_REFRESH_TOKEN_EXPIRY` | Refresh token expiry in minutes | No | `10080` |
| `AUTH_LOGIN_MAX_ATTEMPTS` | Max failed login attempts before lockout | No | `5` |
| `DISABLE_SINGLE_SESSION` | Disable single session enforcement | No | `false` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins | No | `*` |
| `OTP_DURATION` | OTP/token duration in seconds | No | `1800` |
| `PASSWORD_RESET_TTL` | Password reset token TTL in seconds | No | `900` |
| `SENTRY_DSN` | Sentry DSN for error tracking | No | — |
| `SENTRY_ENABLE_TRACKING` | Enable Sentry performance tracking | No | `false` |
| `SENTRY_TRACES_SAMPLE_RATE` | Sentry traces sample rate | No | `0.1` |

> **Warning**: `JWT_KEY` and `AES_KEY` have no default values and must be set via environment variables. Using predictable keys compromises all JWT tokens and encrypted data.

## Project Structure

```
project-root/
├── cmd/
│   ├── api/              — HTTP server entry point
│   ├── migration/        — Database migration tool
│   └── seed/             — Seeders (placeholder)
├── internal/
│   ├── api/              — API service interfaces + implementations
│   ├── core/             — Business logic layer
│   ├── entity/           — GORM domain models
│   ├── framework/
│   │   ├── primary/rest/ — Chi HTTP handlers
│   │   └── secondary/    — Repository and storage adapters
│   ├── infrastructure/   — External service clients (Postgres, Redis, MinIO, Sentry, Viper)
│   ├── middleware/       — HTTP middleware (auth, refresh token)
│   ├── pkg/              — Cross-cutting concerns (health, rate limiter, RBAC, encryption, security)
│   ├── spesial/          — JWT, claims, request info
│   └── value_object/     — DTOs, validators, responses, pagination
├── docs/
│   ├── GO_BACKEND_AUDIT.md — Security audit report with implementation status
│   └── openapi.yaml      — OpenAPI 3.0 specification
├── Dockerfile            — Multi-stage Docker build
├── docker-compose.yml    — Local development environment
├── .env.example          — Example environment variables
├── go.mod                — Go module definition
├── go.sum                — Go module checksums
└── README.md             — This file
```

## Authentication

### Mechanism

The service uses JWT (JSON Web Tokens) for authentication. Two token types are issued:

- **Access Token**: Short-lived (default 15 minutes), used for API authentication
- **Refresh Token**: Long-lived (default 7 days), used to obtain new access tokens

### Login Flow

1. Client obtains a CSRF token via `GET /api/v1/auth/csrf_token`
2. Client sends `POST /api/v1/auth/login` with credentials and `X-CSRF-Token` header
3. Server validates credentials, creates a device session, and returns access + refresh tokens
4. Client stores tokens and includes the access token in the `Authorization: Bearer <token>` header for subsequent requests

### Token Handling

- Access tokens are signed with HS256 and include a JTI (JWT ID) claim
- Refresh tokens are also signed with HS256 and validated against the database
- Token revocation is implemented via a Redis-based JTI blocklist
- Logout adds the access token JTI to the blocklist
- Refresh token rotation invalidates the old refresh token JTI

### Protected Endpoints

All endpoints under `/api/v1` except auth public routes require a valid Bearer token:

```
Authorization: Bearer <access_token>
```

### Authorization

Role-based access control (RBAC) is implemented using bitmask permissions. The `/auth/impersonate/{user_id}` endpoint requires super admin privileges.

## API Overview

### Base URL

- Local: `http://localhost:8080/api/v1`
- Production: `https://api.example.com/api/v1`

### Endpoint Groups

| Group | Base Path | Authentication |
|-------|-----------|----------------|
| Authentication | `/auth` | Mixed (public + protected) |
| Users | `/users` | Bearer token + RBAC |
| Files | `/files` | Bearer token |
| System | `/ping`, `/health`, `/ready` | None |

### Response Format

All API responses follow a consistent structure:

```json
{
  "message": "Human-readable message",
  "data": { ... },
  "meta": {
    "code": 200,
    "error_code": "SUCCESS"
  }
}
```

Error responses:

```json
{
  "message": "Error description",
  "error_code": "ERR_...",
  "status": 400
}
```

### Pagination

List endpoints support pagination via `page` and `limit` query parameters:

```
GET /api/v1/users?page=1&limit=10
```

The maximum limit is 50 items per page.

### Swagger UI

Interactive API documentation is available at:

- Local: `http://localhost:8080/swagger/index.html`
- OpenAPI spec: `http://localhost:8080/swagger/openapi.yaml`

Swagger UI is only available in non-production environments (`APP_ENV != production`).

## Running Tests

### Unit Tests

```bash
go test ./... -count=1
```

### Race Detection

```bash
go test -race ./... -count=1
```

### Static Analysis

```bash
go vet ./...
```

### Code Formatting

```bash
gofmt -l .    # Check formatting
gofmt -w .    # Fix formatting
```

## Development Guidelines

### Code Formatting

Use `gofmt` for all Go source files. The project follows standard Go formatting conventions.

### Naming Conventions

- Use camelCase for variables and functions
- Use PascalCase for exported types and functions
- Package names should be lowercase, single words

### Adding New Endpoints

1. Define the handler in `internal/framework/primary/rest/`
2. Register the route in the corresponding `Start()` method
3. Add middleware (auth, CSRF, RBAC) as needed
4. Document the endpoint in `docs/openapi.yaml`
5. Add tests for the new endpoint

### Adding Middleware

1. Create middleware in `internal/pkg/middleware/`
2. Register it in `cmd/api/main.go` middleware stack
3. Ensure it follows the `func(http.Handler) http.Handler` signature

### Database Migration Workflow

1. Migration runs via `cmd/migration/main.go` using GORM AutoMigrate
2. Run `go run ./cmd/migration` to apply migrations
3. For production, run migrations as part of the deployment pipeline

## Using as Project Starter

This service is designed to be used as a template for your new Go projects. Follow these steps to scaffold a new project from `go-starter-service`.

### 1. Clone and Rename the Module

```bash
git clone <this-repo-url> my-new-project
cd my-new-project
```

Update the Go module name in `go.mod` to match your project:

```go
module github.com/your-org/my-new-project
```

Then update all import paths throughout the codebase:

```bash
# Example: replace go-starter-service with your module path
find . -name "*.go" -exec sed -i 's|go-starter-service|github.com/your-org/my-new-project|g' {} +
```

### 2. Update Project Identity

| File | What to Change |
|------|---------------|
| `go.mod` | `module` path |
| `internal/infrastructure/viper.go` | `APP_SERVICE_NAME` default |
| `internal/core/file.go` | MinIO user tags |
| `cmd/api/main.go` | Swagger UI HTML title |
| `docs/openapi.yaml` | API `info.title` and `info.description` |
| `README.md` | Project title, description, and examples |
| `.env.example` | `APP_SERVICE_NAME` default |

### 3. Configure Environment Variables

Copy `.env.example` to `.env` and update for your environment:

```bash
cp .env.example .env
```

Key variables to adjust:

| Variable | Description |
|----------|-------------|
| `APP_ENV` | `local`, `staging`, or `production` |
| `APP_SERVICE_NAME` | Your service name (used in logging, JWT issuer) |
| `APP_VERSION` | Your application version |
| `HTTP_PORT` | Port to listen on |
| `POSTGRES_CONNECTION` | Your PostgreSQL connection string |
| `REDIS_CONNECTION` | Your Redis connection URL |
| `MINIO_ENDPOINT` | Your MinIO endpoint |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | MinIO credentials |
| `MINIO_BUCKET` | MinIO bucket name |
| `JWT_KEY` | **Generate a strong random key** (required) |
| `AES_KEY` | **Generate a strong random key** (required) |
| `CORS_ALLOWED_ORIGINS` | Comma-separated allowed frontend origins |

**Never commit real secrets.** Use environment variables or a secrets manager in production.

### 4. Generate Cryptographic Keys

Generate secure keys for JWT and AES:

```bash
# JWT key (32+ random bytes, base64)
openssl rand -base64 32

# AES key (32 bytes, base64)
openssl rand -base64 32
```

Set these as `JWT_KEY` and `AES_KEY` in your `.env`.

### 5. Update Database Configuration

- `cmd/migration/main.go`: Update PostgreSQL credentials if not using environment variables
- `docker-compose.yml`: Update `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, and `MINIO_BUCKET`
- `.github/workflows/ci-cd.yml`: Update CI database credentials

### 6. Review and Customize Business Logic

| Area | Files to Review |
|------|----------------|
| Authentication | `internal/core/auth.go`, `internal/core/auth_login.go`, `internal/middleware/auth.go` |
| JWT handling | `internal/spesial/jwt.go` |
| User repository | `internal/framework/secondary/repository/user.go` |
| File storage | `internal/framework/secondary/storage/minio.go` |
| API handlers | `internal/framework/primary/rest/auth.go`, `rest/user.go`, `rest/file.go` |
| API layer | `internal/api/auth.go`, `api/user.go`, `api/file.go` |
| Entity models | `internal/entity/user.go`, `entity/auth.go`, `entity/file.go`, `entity/device_log.go` |
| Validators | `internal/value_object/v3/validator/`, `internal/value_object/v1/validator/` |

### 7. Remove or Customize Example Code

- `internal/api/auth.go` contains several stub methods (`DeleteAccount`, `GetSessions`, `UpdateSession`, `UpdateUser`, `ImpersonateUser`, `ReleaseImpersonation`, `SwitchRole`, `UpdateProfile`, `UpdateProfilePhoto`, `UpdateUserPasswordExpiration`, `ValidateCsrfToken`, `VerifyUserPassword`) that return `nil`. Implement these based on your requirements.
- Commented-out code in `rest/user.go`, `api/auth.go`, `cmd/migration/main.go` can be removed or restored as needed.

### 8. Run and Verify

```bash
# Install dependencies
go mod download

# Format code
gofmt -w .

# Vet code
go vet ./...

# Build
go build ./...

# Run tests
go test ./...

# Run migration
go run ./cmd/migration

# Start the server
go run ./cmd/api
```

### 9. Docker Deployment

```bash
# Build image
docker build -t my-new-project:latest .

# Or use docker-compose
docker-compose up -d
```

Update `docker-compose.yml` environment variables before deploying.

### 10. CI/CD

The `.github/workflows/ci-cd.yml` workflow runs `gofmt`, `go vet`, `go test`, and `go test -race` on every push/PR. Update the `go-version` and environment variables as needed.

### Quick Checklist

- [ ] `go.mod` module path updated
- [ ] All import paths updated (`go-starter-service` → your module)
- [ ] `APP_SERVICE_NAME` updated in `viper.go` and `.env.example`
- [ ] `JWT_KEY` and `AES_KEY` generated and set (never hardcoded)
- [ ] Database credentials updated in `docker-compose.yml` and CI
- [ ] MinIO bucket name updated
- [ ] CORS allowed origins updated for your frontend
- [ ] API title/description updated in `docs/openapi.yaml`
- [ ] README updated with your project details
- [ ] Stub methods in `api/auth.go` implemented or removed
- [ ] Tests added for your custom business logic


### Production Build

```bash
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/api
```

### Environment Configuration

Set all required environment variables from the table above. Use a secrets manager (AWS Secrets Manager, HashiCorp Vault, or Kubernetes secrets) for `JWT_KEY` and `AES_KEY` in production.

### Database Migration

Run migrations before deploying:

```bash
go run ./cmd/migration
```

### Running the Application

```bash
./server
```

### Health Check Endpoints

- `GET /health` — Detailed health check including PostgreSQL and Redis
- `GET /ready` — Simple readiness check
- `GET /ping` — Heartbeat endpoint

## License

The license for this project is defined in the LICENSE file (if present). No specific license is declared in this repository.
