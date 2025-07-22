# Refyne Platform - Technical Specification Document

## 1. Executive Summary

**Project Name:** Refyne - Community Growth Platform  
**Version:** 1.0.0  
**Document Version:** 1.0  
**Last Updated:** July 22, 2025  

Refyne is an AI-powered community growth platform that helps creators and businesses protect, understand, engage with, and grow their social media communities. The platform features Otto, an AI assistant that provides context-aware moderation, sentiment analysis, automated engagement, and lead generation capabilities.

## 2. System Architecture Overview

### 2.1 High-Level Architecture
```
Frontend (Next.js) ↔ API Gateway ↔ Backend Microservices (Go) ↔ Database Layer
                                   ↕
                          AI/ML Services (Google Gemini)
                                   ↕
                            External APIs (Instagram)
```

### 2.2 Technology Stack

**Backend:**
- **Language:** Go 1.24.4
- **Framework:** Gin HTTP Framework
- **Architecture:** Microservices with Domain-Driven Design
- **Dependency Injection:** Google Wire
- **Job Queue:** River (PostgreSQL-based)

**Database:**
- **Primary Database:** PostgreSQL (user data, settings, comments, leads)
- **Cache/Sessions:** Redis (caching, session storage, message broker)
- **Queue:** Redis Streams for real-time task processing

**Infrastructure:**
- **Authentication:** JWT (JSON Web Tokens)
- **Security:** HTTPS/TLS, encrypted secrets management
- **Monitoring:** Prometheus metrics
- **Logging:** Structured logging with Zap
- **Migrations:** golang-migrate

**External Integrations:**
- **AI/ML:** Google Gemini API
- **Social Media:** Instagram API
- **Email:** SMTP configuration for notifications

## 3. Domain Architecture

### 3.1 Domain Structure
The application follows Domain-Driven Design with clear separation of concerns:

```
internal/
├── api/                    # API layer (routes, middleware)
├── bootstrap/              # Application bootstrapping
├── config/                 # Configuration management
├── database/              # Database connection and migrations
├── domain/                # Business domains
│   ├── auth/              # Authentication domain
│   ├── user/              # User management domain
│   └── email/             # Email services domain
└── shared/                # Shared utilities and services
```

### 3.2 Core Domains

#### 3.2.1 Authentication Domain (`internal/domain/auth/`)
**Purpose:** Handles user authentication, registration, and session management

**Components:**
- **Handler:** `AuthHandler` - HTTP request handlers
- **Service:** `AuthService` - Business logic for auth operations
- **Routes:** Auth-specific route definitions
- **Models:** Authentication request/response models

**Key Operations:**
- User registration
- User login
- JWT token management
- Password handling and validation

#### 3.2.2 User Domain (`internal/domain/user/`)
**Purpose:** Manages user profiles, settings, and account operations

**Components:**
- **Repository:** `CoreUserRepository` - Database operations
- **Models:** User data models
- **Account Subdomain:** Account-specific operations

**Key Operations:**
- CRUD operations for users
- User profile management
- Account status management
- User settings and preferences

#### 3.2.3 Email Domain (`internal/domain/email/`)
**Purpose:** Handles all email-related functionality

**Components:**
- **Service:** Email sending and management
- **Jobs:** Periodic and scheduled email tasks
- **Models:** Email templates and configurations

## 4. Database Design

### 4.1 Primary Tables

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    
    -- Account Status
    status VARCHAR(20) NOT NULL DEFAULT 'Pending' 
        CHECK (status IN ('Pending', 'Active', 'Banned', 'Deleted')),
    is_active BOOLEAN NOT NULL DEFAULT false,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    
    -- Security
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

**Indexes:**
- `idx_users_email` - Email lookup
- `idx_users_username` - Username lookup
- `idx_users_status` - Status filtering
- `idx_users_active_verified` - Active user queries
- `idx_users_created_at` - Chronological queries

### 4.2 Job Queue Tables (River)
The application uses River for job queue management with PostgreSQL, providing:
- Reliable job processing
- Job scheduling and retry mechanisms
- Worker management
- Job status tracking

## 5. API Specification

### 5.1 Base Configuration
- **Base URL:** `/api/v1`
- **Authentication:** Bearer JWT tokens
- **Content-Type:** `application/json`
- **Request ID:** All requests include unique request IDs for tracking

### 5.2 Authentication Endpoints

#### POST `/api/v1/auth/register`
**Purpose:** Register a new user account

**Request Body:**
```json
{
    "username": "string (required)",
    "password": "string (required)",
    "email": "string (required, email format)"
}
```

**Response (201 Created):**
```json
{
    "message": "User registered successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request format or validation errors
- `409 Conflict` - Username or email already exists

#### POST `/api/v1/auth/login`
**Purpose:** Authenticate user and receive JWT token

**Request Body:**
```json
{
    "email": "string (required, email format)",
    "password": "string (required)"
}
```

**Response (200 OK):**
```json
{
    "message": "User logged in successfully",
    "token": "jwt_token_string",
    "user": {
        "id": "uuid",
        "username": "string",
        "email": "string",
        "is_verified": boolean
    }
}
```

#### POST `/api/v1/auth/refresh`
**Purpose:** Refresh JWT token (placeholder)
**Status:** Not implemented

#### POST `/api/v1/auth/logout`
**Purpose:** Logout user and invalidate token (placeholder)
**Status:** Not implemented

### 5.3 Metrics Endpoint

#### GET `/metrics`
**Purpose:** Prometheus metrics for monitoring
**Authentication:** None required
**Response:** Prometheus format metrics

## 6. Service Layer Architecture

### 6.1 Dependency Injection
The application uses Google Wire for compile-time dependency injection:

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
    
    // API Layer
    api.ProviderSet,
    
    // Application layer
    bootstrap.ProviderSet,
)
```

### 6.2 Error Handling
- Custom error types with `AppError` structure
- HTTP status code mapping
- Structured error logging
- Client-safe error responses

### 6.3 Middleware Stack
- **Request ID:** Unique identifier for request tracing
- **Logging:** Request/response logging
- **Recovery:** Panic recovery and graceful error handling
- **Prometheus:** Metrics collection (optional)

## 7. Security Specifications

### 7.1 Authentication & Authorization
- **JWT Tokens:** Stateless authentication
- **Password Security:** Bcrypt hashing
- **Session Management:** Token-based sessions
- **Request Validation:** Input validation and sanitization

### 7.2 Data Protection
- **HTTPS/TLS:** Enforced for all communication
- **Password Storage:** Never stored in plain text
- **Sensitive Data:** Excluded from JSON responses
- **Database Security:** Connection pooling with secure configurations

### 7.3 Rate Limiting & Protection
- **Input Validation:** Gin binding with validation tags
- **SQL Injection Prevention:** Parameterized queries
- **Request Size Limits:** Configurable request body limits

## 8. Configuration Management

### 8.1 Environment Variables
```go
type Config struct {
    Environment     string          // development/production
    Port            string          // Server port
    AutoMigrate     bool           // Run migrations on startup
    SMTPConfig      SMTPConfig     // Email configuration
    InstagramConfig InstagramConfig // Instagram API config
}
```

### 8.2 Database Configuration
- **Connection Pooling:** pgxpool for PostgreSQL
- **Migration Management:** golang-migrate with versioned SQL files
- **Connection Security:** SSL/TLS for production environments

## 9. Operational Specifications

### 9.1 Logging
- **Structure:** JSON-formatted logs with Zap
- **Levels:** Debug, Info, Warn, Error, Fatal
- **Context:** Request IDs, user context, service context
- **Storage:** File-based logging with rotation

### 9.2 Monitoring & Metrics
- **Metrics:** Prometheus integration
- **Health Checks:** Application health endpoints
- **Performance:** Request duration, database query metrics
- **Alerts:** Error rate and response time monitoring

### 9.3 Deployment
- **Build:** Go binary compilation
- **Migrations:** Automatic database migrations
- **Graceful Shutdown:** Context-based shutdown handling
- **Process Management:** Signal handling for clean shutdowns

## 10. Development Specifications

### 10.1 Code Organization
- **Domain-Driven Design:** Clear domain boundaries
- **Clean Architecture:** Separated concerns and dependencies
- **Interface Segregation:** Small, focused interfaces
- **Dependency Injection:** Compile-time DI with Wire

### 10.2 Testing Strategy
- **Unit Tests:** Individual component testing
- **Integration Tests:** Database and API testing
- **Mock Dependencies:** Testable service interfaces
- **Test Data:** Isolated test database setup

### 10.3 Build & Deployment
- **Makefile:** Build automation
- **Wire Generation:** Automatic dependency injection
- **Binary Output:** Single executable file
- **Configuration:** Environment-based configuration

## 11. Future Roadmap Integration

### 11.1 Planned Microservices
Based on the technical plan, the following services are planned:

1. **ModerationService:** Real-time comment analysis and filtering
2. **AnalyticsService:** Data processing and aggregation
3. **AutomationService:** Visual workflow pipeline execution
4. **InstagramService:** Social media platform integration

### 11.2 AI/ML Integration Points
- **Google Gemini API:** Language understanding and processing
- **Caching Layer:** AI response caching for cost optimization
- **Custom Models:** Enterprise client model fine-tuning

### 11.3 Scalability Considerations
- **Kubernetes:** Container orchestration for scaling
- **Redis Streams:** High-throughput message processing
- **WebSockets:** Real-time dashboard updates
- **Data Warehouse:** Analytics query optimization

## 12. Appendices

### 12.1 Key Dependencies
- `gin-gonic/gin` - HTTP web framework
- `google/wire` - Dependency injection
- `jackc/pgx/v5` - PostgreSQL driver
- `riverqueue/river` - Job queue system
- `golang-migrate/migrate` - Database migrations
- `prometheus/client_golang` - Metrics collection

### 12.2 Development Commands
```bash
# Build application
make build

# Run migrations
make migrate-up

# Generate Wire dependencies
wire gen ./cmd

# Run application
./bin/app
```

### 12.3 Environment Setup
Required environment variables:
- `APP_ENV` - Application environment
- `PORT` - Server port
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - JWT signing secret

---

**Document Prepared By:** Technical Architecture Team  
**Review Status:** Draft  
**Next Review Date:** August 22, 2025
