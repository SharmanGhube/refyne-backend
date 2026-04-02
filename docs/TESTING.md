# Testing Guide - Refyne Backend

## Overview

This guide covers:
1. Running automated E2E tests
2. Manual testing procedures with curl
3. Verifying database changes
4. Monitoring service health
5. Testing specific domains

---

## Part 1: Automated E2E Tests

### Prerequisites

- PostgreSQL running (Docker: `docker-compose up -d db`)
- Redis running (Docker: `docker-compose up -d redis`)
- Go 1.23 installed
- Environment set: `APP_ENV=test`

### Running All Tests

```bash
# Run all E2E tests with verbose output
cd d:/Refyne/refyne-backend
go test ./tests -v

# Run with coverage report
go test ./tests -v -cover

# Run specific test
go test ./tests -v -run TestUserRegistrationFlow

# Skip integration tests (for faster CI)
go test ./tests -short
```

### Test Suite Structure

**File:** `tests/e2e_test.go`

| Test | Purpose | Duration |
|---|---|---|
| `TestUserRegistrationFlow` | Validate user signup, duplicate prevention | ~2s |
| `TestAuthenticationFlow` | Test login, token generation | ~2s |
| `TestRateLimiting` | Verify rate limiting protection | ~5s |
| `TestUserProfileManagement` | Test profile CRUD operations | ~2s |
| `TestHealthChecks` | Verify health check endpoints | ~1s |
| `TestValidationMiddleware` | Test input validation | ~1s |
| `TestErrorHandling` | Test error responses (404, 405) | ~1s |
| `TestSecurityHeaders` | Verify security headers present | ~1s |

**Total Time:** ~15-20 seconds

### Expected Test Output

```
=== RUN   TestUserRegistrationFlow
    e2e_test.go:XX: Registration should succeed
=== RUN   TestUserRegistrationFlow/Register_new_user_with_valid_email
    --- PASS: TestUserRegistrationFlow (XXms)
    === RUN   TestUserRegistrationFlow/Reject_registration_with_duplicate_email
        --- PASS: TestUserRegistrationFlow/Reject_registration_with_duplicate_email (XXms)

...

=== RUN   TestIntegrationSummary
    e2e_test.go:XX: E2E Integration Tests Summary:
    e2e_test.go:XX: ✓ User registration and duplicate prevention
    e2e_test.go:XX: ✓ Authentication flow (login, tokens)
    e2e_test.go:XX: ✓ Rate limiting protection

PASS
ok  	github.com/refynehq/refyne-backend/tests	15.234s
```

---

## Part 2: Manual Testing with curl

### Prerequisites

Start the application:
```bash
make run
# or
go run cmd/main.go
```

App runs on: `http://localhost:8080`

### 2.1 Health Check

**Verify the app is running:**

```bash
curl http://localhost:8080/api/health
```

**Expected Response:**
```json
{
  "status": "ok",
  "timestamp": "2026-04-02T10:30:00Z"
}
```

### 2.2 User Registration

**Create a new user:**

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "SecurePass123!",
    "name": "Test User"
  }'
```

**Expected Response (201):**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "testuser@example.com",
  "name": "Test User",
  "created_at": "2026-04-02T10:30:00Z"
}
```

**Error Response (409 - Duplicate):**
```json
{
  "error": "User with this email already exists",
  "status_code": 409
}
```

### 2.3 User Login

**Login with email and password:**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "SecurePass123!"
  }'
```

**Expected Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

**Save these tokens for authenticated requests:**
```bash
export ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export REFRESH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 2.4 Get User Profile

**Retrieve authenticated user profile:**

```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response (200):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "testuser@example.com",
  "name": "Test User",
  "username": null,
  "avatar_url": null,
  "created_at": "2026-04-02T10:30:00Z",
  "updated_at": "2026-04-02T10:30:00Z"
}
```

### 2.5 Update User Profile

**Update user information:**

```bash
curl -X PUT http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "username": "testuser123"
  }'
```

**Expected Response (200):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "testuser@example.com",
  "name": "Updated Name",
  "username": "testuser123",
  "updated_at": "2026-04-02T10:31:00Z"
}
```

### 2.6 Refresh Access Token

**Get new access token using refresh token:**

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'$REFRESH_TOKEN'"
  }'
```

**Expected Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

### 2.7 Logout (Blacklist Token)

**Logout and invalidate token:**

```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response (200):**
```json
{
  "message": "Logged out successfully"
}
```

**After logout, using same token should fail (401):**
```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 2.8 Rate Limiting Test

**Send rapid requests to test rate limiting:**

```bash
# Send 110 requests rapidly
for i in {1..110}; do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/health
done | sort | uniq -c
```

**Expected Output:**
```
    100 200     # First 100 requests succeed
     10 429     # After 100, requests get rate-limited (429 Too Many Requests)
```

### 2.9 Test Invalid Requests

**Test JSON validation:**

```bash
# Invalid JSON
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d 'INVALID'
```

**Expected Response (400):**
```json
{
  "error": "Invalid request body",
  "details": "..."
}
```

**Test missing required fields:**

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

**Expected Response (400):**
```json
{
  "error": "Validation error",
  "fields": {
    "password": "Password is required",
    "name": "Name is required"
  }
}
```

### 2.10 Test Error Responses

**Test 404 Not Found:**

```bash
curl -X GET http://localhost:8080/api/nonexistent
```

**Expected Response (404):**
```json
{
  "error": "Not found",
  "path": "/api/nonexistent"
}
```

**Test 405 Method Not Allowed:**

```bash
curl -X PATCH http://localhost:8080/api/health
```

**Expected Response (405):**
```json
{
  "error": "Method not allowed"
}
```

---

## Part 3: Database Verification

### Query User Data

**Connect to PostgreSQL:**

```bash
psql postgres://user:password@localhost:5432/refyne
```

**Check user was created:**

```sql
SELECT id, email, name, created_at FROM users WHERE email = 'testuser@example.com';
```

**expected Output:**
```
                  id                  |       email        |  name     |        created_at
--------------------------------------+--------------------+-----------+------------------------
 123e4567-e89b-12d3-a456-426614174000 | testuser@example.com | Test User | 2026-04-02 10:30:00+00
```

**Check password is hashed (not plaintext):**

```sql
SELECT id, email, password_hash FROM users WHERE email = 'testuser@example.com';
```

**Expected:** Password should start with `$2a$` (bcrypt hash) or similar

**List all users:**

```sql
SELECT COUNT(*) as total_users FROM users;
```

**Check token blacklist (after logout):**

```sql
-- Token blacklist uses Redis, not PostgreSQL
-- See Part 4 below for Redis verification
```

---

## Part 4: Redis Verification

### Check Redis Data

**Connect to Redis:**

```bash
redis-cli -p 6379
# Or with password: redis-cli -p 6379 -a yourpassword
```

**Check token blacklist:**

```redis
KEYS "*blacklist*"
TTL key_name
```

**Check OTP storage:**

```redis
GET "otp:testuser@example.com"
TTL "otp:testuser@example.com"
```

**Check rate limit data:**

```redis
KEYS "*ratelimit*"
GET key_name
```

**Clear all Redis data (for testing):**

```redis
FLUSHDB
```

---

## Part 5: Health Check Endpoints

### Basic Health

```bash
curl http://localhost:8080/api/health
```

### Detailed Health (with service status)

```bash
curl http://localhost:8080/api/health/detailed
```

**Shows:**
- Database connection status
- Redis connection status
- Migration status

### Readiness Probe (Kubernetes)

```bash
curl http://localhost:8080/api/health/ready
```

**Returns 200 if ready, 503 if not**

### Liveness Probe (Kubernetes)

```bash
curl http://localhost:8080/api/health/live
```

**Returns 200 if alive**

---

## Part 6: Testing Specific Domains

### Authentication Domain

```bash
# Test registration flow
curl -X POST http://localhost:8080/api/auth/register ...

# Test login
curl -X POST http://localhost:8080/api/auth/login ...

# Test refresh
curl -X POST http://localhost:8080/api/auth/refresh ...

# Test logout
curl -X POST http://localhost:8080/api/auth/logout ...
```

### User Domain

```bash
# Get profile
curl http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer $TOKEN"

# Update profile
curl -X PUT http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer $TOKEN"

# Get settings
curl http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer $TOKEN"

# Update settings
curl -X PUT http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"language": "en", "timezone": "UTC"}'
```

### Subscription Domain

```bash
# Get checkout URL
curl -X POST http://localhost:8080/api/subscription/checkout \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tier": "starter",
    "currency": "USD"
  }'

# Webhook (local testing with Paddle)
curl -X POST http://localhost:8080/api/webhook/paddle \
  -H "Content-Type: application/json" \
  -H "Paddle-Signature: ..." \
  -d '{
    "data": { ... },
    "event_type": "subscription.created"
  }'
```

---

## Part 7: CI/CD Test Automation

### GitHub Actions

Automated tests run on:
- Every push to `main` branch
- Every pull request

View results at: https://github.com/your-org/refyne-backend/actions

### Local Simulation

```bash
# Run linting first (same as CI)
golangci-lint run ./...

# Run tests with coverage
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Part 8: Troubleshooting Test Failures

### "Connection refused" (PostgreSQL)

```bash
# Check if PostgreSQL is running
docker-compose ps db

# Start database
docker-compose up -d db

# Wait for readiness
docker-compose exec db pg_isready -U postgres
```

### "Connection refused" (Redis)

```bash
# Check if Redis is running
docker-compose ps redis

# Start Redis
docker-compose up -d redis

# Test Redis connection
redis-cli ping  # Should return PONG
```

### Test timeout

```bash
# Run with longer timeout
go test ./tests -v -timeout 30s

# If still failing, check:
# 1. Database migrations are complete
# 2. Both PostgreSQL and Redis are running
# 3. No network issues between containers
```

### Auth token expired

```bash
# Generate new token
curl -X POST http://localhost:8080/api/auth/login ...

# Use the new token in Authorization header
curl ... -H "Authorization: Bearer NEW_TOKEN"
```

---

## Summary

| Layer | Tests | Tools |
|---|---|---|
| Unit | Individual functions | Go testing + testify |
| Integration | Service interactions | E2E tests in tests/ |
| System | Full workflows | curl + manual testing |
| Database | Data verification | psql queries |
| Cache | State verification | redis-cli |

**Quick Test Command:**

```bash
# Run all tests
go test ./... -v -cover

# Fast test (skip integration)
go test ./... -short

# Test one domain
go test ./internal/domains/auth/... -v
```
