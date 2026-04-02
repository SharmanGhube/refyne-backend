# Testing Guide - Refyne Backend

This guide covers how to run automated tests and manually verify system functionality.

## Quick Start

### Run E2E Tests
```bash
# Start the backend (if not already running)
make run

# In another terminal, run tests
go test ./tests -v

# Or run specific test
go test ./tests -v -run TestUserRegistrationFlow
```

**Requirements:**
- Backend running at `localhost:8080`
- PostgreSQL connected
- Redis available (for rate limiting tests)

## E2E Tests Coverage

The test suite validates 12 major test categories:

1. **User Registration Flow** - Registration, duplicate prevention, validation
2. **Authentication Flow** - Login, token generation, invalid credentials
3. **Rate Limiting** - Request throttling, 429 responses
4. **User Profile Management** - Get/update profile, auth required
5. **Health Checks** - Basic, detailed, readiness, liveness probes
6. **Request Validation** - Missing fields, invalid types
7. **Error Handling** - 404s, 405s, proper error format
8. **Workspace Management** - Create, list, get, update, delete workspaces
9. **Workspace Members** - List, invite, remove members with role checks
10. **Token Blacklist/Logout** - Token invalidation after logout
11. **Subscription Status** - Check status, create checkout URL
12. **Security Headers** - CORS, Content-Type, security headers

## Running Tests

```bash
# All tests
go test ./tests -v

# Specific test function
go test ./tests -v -run TestWorkspaceManagement

# With verbose logging
go test ./tests -v -v

# Quick mode (skip long-running tests)
go test ./tests -short
```

## Environment Setup

1. Start PostgreSQL and Redis:
```bash
docker-compose up -d
```

2. Start backend:
```bash
make run
```

3. Run tests (in another terminal):
```bash
go test ./tests -v
```

## Test Execution

Each test:
- Verifies the backend is accessible
- Makes HTTP requests to real endpoints
- Validates response status codes
- Checks response format (JSON)
- Verifies data consistency

## Manual Testing

Example commands for manual verification:

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Pass123!","name":"Test"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Pass123!"}'

# Use token
TOKEN="your_access_token_here"
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer $TOKEN"

# Create workspace
curl -X POST http://localhost:8080/api/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Workspace","description":"Test"}'
```

## CI/CD Integration

Tests run automatically in GitHub Actions:
- On every push to main
- On pull requests
- Fails if any test doesn't pass

See `.github/workflows/test.yml` for details.

## Troubleshooting

**Backend not starting**
```bash
make clean && make build && make run
```

**Tests timeout**
- Increase timeout in test code: `client.Timeout = 30 * time.Second`
- Check if services are running

**Rate limit tests fail**
- Ensure Redis is running: `docker ps | grep redis`
- Check RATE_LIMIT_ENABLED=true in .env

**Workspace tests fail**
- Verify migrations ran: `SELECT * FROM workspaces;` in psql
- Check database connection in logs

## Performance

Typical test execution time: 30-60 seconds

Per-test timing (localhost):
- Registration: 50-150ms
- Login: 100-200ms
- Workspace ops: 30-80ms
- Health check: <10ms

## Next Steps

1. Run tests on each code change
2. Monitor test execution time
3. Add tests for new features
4. Check CI/CD pipeline status
