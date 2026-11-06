# Go Memolab Service

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
- [Deployment](#deployment)
- [License](#license)

## Project Overview

**Go Memolab Service** is a Go-based backend API service designed for secure user authentication and file management.

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
cd go-memolab-service
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
CREATE DATABASE memolab;
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
| `APP_SERVICE_NAME` | Service name for logging and tracing | No | `memolab-service` |
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

## Deployment

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
