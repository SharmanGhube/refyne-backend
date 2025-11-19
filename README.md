# Refyne Backend - Community Growth Platform API

A production-ready Go backend for the Refyne platform, featuring AI-powered community management and engagement tools.

## 🚀 Tech Stack

- **Language:** Go 1.23
- **Framework:** Gin HTTP Framework
- **Database:** PostgreSQL with pgx driver
- **Architecture:** Domain-Driven Design (DDD)
- **Dependency Injection:** Google Wire
- **Job Queue:** River (PostgreSQL-based)
- **Authentication:** JWT tokens
- **Logging:** Structured logging with Zap

## 📋 Features

### ✅ Implemented
- **Authentication System**
  - User registration with validation
  - OTP-based passwordless login
  - JWT token generation (access + refresh tokens)
  - Token refresh mechanism
  - Account verification
  - Logout (single device)
  - Logout all devices
  - Token blacklist management
  
- **Security**
  - JWT authentication middleware
  - Password hashing with bcrypt
  - Request ID tracking
  - Token expiration and validation
  - Token blacklist for logout
  
- **Infrastructure**
  - Database migrations
  - Connection pooling
  - Structured error handling
  - Request/response logging
  - Health check endpoint

### 🔜 Coming Soon
- Password reset flow
- Email verification service (SMTP)
- User profile management
- Workspace management
- AI integration (Google Gemini)
- Rate limiting

## 📁 Project Structure

```
refyne-backend/
├── cmd/                        # Application entry point
│   ├── main.go                # Main application
│   ├── wire.go                # Wire injectors config
│   └── wire_gen.go            # Generated DI code
├── internal/
│   ├── api/                   # API layer
│   │   ├── router.go          # Route definitions
│   │   └── middlewares/       # HTTP middlewares
│   │       ├── requestID.go   # Request tracking
│   │       └── auth.go        # JWT authentication
│   ├── bootstrap/             # Application bootstrap
│   ├── config/                # Configuration management
│   ├── database/              # Database layer
│   │   ├── connection.go      # DB connection
│   │   ├── pool.go            # Connection pooling
│   │   └── migrations/        # SQL migrations
│   ├── domains/               # Business domains (DDD)
│   │   ├── auth/              # Authentication domain
│   │   │   ├── handler/       # HTTP handlers
│   │   │   ├── services/      # Business logic
│   │   │   ├── routes/        # Route registration
│   │   │   └── utils/         # Auth utilities (JWT, OTP, password)
│   │   └── user/              # User domain
│   │       ├── core/          # Core user logic
│   │       │   └── repository/ # Data access layer
│   │       ├── models/        # User models
│   │       └── utils/         # User utilities
│   └── shared/                # Shared components
│       ├── handlerRegistry/   # Handler registry
│       └── river/             # Job queue
├── pkg/                       # Public packages
│   ├── error/                 # Error handling
│   └── logging/               # Logging utilities
└── docs/                      # Documentation

```

## 🛠️ Setup & Installation

### Prerequisites
- Go 1.23 or higher
- PostgreSQL 14+
- Git

### Environment Setup

1. Clone the repository:
```bash
git clone https://github.com/refynehq/refyne-backend.git
cd refyne-backend
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Configure your `.env` file:
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=refyne_dev
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# App
APP_ENV=development
PORT=8080
AUTO_MIGRATE=true
```

### Build & Run

1. Install dependencies:
```bash
go mod download
```

2. Generate Wire dependencies:
```bash
cd cmd && wire
```

3. Build the application:
```bash
make build
```

4. Run the application:
```bash
make run
```

Or use the Makefile shortcuts:
- `make wire` - Generate DI code
- `make build` - Build binary
- `make run` - Build and run
- `make clean` - Clean generated files

## 📡 API Endpoints

### Health Check
```
GET /api/health
```

### Authentication (Public)

#### Register User
```
POST /api/auth/register
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

#### Request OTP
```
POST /api/auth/request-otp
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

#### Login (Verify OTP)
```
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "otp": "123456"
}
```

#### Refresh Token
```
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your_refresh_token"
}
```

#### Verify Account
```
POST /api/auth/verify
Content-Type: application/json

{
  "token": "verification_token"
}
```

### Authentication (Protected)

#### Logout
```
POST /api/auth/logout
Authorization: Bearer <access_token>
```

#### Logout All Devices
```
POST /api/auth/logout-all
Authorization: Bearer <access_token>
```

#### Test Protected Route
```
GET /api/protected/me
Authorization: Bearer <access_token>
```

## 🔐 Authentication Flow

### Registration & OTP Login Flow
1. User registers → `POST /api/auth/register`
2. User requests OTP → `POST /api/auth/request-otp` (email + password)
3. User receives OTP (currently in response, will be emailed)
4. User logs in with OTP → `POST /api/auth/login` (email + OTP)
5. Receive access_token and refresh_token

### Using Protected Routes
```bash
# Include JWT token in Authorization header
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/protected/me
```

### Token Refresh
When access token expires (15 minutes):
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "your_refresh_token"}'
```

## 🗄️ Database Migrations

Migrations are automatically applied on startup when `AUTO_MIGRATE=true`.

Manual migration commands:
```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

Current migrations:
- `000001` - Users table
- `000002` - User settings table

## 🔧 Development

### Code Generation
```bash
# Generate Wire dependency injection
cd cmd && wire
```

### Project Conventions
- **Handlers**: HTTP request/response handling
- **Services**: Business logic layer
- **Repositories**: Data access layer
- **Models**: Data structures
- **Utils**: Helper functions
- **Errors**: Custom error types with proper HTTP status codes

### Error Handling
All errors use the custom `AppError` type with:
- Error code
- HTTP status code
- Error type (validation, unauthorized, internal, etc.)
- Severity level
- Context information

## 📝 Logging

Structured logging with request ID tracking:
```go
logger.Info("Message", 
    zap.String("requestID", requestID),
    zap.String("userID", userID))
```

All requests include `X-Request-ID` header for tracking.

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## 📊 Monitoring

- Health check: `GET /api/health`
- Prometheus metrics: (Coming soon)

## 🔒 Security Features

- ✅ JWT-based authentication
- ✅ Password hashing (bcrypt, cost 12)
- ✅ Token blacklist for logout
- ✅ Request ID tracking
- ✅ Input validation
- ✅ SQL injection prevention (parameterized queries)
- ⏳ Rate limiting (coming soon)
- ⏳ CORS configuration (coming soon)

## 📖 Documentation

Additional documentation:
- [Technical Specification](docs/TECHNICAL_SPECIFICATION.md)
- [API Specification](docs/API_SPECIFICATION.md)
- [Development Workflow](docs/DEVELOPMENT_WORKFLOW.md)
- [Product Specification](docs/PRODUCT_SPECIFICATION.md)

## 🤝 Contributing

1. Follow Go best practices and existing code style
2. Write tests for new features
3. Update documentation
4. Use conventional commits

## 📄 License

Proprietary - Refyne HQ

## 📞 Support

For issues or questions, contact: dev@refynehq.com

---

**Version:** 1.0.0  
**Last Updated:** November 18, 2025