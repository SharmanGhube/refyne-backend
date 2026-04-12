# Contributing to Refyne Backend

**Project:** Refyne Platform Backend  
**Language:** Go 1.24+ · **Framework:** Gin · **Architecture:** Domain-Driven Design (DDD)  
**DI:** Google Wire · **Database:** PostgreSQL (pgx/sqlx) · **Queue:** River  
**Last Updated:** July 2025

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Structure](#project-structure)
3. [Domain Anatomy](#domain-anatomy)
4. [Adding a New Domain](#adding-a-new-domain)
5. [Adding a New Feature to an Existing Domain](#adding-a-new-feature-to-an-existing-domain)
6. [Dependency Injection (Wire)](#dependency-injection-wire)
7. [Routing Conventions](#routing-conventions)
8. [Handler Conventions](#handler-conventions)
9. [Service Conventions](#service-conventions)
10. [Repository Conventions](#repository-conventions)
11. [Error Handling](#error-handling)
12. [Middleware](#middleware)
13. [Database Migrations](#database-migrations)
14. [Testing](#testing)
15. [Build & Run](#build--run)
16. [Code Style & Linting](#code-style--linting)
17. [Commit & PR Guidelines](#commit--pr-guidelines)
18. [Quick Reference Cheat Sheet](#quick-reference-cheat-sheet)

---

## Architecture Overview

The backend follows **Domain-Driven Design (DDD)** with **Clean Architecture** layering:

```
HTTP Request
    │
    ▼
┌──────────┐
│  Router   │  internal/api/router.go — Assembles all domain routes
└────┬─────┘
     │
     ▼
┌──────────┐
│  Routes   │  domains/{domain}/routes/ — Defines endpoints, middleware, rate limits
└────┬─────┘
     │
     ▼
┌──────────┐
│ Handler   │  domains/{domain}/handler/ — HTTP request/response; input validation
└────┬─────┘
     │
     ▼
┌──────────┐
│ Service   │  domains/{domain}/services/ — Business logic; orchestration
└────┬─────┘
     │
     ▼
┌──────────┐
│Repository │  domains/{domain}/repository/ — Database access; SQL queries
└──────────┘
```

**Key principles:**
- Handlers **never** touch the database directly.
- Services contain all business logic and orchestrate repository calls.
- Repositories are the only layer that writes SQL.
- Each domain is self-contained with its own `wire.go`, `registry.go`, routes, handlers, services, repositories, models, and errors.

---

## Project Structure

```
refyne-backend/
├── cmd/
│   ├── main.go              # Application entry point
│   ├── wire.go              # Root Wire injector — imports all domain ProviderSets
│   └── wire_gen.go          # Auto-generated Wire code (DO NOT EDIT)
│
├── internal/
│   ├── api/
│   │   ├── router.go        # Assembles all domain routes into the Gin engine
│   │   ├── health.go        # Health check endpoints
│   │   ├── wire.go          # API-layer Wire providers
│   │   └── middlewares/     # Global HTTP middleware (auth, CORS, rate-limit, etc.)
│   │
│   ├── bootstrap/           # App lifecycle (start, stop, graceful shutdown)
│   ├── config/              # Configuration loading & validation
│   ├── database/            # DB connection, pool, migrations
│   │   └── migrations/      # SQL migration files (up/down)
│   ├── monitoring/          # Prometheus metrics
│   │
│   ├── domains/             # ★ Business domains — the core of the app
│   │   ├── auth/            # Authentication & security
│   │   ├── user/            # User profiles & settings
│   │   ├── workspace/       # Multi-workspace management
│   │   ├── subscription/    # Billing (Paddle integration)
│   │   ├── ai/              # AI features
│   │   ├── email/           # Email sending
│   │   ├── notification/    # Push/in-app notifications
│   │   ├── context/         # Context documents
│   │   └── otto/            # Otto AI assistant
│   │
│   └── shared/              # Cross-cutting concerns
│       ├── handlerRegistry/ # Central registry wiring all domain handlers
│       ├── redis/           # Redis client setup
│       ├── river/           # River job queue setup
│       ├── audit/           # Audit logging
│       ├── device/          # Device tracking
│       └── validation/      # Shared validation helpers
│
├── pkg/                     # Public/reusable packages
│   ├── error/               # AppError types & error codes
│   └── logging/             # Zap-based structured logging
│
├── tests/
│   └── e2e_test.go          # End-to-end tests
│
└── docs/                    # Documentation
    ├── strategy/            # Business strategy docs
    ├── memory/              # Session context & implementation logs
    └── temp/                # Working docs (folder structure, etc.)
```

---

## Domain Anatomy

Every domain under `internal/domains/{name}/` follows this structure:

```
domains/{name}/
├── wire.go           # Wire ProviderSet — exports all providers for this domain
├── registry.go       # Registry struct — bundles all handlers for the handler registry
├── handler/
│   ├── {name}_handler.go       # Handler interface
│   └── {name}_handler_impl.go  # Handler implementation
├── services/
│   ├── {name}_service.go       # Service interface (if applicable)
│   └── {name}_service_impl.go  # Service implementation
├── repository/
│   └── {name}_repository.go    # Repository struct with DB methods
├── models/
│   └── {name}.go               # Domain models / DTOs
├── routes/
│   └── {name}_routes.go        # Gin route registration
├── errors/
│   └── errors.go               # Domain-specific error definitions
└── config/                     # Domain-specific config (optional)
    └── config.go
```

### Example: Subscription Domain

```
domains/subscription/
├── wire.go                           # ProviderSet aggregating sub-packages
├── registry.go                       # SubscriptionRegistry wrapping handlers
├── config/
│   ├── paddle_config.go              # Paddle-specific configuration
│   └── wire.go
├── handler/
│   ├── subscription_handler.go       # Interface
│   ├── subscription_handler_impl.go  # Implementation
│   └── wire.go
├── services/
│   ├── paddle_service.go             # Paddle API service
│   ├── paddle_factory.go             # Mock/sandbox factory
│   ├── webhook_service_impl.go       # Webhook processing
│   └── wire.go
├── repository/
│   └── ...
├── models/
│   └── ...
├── routes/
│   └── subscription_routes.go
└── errors/
    └── errors.go
```

---

## Adding a New Domain

Follow these steps **in order**. Each step builds on the previous one.

### Step 1: Create the Directory Structure

```bash
mkdir -p internal/domains/billing/{handler,services,repository,models,routes,errors}
```

### Step 2: Define Models

```go
// internal/domains/billing/models/invoice.go
package models

import "time"

type Invoice struct {
    ID        string    `db:"id"          json:"id"`
    UserID    string    `db:"user_id"     json:"user_id"`
    Amount    int64     `db:"amount"      json:"amount"`
    Status    string    `db:"status"      json:"status"`
    CreatedAt time.Time `db:"created_at"  json:"created_at"`
}

type CreateInvoiceRequest struct {
    Amount int64  `json:"amount" binding:"required,gt=0"`
    UserID string `json:"user_id" binding:"required"`
}
```

### Step 3: Create the Repository

```go
// internal/domains/billing/repository/billing_repository.go
package repository

import (
    "context"
    "github.com/jmoiron/sqlx"
    "github.com/refynehq/refyne-backend/internal/domains/billing/models"
)

type BillingRepository struct {
    db *sqlx.DB
}

func NewBillingRepository(db *sqlx.DB) *BillingRepository {
    return &BillingRepository{db: db}
}

func (r *BillingRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
    query := `INSERT INTO invoices (id, user_id, amount, status, created_at)
              VALUES ($1, $2, $3, $4, NOW())`
    _, err := r.db.ExecContext(ctx, query, invoice.ID, invoice.UserID, invoice.Amount, invoice.Status)
    return err
}
```

### Step 4: Create the Service

```go
// internal/domains/billing/services/billing_service.go
package services

import (
    "context"
    "github.com/refynehq/refyne-backend/internal/domains/billing/models"
    "github.com/refynehq/refyne-backend/internal/domains/billing/repository"
)

type BillingService struct {
    repo *repository.BillingRepository
}

func NewBillingService(repo *repository.BillingRepository) *BillingService {
    return &BillingService{repo: repo}
}

func (s *BillingService) CreateInvoice(ctx context.Context, req *models.CreateInvoiceRequest) (*models.Invoice, error) {
    invoice := &models.Invoice{
        ID:     generateID(), // your ID generation logic
        UserID: req.UserID,
        Amount: req.Amount,
        Status: "pending",
    }
    if err := s.repo.CreateInvoice(ctx, invoice); err != nil {
        return nil, err
    }
    return invoice, nil
}
```

### Step 5: Create the Handler Interface + Implementation

```go
// internal/domains/billing/handler/billing_handler.go
package handler

import "github.com/gin-gonic/gin"

// BillingHandler defines the HTTP handler interface for billing
type BillingHandler interface {
    CreateInvoice(c *gin.Context)
    GetInvoice(c *gin.Context)
}
```

```go
// internal/domains/billing/handler/billing_handler_impl.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/refynehq/refyne-backend/internal/domains/billing/models"
    "github.com/refynehq/refyne-backend/internal/domains/billing/services"
)

type billingHandlerImpl struct {
    service *services.BillingService
}

func NewBillingHandler(service *services.BillingService) BillingHandler {
    return &billingHandlerImpl{service: service}
}

func (h *billingHandlerImpl) CreateInvoice(c *gin.Context) {
    var req models.CreateInvoiceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    invoice, err := h.service.CreateInvoice(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": invoice})
}

func (h *billingHandlerImpl) GetInvoice(c *gin.Context) {
    // implementation
}
```

### Step 6: Create Routes

```go
// internal/domains/billing/routes/billing_routes.go
package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/refynehq/refyne-backend/internal/api/middlewares"
    handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
)

func SetupBillingRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
    billingHandler := registry.Billing.BillingHandler

    billingGroup := router.Group("/billing")
    {
        protected := billingGroup.Group("")
        protected.Use(middlewares.AuthMiddleware())
        {
            protected.POST("/invoices", billingHandler.CreateInvoice)
            protected.GET("/invoices/:id", billingHandler.GetInvoice)
        }
    }
}
```

### Step 7: Create the Registry

```go
// internal/domains/billing/registry.go
package billing

import (
    "github.com/refynehq/refyne-backend/internal/domains/billing/handler"
)

type BillingRegistry struct {
    handler.BillingHandler
}

func NewBillingRegistry(billingHandler handler.BillingHandler) *BillingRegistry {
    return &BillingRegistry{
        BillingHandler: billingHandler,
    }
}
```

### Step 8: Create the Wire ProviderSet

```go
// internal/domains/billing/wire.go
package billing

import (
    "github.com/google/wire"
    "github.com/refynehq/refyne-backend/internal/domains/billing/handler"
    "github.com/refynehq/refyne-backend/internal/domains/billing/repository"
    "github.com/refynehq/refyne-backend/internal/domains/billing/services"
)

var ProviderSet = wire.NewSet(
    // Registry
    NewBillingRegistry,

    // Handlers
    handler.NewBillingHandler,

    // Services
    services.NewBillingService,

    // Repositories
    repository.NewBillingRepository,
)
```

### Step 9: Wire Into the Application (3 files to update)

**A. Add to the Handler Registry** (`internal/shared/handlerRegistry/registry.go`):

```go
// Add import
billing "github.com/refynehq/refyne-backend/internal/domains/billing"

// Add field to struct
type HandlerRegistry struct {
    // ... existing fields ...
    Billing *billing.BillingRegistry
}

// Add parameter to constructor
func NewHandlerRegistry(
    // ... existing params ...
    br *billing.BillingRegistry,
) *HandlerRegistry {
    return &HandlerRegistry{
        // ... existing assignments ...
        Billing: br,
    }
}
```

**B. Add to the root Wire injector** (`cmd/wire.go`):

```go
// Add import
billing "github.com/refynehq/refyne-backend/internal/domains/billing"

// Add to AppSet
var AppSet = wire.NewSet(
    // ... existing sets ...
    billing.ProviderSet,  // Add this line under "Domain Layer"
)
```

**C. Add routes to the router** (`internal/api/router.go`):

```go
// Add import
billing "github.com/refynehq/refyne-backend/internal/domains/billing/routes"

// Inside NewRouter function, add:
billing.SetupBillingRoutes(apiRoutes, registry)
```

### Step 10: Regenerate Wire and Build

```bash
cd cmd && wire
go build ./...
```

---

## Adding a New Feature to an Existing Domain

To add a new endpoint to an existing domain (e.g., adding "delete invoice" to subscription):

1. **Add the method** to the handler interface (`handler/{name}_handler.go`)
2. **Implement the method** in the handler impl (`handler/{name}_handler_impl.go`)
3. **Add business logic** to the service (or create a new service method)
4. **Add DB queries** to the repository (or create a new repository method)
5. **Register the route** in `routes/{name}_routes.go`
6. **No Wire changes needed** — Wire only needs updating if you add new provider constructors

---

## Dependency Injection (Wire)

### How Wire Works in This Project

```
cmd/wire.go (root injector)
    │
    ├── config.ProviderSet
    ├── database.ProviderSet
    ├── logging.ProviderSet
    ├── redis.ProviderSet
    ├── river.ProviderSet
    ├── handlerregistry.ProviderSet
    │
    ├── auth.ProviderSet          ◄── Each domain exports a ProviderSet
    ├── user.ProviderSet
    ├── workspace.ProviderSet
    ├── subscription.ProviderSet
    ├── ai.ProviderSet
    ├── email.ProviderSet
    ├── notification.ProviderSet
    ├── otto.ProviderSet
    ├── context.ProviderSet
    │
    ├── api.ProviderSet
    └── bootstrap.ProviderSet
```

### Rules for Wire

1. **Every constructor returns a concrete type or an interface** — Wire matches types.
2. **Handler constructors must return the interface type**, not the struct:
   ```go
   // ✅ Correct — returns interface
   func NewBillingHandler(svc *services.BillingService) BillingHandler { ... }
   
   // ❌ Wrong — returns concrete type when interface is expected
   func NewBillingHandler(svc *services.BillingService) *billingHandlerImpl { ... }
   ```
3. **If you have interface bindings**, use `wire.Bind`:
   ```go
   wire.Bind(new(handler.BillingHandler), new(*handler.billingHandlerImpl)),
   ```
4. **Never edit `wire_gen.go`** — it is auto-generated.
5. **Regenerate** after any Wire change: `cd cmd && wire`

### Provider extraction helpers

If config needs transformation before injection:
```go
// Extract a specific value from Config for injection
func ProvideFrontendURL(cfg *config.Config) string {
    return cfg.FrontendURL
}
```

---

## Routing Conventions

### URL Structure
```
/api/{domain}/{resource}
/api/{domain}/{resource}/:id
/api/{domain}/{resource}/:id/{sub-resource}
```

### Examples from the Codebase
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout              (protected)

POST   /api/subscription/checkout    (protected, rate-limited)
GET    /api/subscription/status      (protected)
POST   /api/subscription/portal      (protected)
POST   /api/webhooks/paddle          (public, signature-verified)

GET    /api/user/profile             (protected)
PUT    /api/user/profile             (protected)
POST   /api/user/onboarding          (protected)
```

### Route Setup Pattern
```go
func SetupXRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
    handler := registry.X.XHandler

    group := router.Group("/x")
    {
        // Public routes
        group.GET("/public-endpoint", handler.PublicMethod)

        // Protected routes
        protected := group.Group("")
        protected.Use(middlewares.AuthMiddleware())
        {
            protected.GET("/private-endpoint", handler.PrivateMethod)
        }
    }
}
```

---

## Handler Conventions

- **Interface-based**: Define an interface, implement separately.
- **Input validation**: Use `c.ShouldBindJSON(&req)` with `binding` struct tags.
- **Context**: Always pass `c.Request.Context()` down to services.
- **User identity**: Extract from middleware context:
  ```go
  userID, _ := middlewares.GetUserID(c)
  email, _ := middlewares.GetUserEmail(c)
  ```
- **Response format**: Use `gin.H{}` for responses.
- **Handlers are thin** — they parse input, call a service, format output.

---

## Service Conventions

- Services receive **repository pointers** and other services via constructor injection.
- All methods accept `context.Context` as the first parameter.
- Services contain **business rules**, **validation beyond HTTP-level**, and **orchestration** between repositories.
- Services should not import `gin` or any HTTP library.

---

## Repository Conventions

- Repositories receive `*sqlx.DB` via constructor injection.
- Use **parameterized queries** — never string-concatenate SQL:
  ```go
  // ✅ Correct
  query := `SELECT * FROM users WHERE id = $1`
  r.db.GetContext(ctx, &user, query, id)

  // ❌ Never
  query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)
  ```
- Use `sqlx` methods: `GetContext`, `SelectContext`, `ExecContext`, `NamedExecContext`.
- Use `db` struct tags on models for column mapping.

---

## Error Handling

The project uses a custom `AppError` type from `pkg/error/`:

```go
import pkgerr "github.com/refynehq/refyne-backend/pkg/error"

// Creating errors
err := pkgerr.NewAppError(
    pkgerr.ErrCodeNotFound,
    http.StatusNotFound,
    "invoice not found",
)

// Domain-specific error definitions
// internal/domains/billing/errors/errors.go
package errors

var (
    ErrInvoiceNotFound = pkgerr.NewAppError(...)
    ErrInvoiceExpired  = pkgerr.NewAppError(...)
)
```

---

## Middleware

Available middleware in `internal/api/middlewares/`:

| Middleware | Purpose |
|---|---|
| `AuthMiddleware()` | JWT token validation; sets `user_id`, `email`, `username` in context |
| `CORSMiddleware()` | Cross-Origin Resource Sharing |
| `RequestIDMiddleware()` | Generates `X-Request-ID` header |
| `SecurityHeadersMiddleware()` | Security response headers |
| `InputValidationMiddleware()` | Request body sanitization |
| `ValidateRequestSize(n)` | Max request body size |
| `NewInMemoryRateLimiter()` | Per-endpoint rate limiting |

### Rate Limiting Example
```go
rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

protected.POST("/checkout",
    rateLimiter.Middleware(middlewares.RateLimitRule{
        Requests: 10,
        Window:   time.Hour,
        KeyFunc: func(c *gin.Context) string {
            userID, _ := c.Get("user_id")
            return "billing:checkout:" + userID.(string)
        },
    }),
    handler.CreateCheckout,
)
```

---

## Database Migrations

Migration files live in `internal/database/migrations/` using sequential numbering:

```
000001_create_users_table.up.sql
000001_create_users_table.down.sql
000002_create_user_settings_table.up.sql
000002_create_user_settings_table.down.sql
```

### Creating a New Migration

1. Create numbered up/down SQL files in `internal/database/migrations/`
2. The `up.sql` creates tables/columns; `down.sql` reverses them
3. Migrations auto-run on startup when `AUTO_MIGRATE=true`

### Conventions
- Use `IF NOT EXISTS` for creates
- Always provide a corresponding `down` migration
- Use `UUID` for primary keys
- Use `timestamptz` for timestamps
- Add `created_at` and `updated_at` to all tables

---

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run a specific package
go test ./internal/domains/auth/...

# Run e2e tests
go test ./tests/...
```

### Test Structure
- Unit tests: co-located in the same package as the code (`_test.go` suffix)
- E2E tests: `tests/e2e_test.go`

---

## Build & Run

```bash
# Install dependencies
go mod download

# Generate Wire DI code
cd cmd && wire

# Build
make build          # or: go build -o bin/app ./cmd/

# Run (development)
make run            # or: go run ./cmd/

# Hot reload (requires Air)
air
```

### Makefile Targets

| Command | Action |
|---|---|
| `make wire` | Regenerate Wire DI code |
| `make build` | Build the binary |
| `make run` | Build and run |
| `make clean` | Remove generated files |
| `make migrate-up` | Apply database migrations |
| `make migrate-down` | Rollback migrations |

### Environment Variables

Required variables in `.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=refyne_dev
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key

# App
APP_ENV=development
PORT=8080
AUTO_MIGRATE=true
FRONTEND_URL=http://localhost:3000
```

---

## Code Style & Linting

- Follow standard Go conventions (`gofmt`, `goimports`)
- Use `.golangci.yml` for linting rules
- Package names: lowercase, single word (e.g., `handler`, `services`, `models`)
- File names: snake_case (e.g., `billing_handler_impl.go`)
- Interface names: PascalCase, no `I` prefix (e.g., `BillingHandler`, not `IBillingHandler`)
- Struct names: PascalCase (e.g., `BillingService`)
- Constructor pattern: `NewXxx` (e.g., `NewBillingService`)

---

## Commit & PR Guidelines

### Commit Messages
Use conventional commits:
```
feat(billing): add invoice creation endpoint
fix(auth): handle expired OTP gracefully
refactor(subscription): extract webhook processing
docs: update CONTRIBUTING.md
chore: regenerate wire_gen.go
```

### Pull Requests
1. Follow the existing patterns documented here
2. Include tests for new features
3. Update `CONTRIBUTING.md` if adding new patterns
4. Run `cd cmd && wire` if Wire configuration changed
5. Ensure `go build ./...` passes
6. No changes to `wire_gen.go` should be committed without running `wire`

---

## Quick Reference Cheat Sheet

### "I want to add a new API endpoint"
→ See [Adding a New Feature to an Existing Domain](#adding-a-new-feature-to-an-existing-domain)

### "I want to add an entirely new feature area"
→ See [Adding a New Domain](#adding-a-new-domain) — follow all 10 steps

### "I want to add a new database table"
→ See [Database Migrations](#database-migrations) — add numbered SQL files

### "My Wire build is failing"
→ Check that:
1. Constructor return types match what consumers expect (interface vs struct)
2. All providers are included in the domain's `wire.go` ProviderSet
3. The domain's ProviderSet is imported in `cmd/wire.go`
4. Run `cd cmd && wire` to see the actual error

### "I need to access the current user in a handler"
→ Use `middlewares.GetUserID(c)` after `AuthMiddleware()` is applied

### "Where do I put shared utilities?"
→ `internal/shared/` for internal shared code; `pkg/` for public reusable packages

### Key Files to Know

| File | Purpose |
|---|---|
| `cmd/wire.go` | Root DI injector — all domains registered here |
| `cmd/wire_gen.go` | Auto-generated — **never edit manually** |
| `internal/api/router.go` | Route assembly — all domain routes registered here |
| `internal/shared/handlerRegistry/registry.go` | Central handler registry — all domain registries wired here |
| `internal/config/config.go` | App configuration struct |
| `internal/database/init.go` | Database initialization |
| `pkg/error/` | Custom error types used throughout |
| `pkg/logging/` | Structured logging setup |
