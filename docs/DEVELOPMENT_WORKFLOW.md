# Refyne - Development Workflow & Codebase Guide

## Document Overview
This document provides a practical guide for working with your current Refyne codebase, including development workflows, code organization patterns, and step-by-step implementation instructions.

---

## 1. Current Codebase Analysis

### 1.1 What You Already Have ✅
```
✅ Project structure with Domain-Driven Design
✅ Authentication system foundation (handlers, services, routes)
✅ User management with repository pattern
✅ Database migrations and connection handling
✅ Google Wire dependency injection setup
✅ Logging system with Zap
✅ River job queue system foundation
✅ Basic middleware (request ID, recovery)
✅ Prometheus metrics setup
✅ Error handling system
✅ Configuration management
```

### 1.2 What Needs Implementation 🔨
```
🔨 Complete user registration/login logic
🔨 Instagram OAuth integration
🔨 Google Gemini AI integration
🔨 Comment sync and storage system
🔨 Moderation engine
🔨 Context document management
🔨 Otto chat system
🔨 Response automation
🔨 Analytics and dashboard APIs
🔨 Frontend interface
```

---

## 2. Development Workflow

### 2.1 Daily Development Cycle
```bash
# 1. Start your development day
git pull origin main
make wire              # Regenerate dependency injection
make migrate-up        # Ensure DB is up to date
make dev              # Start development server

# 2. Development work
# - Make your changes
# - Test frequently with make test
# - Check logs in logs/app.log

# 3. End of day
make test             # Run full test suite
git add .
git commit -m "feat: description of changes"
git push origin feature-branch
```

### 2.2 Feature Development Pattern
```bash
# For each new feature:
git checkout -b feature/feature-name
make wire             # If you added new dependencies
make migrate-up       # If you created new migrations
make test            # Ensure tests pass
make dev             # Start development
```

---

## 3. Step-by-Step Implementation Guide

### Week 1: Complete Authentication System

#### Step 1: Complete User Registration Logic
**File:** `internal/domain/auth/services/onboarding.go`

```go
// You need to complete this function
func (s *authService) RegisterUser(c *gin.Context, username, password, email string) *errors.AppError {
    // 1. Add input validation
    if username == "" || password == "" || email == "" {
        return authErrors.NewInvalidEmailError(c, "All fields are required")
    }
    
    // 2. Validate email format
    user := &userModels.User{Email: email}
    if !user.IsValidEmail() {
        return authErrors.NewInvalidEmailError(c, email)
    }
    
    // 3. Validate password policy
    policy := auth.DefaultPasswordPolicy()
    if err := policy.Validate(password); err != nil {
        return authErrors.NewInvalidPasswordError(c, err.Error())
    }
    
    // 4. Check if user exists
    exists, appErr := s.coreUserRepo.UserExistsByEmail(c, email)
    if appErr != nil {
        return appErr
    }
    if exists {
        return authErrors.NewUserAlreadyExistsError(c, "email", email)
    }
    
    // 5. Hash password
    hashedPassword, err := auth.HashPassword(password)
    if err != nil {
        return authErrors.NewPasswordHashingFailedError(c, err)
    }
    
    // 6. Create user
    newUser := &userModels.User{
        ID:           uuid.New().String(),
        Email:        email,
        Username:     username,
        PasswordHash: hashedPassword,
        Status:       "Pending",
        IsActive:     false,
        IsVerified:   false,
    }
    
    // 7. Save to database
    if appErr := s.coreUserRepo.CreateUser(c, newUser); appErr != nil {
        return appErr
    }
    
    return nil
}
```

#### Step 2: Complete Login Logic
**File:** `internal/domain/auth/services/onboarding.go`

```go
func (s *authService) LoginUser(c *gin.Context, email, password string) (string, *errors.AppError) {
    // 1. Get user by email
    user, appErr := s.coreUserRepo.GetUserByEmail(c, email)
    if appErr != nil {
        return "", appErr
    }
    
    // 2. Verify password
    if !auth.CheckPasswordHash(password, user.PasswordHash) {
        return "", authErrors.NewInvalidPasswordError(c, "Invalid credentials")
    }
    
    // 3. Check if user is active
    if !user.IsActive {
        return "", authErrors.NewUserNotActiveError(c, "Account not activated")
    }
    
    // 4. Generate JWT token
    token, err := auth.GenerateJWT(c, user.Username, user.ID, user.Email, 24*60*60) // 24 hours
    if err != nil {
        return "", authErrors.NewJWTGenerationFailedError(c, err)
    }
    
    // 5. Update last login (optional)
    // TODO: Add update last login logic
    
    return token, nil
}
```

#### Step 3: Complete Database Operations
**File:** `internal/domain/user/repository/crud.go`

You need to complete the CRUD operations. Here's an example:

```go
func (r *coreUserRepository) CreateUser(c *gin.Context, user *userModels.User) *errors.AppError {
    r.logger.Info("Creating New User", zap.String("requestID", middlewares.GetRequestID(c)))

    // Set timestamps
    now := time.Now()
    user.CreatedAt = now.Format(time.RFC3339)
    user.UpdatedAt = now.Format(time.RFC3339)

    // Execute the insert query
    _, err := r.db.NamedExec(insertUserQuery, user)
    if err != nil {
        r.logger.Error("Failed to create user", zap.Error(err))
        return userErrors.NewDatabaseError(c, "create user", err)
    }

    r.logger.Info("User created successfully", zap.String("userID", user.ID))
    return nil
}
```

### Week 2: Workspace and Instagram Integration

#### Step 1: Create Workspace Domain
**Create new files:**

1. `internal/domain/workspace/models/workspace.go`
2. `internal/domain/workspace/repository/workspace_repo.go`
3. `internal/domain/workspace/services/workspace_service.go`
4. `internal/domain/workspace/handlers/workspace_handler.go`
5. `internal/domain/workspace/routes/workspace_routes.go`
6. `internal/domain/workspace/wire.go`

#### Step 2: Instagram Integration Domain
**Create new files:**

1. `internal/domain/instagram/models/instagram.go`
2. `internal/domain/instagram/services/oauth_service.go`
3. `internal/domain/instagram/services/sync_service.go`
4. `internal/domain/instagram/handlers/instagram_handler.go`
5. `internal/domain/instagram/routes/instagram_routes.go`
6. `internal/domain/instagram/wire.go`

#### Step 3: Update Wire Configuration
**File:** `cmd/wire.go`

```go
var AppSet = wire.NewSet(
    // Core infrastructure
    config.ProviderSet,
    database.ProviderSet,
    logging.ProviderSet,

    // Shared Services
    riverqueue.ProviderSet,
    registry.ProviderSet,

    // Domain layer
    user.ProviderSet,
    auth.ProviderSet,
    workspace.ProviderSet,    // Add this
    instagram.ProviderSet,    // Add this

    // API Layer
    api.ProviderSet,

    // Application layer
    bootstrap.ProviderSet,
)
```

### Week 3-4: AI Integration

#### Step 1: Create AI Domain
**Create new files:**

1. `internal/domain/ai/models/analysis.go`
2. `internal/domain/ai/services/gemini_service.go`
3. `internal/domain/ai/services/moderation_service.go`
4. `internal/domain/ai/services/chat_service.go`
5. `internal/domain/ai/wire.go`

#### Step 2: Gemini Service Implementation
**File:** `internal/domain/ai/services/gemini_service.go`

```go
package ai

import (
    "context"
    "encoding/json"
    
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type GeminiService interface {
    AnalyzeComment(ctx context.Context, comment string, context []string) (*CommentAnalysis, error)
    GenerateResponse(ctx context.Context, comment string, templates []string) (*ResponseSuggestion, error)
    ChatWithOtto(ctx context.Context, query string, userData map[string]interface{}) (*ChatResponse, error)
}

type geminiService struct {
    client *genai.Client
    logger *zap.Logger
}

func NewGeminiService(apiKey string) (GeminiService, error) {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        return nil, err
    }
    
    return &geminiService{
        client: client,
        logger: logging.GetServiceLogger("GeminiService"),
    }, nil
}
```

---

## 4. Database Migration Strategy

### 4.1 Create Migrations in Order

```bash
# Week 1: Complete user system
migrate create -ext sql -dir internal/database/migrations -seq update_users_table_subscription

# Week 2: Workspace and social accounts
migrate create -ext sql -dir internal/database/migrations -seq create_workspaces_table
migrate create -ext sql -dir internal/database/migrations -seq create_social_accounts_table
migrate create -ext sql -dir internal/database/migrations -seq create_media_table
migrate create -ext sql -dir internal/database/migrations -seq create_comments_table

# Week 3: AI and context
migrate create -ext sql -dir internal/database/migrations -seq create_context_documents_table
migrate create -ext sql -dir internal/database/migrations -seq create_moderation_rules_table

# Week 4: Automation
migrate create -ext sql -dir internal/database/migrations -seq create_response_templates_table
migrate create -ext sql -dir internal/database/migrations -seq create_automated_responses_table
```

### 4.2 Migration Example
**File:** `internal/database/migrations/002_create_workspaces_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Settings
    timezone VARCHAR(50) DEFAULT 'UTC',
    default_moderation_strictness VARCHAR(10) DEFAULT 'medium',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_workspaces_owner ON workspaces(owner_user_id);
```

---

## 5. Testing Strategy for Current Codebase

### 5.1 Unit Tests Structure
```
tests/
├── unit/
│   ├── auth/
│   │   ├── services_test.go
│   │   └── handlers_test.go
│   ├── user/
│   │   └── repository_test.go
│   └── ai/
│       └── gemini_test.go
├── integration/
│   ├── auth_flow_test.go
│   └── instagram_sync_test.go
└── e2e/
    └── user_journey_test.go
```

### 5.2 Example Test
**File:** `tests/unit/auth/services_test.go`

```go
package auth_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/gin-gonic/gin"
    
    authServices "github.com/refynehq/refyne-backend/internal/domain/auth/services"
)

func TestAuthService_RegisterUser(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    service := authServices.NewAuthService(mockRepo)
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(nil)
    
    // Test cases
    tests := []struct {
        name     string
        username string
        password string
        email    string
        wantErr  bool
    }{
        {
            name:     "Valid registration",
            username: "testuser",
            password: "Password123!",
            email:    "test@example.com",
            wantErr:  false,
        },
        {
            name:     "Invalid email",
            username: "testuser",
            password: "Password123!",
            email:    "invalid-email",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.RegisterUser(c, tt.username, tt.password, tt.email)
            if tt.wantErr {
                assert.NotNil(t, err)
            } else {
                assert.Nil(t, err)
            }
        })
    }
}
```

---

## 6. Debugging and Development Tips

### 6.1 Debugging with Your Current Setup

```bash
# View logs in real-time
tail -f logs/app.log

# Check database connections
make db-status

# Test specific endpoints
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"Password123!","email":"test@example.com"}'

# Monitor metrics
curl http://localhost:8080/metrics
```

### 6.2 Common Issues and Solutions

| Issue | Solution |
|-------|----------|
| Wire generation fails | Run `make wire` after adding new dependencies |
| Database connection errors | Check DB_* environment variables |
| Migration fails | Check migration syntax and dependencies |
| JWT token issues | Verify JWT_SECRET is set |
| API 404 errors | Check route registration in router.go |

---

## 7. Code Organization Best Practices

### 7.1 Domain Structure Pattern
```
internal/domain/[domain]/
├── models/          # Data structures
├── repository/      # Data access layer
│   ├── interfaces.go
│   ├── implementation.go
│   └── errors/
├── services/        # Business logic
│   ├── interfaces.go
│   ├── implementation.go
│   └── errors/
├── handlers/        # HTTP handlers
│   ├── interfaces.go
│   └── implementation.go
├── routes/          # Route definitions
└── wire.go         # Dependency injection
```

### 7.2 Error Handling Pattern
```go
// Always use your custom error system
func (s *service) DoSomething(c *gin.Context, input string) *errors.AppError {
    if input == "" {
        return customErrors.NewValidationError(c, "input cannot be empty")
    }
    
    result, err := s.repo.GetData(input)
    if err != nil {
        return customErrors.NewDatabaseError(c, "get data", err)
    }
    
    return nil
}
```

---

## 8. Environment Configuration

### 8.1 Required Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=refyne_user
DB_PASSWORD=your_password
DB_NAME=refyne_db
DB_SSL_MODE=disable

# Application
APP_ENV=development
PORT=8080
JWT_SECRET=your_jwt_secret_here
AUTO_MIGRATE=true

# External APIs
GOOGLE_GEMINI_API_KEY=your_gemini_key
INSTAGRAM_CLIENT_ID=your_instagram_client_id
INSTAGRAM_CLIENT_SECRET=your_instagram_client_secret

# Email (for notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email
SMTP_PASSWORD=your_app_password
```

### 8.2 Development Setup Script
```bash
#!/bin/bash
# scripts/dev-setup.sh

echo "Setting up Refyne development environment..."

# 1. Install dependencies
go mod tidy

# 2. Generate Wire dependencies
go generate ./...

# 3. Set up database
createdb refyne_db
psql refyne_db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"

# 4. Run migrations
migrate -path internal/database/migrations -database "postgresql://localhost/refyne_db?sslmode=disable" up

# 5. Start development server
air # or make dev
```

---

## 9. Next Steps Checklist

### Immediate (This Week)
- [ ] Complete user registration and login logic
- [ ] Fix any failing tests
- [ ] Add password hashing utilities
- [ ] Complete user repository CRUD operations
- [ ] Add basic input validation

### Week 2
- [ ] Create workspace domain structure
- [ ] Add workspace CRUD operations
- [ ] Implement Instagram OAuth flow
- [ ] Create Instagram API client
- [ ] Add basic post/comment sync

### Week 3
- [ ] Integrate Google Gemini API
- [ ] Implement comment analysis
- [ ] Add moderation engine
- [ ] Create context document system

### Week 4
- [ ] Build Otto chat interface
- [ ] Add response automation
- [ ] Create dashboard APIs
- [ ] Add analytics calculations

---

This workflow guide gives you a clear path forward with your existing codebase. Start with completing the authentication system, then gradually add new domains following the established patterns in your project.

**Key Success Factors:**
1. Follow your existing architectural patterns
2. Complete one domain at a time
3. Test frequently and early
4. Use your existing error handling and logging systems
5. Keep migrations atomic and reversible
