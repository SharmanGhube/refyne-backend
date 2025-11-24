# Security Testing Guide - Phase 1.5

## Overview
This guide provides step-by-step testing procedures for all 10 security features implemented in Phase 1.5. Use Postman or any HTTP client to verify each feature.

**Server URL:** `http://localhost:8080`

---

## ✅ PHASE 1: Health Checks & Security Headers

### Test 1.1: Basic Health Check
```
GET http://localhost:8080/api/health
```

**Expected Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-22T21:52:00Z"
}
```

**Verify Security Headers Present:**
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Content-Security-Policy: default-src 'self'`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), microphone=(), camera=()`

### Test 1.2: Detailed Health Check
```
GET http://localhost:8080/api/health/detailed
```

**Expected Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "...",
  "database": {
    "status": "healthy",
    "max_open_connections": 10,
    "open_connections": 2,
    "in_use": 1,
    "idle": 1
  },
  "redis": {
    "status": "healthy",
    "ping": "PONG"
  }
}
```

### Test 1.3: Readiness Check
```
GET http://localhost:8080/api/health/ready
```

**Expected:** 200 OK if all services are ready

### Test 1.4: Liveness Check
```
GET http://localhost:8080/api/health/live
```

**Expected:** 200 OK if server is alive

---

## ✅ PHASE 2: Input Validation & XSS Prevention

### Test 2.1: Valid Registration
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "John",
  "lastname": "Doe",
  "username": "johndoe123",
  "email": "john@example.com",
  "password": "SecurePass123!@#"
}
```

**Expected:** 201 Created with user data

### Test 2.2: XSS Attack - Script Tag
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "<script>alert('XSS')</script>",
  "lastname": "Doe",
  "username": "hacker",
  "email": "hack@example.com",
  "password": "SecurePass123!@#"
}
```

**Expected:** 400 Bad Request - "Malicious content detected"

### Test 2.3: XSS Attack - Event Handler
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "John",
  "lastname": "<img src=x onerror=alert('XSS')>",
  "username": "hacker2",
  "email": "hack2@example.com",
  "password": "SecurePass123!@#"
}
```

**Expected:** 400 Bad Request - "Malicious content detected"

### Test 2.4: SQL Injection Attempt
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "John",
  "lastname": "Doe",
  "username": "admin' OR '1'='1",
  "email": "sql@example.com",
  "password": "SecurePass123!@#"
}
```

**Expected:** 400 Bad Request - "Malicious content detected"

### Test 2.5: Invalid Email Format
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "John",
  "lastname": "Doe",
  "username": "johndoe",
  "email": "invalid-email",
  "password": "SecurePass123!@#"
}
```

**Expected:** 400 Bad Request with validation errors

### Test 2.6: Weak Password
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "John",
  "lastname": "Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "password": "weak"
}
```

**Expected:** 400 Bad Request - Password validation error

### Test 2.7: Invalid Username (Too Short)
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "John",
  "lastname": "Doe",
  "username": "ab",
  "email": "john@example.com",
  "password": "SecurePass123!@#"
}
```

**Expected:** 400 Bad Request - Username must be 3-30 characters

### Test 2.8: Oversized Request (>10MB)
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "[11MB of data]",
  ...
}
```

**Expected:** 413 Request Entity Too Large

---

## ✅ PHASE 3: Rate Limiting

### Test 3.1: Normal Request Rate
Make 3 registration requests within 1 minute:
```
POST http://localhost:8080/api/auth/register
[same payload as Test 2.1 with different emails]
```

**Expected:** All requests succeed (200/201)

### Test 3.2: Rate Limit Exceeded
Make 101 requests rapidly (more than 100 in the window):
```
POST http://localhost:8080/api/auth/register
[repeat 101 times]
```

**Expected:** 
- First 100 requests: Success
- Request 101: **429 Too Many Requests**
```json
{
  "error": "rate limit exceeded",
  "retry_after": 60
}
```

### Test 3.3: Rate Limit Reset
Wait 1 minute after hitting rate limit, then try again:
```
POST http://localhost:8080/api/auth/register
```

**Expected:** Request succeeds (rate limit reset)

---

## ✅ PHASE 4: Account Lockout

### Test 4.1: Failed Login Attempts
Make 5 failed login attempts with wrong password:
```
POST http://localhost:8080/api/auth/request-otp
Content-Type: application/json

{
  "email": "john@example.com"
}
```
Then try to login 5 times with wrong OTP:
```
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "otp": "000000"
}
```

**Expected:**
- First 4 attempts: 401 Unauthorized - "Invalid OTP"
- 5th attempt: 423 Locked - "Account locked due to too many failed attempts"

### Test 4.2: Account Locked Status
Try to login with correct credentials after lockout:
```
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "otp": "[correct OTP]"
}
```

**Expected:** 423 Locked - "Account is locked. Try again in X minutes"

### Test 4.3: Lockout Expiry
Wait 15 minutes (or manually reset in database), then try login:
```sql
-- Manual reset (optional):
UPDATE account_lockouts SET locked_until = NOW() WHERE user_id = '[user_id]';
```

**Expected:** Login succeeds after lockout expires

---

## ✅ PHASE 5: Audit Logging

### Test 5.1: Verify Audit Logs Created
After performing any auth action (register/login/logout), check database:
```sql
SELECT event_type, user_id, ip_address, user_agent, created_at, metadata 
FROM audit_logs 
ORDER BY created_at DESC 
LIMIT 10;
```

**Expected Audit Events:**
- `user_registered`
- `otp_requested`
- `otp_verified`
- `user_login`
- `user_logout`
- `password_changed`
- `failed_login_attempt`
- `account_locked`

### Test 5.2: Verify Metadata Captured
Check that audit logs include:
- IP address
- User agent
- Request ID
- Additional context (email, username, reason, etc.)

---

## ✅ PHASE 6: Error Handling & Request IDs

### Test 6.1: Validation Error
```
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "firstname": "",
  "email": "invalid"
}
```

**Expected Response (400):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "request_id": "req-abc123",
    "details": [
      {
        "field": "firstname",
        "message": "firstname is required"
      },
      {
        "field": "email",
        "message": "invalid email format"
      }
    ]
  }
}
```

**Verify:**
- Response includes `request_id`
- Error structure is consistent
- Field-level details provided

### Test 6.2: Not Found Error
```
GET http://localhost:8080/api/nonexistent
```

**Expected:** 404 Not Found with request_id

### Test 6.3: Internal Server Error
Trigger an error (if possible), verify response includes:
- Sanitized error message (no stack traces)
- Request ID for debugging
- Consistent error structure

---

## ✅ PHASE 7: Token Invalidation on Password Change

### Test 7.1: Login and Get Token
```
POST http://localhost:8080/api/auth/request-otp
POST http://localhost:8080/api/auth/login

Save the access_token and refresh_token
```

### Test 7.2: Access Protected Endpoint
```
GET http://localhost:8080/api/protected/me
Authorization: Bearer [access_token]
```

**Expected:** 200 OK with user data

### Test 7.3: Change Password
```
POST http://localhost:8080/api/auth/reset-password
Content-Type: application/json

{
  "token": "[reset_token]",
  "new_password": "NewSecurePass123!@#"
}
```

### Test 7.4: Old Token Should Fail
```
GET http://localhost:8080/api/protected/me
Authorization: Bearer [old_access_token]
```

**Expected:** 401 Unauthorized - "Token has been invalidated"

### Test 7.5: New Login Required
```
POST http://localhost:8080/api/auth/login
[Get new tokens]

GET http://localhost:8080/api/protected/me
Authorization: Bearer [new_access_token]
```

**Expected:** 200 OK (new tokens work)

---

## ✅ PHASE 8: Device Fingerprinting & Suspicious Login Detection

### Test 8.1: Normal Login
```
POST http://localhost:8080/api/auth/login
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0
```

**Expected:** 200 OK, device session created

### Test 8.2: Check Device Sessions
Query database:
```sql
SELECT device_fingerprint, device_name, device_type, browser, os, 
       ip_address, is_suspicious, created_at 
FROM device_sessions 
WHERE user_id = '[user_id]' 
ORDER BY created_at DESC;
```

**Expected:** Device session record with:
- Device fingerprint (SHA256)
- Parsed device info (browser, OS, type)
- IP address
- `is_suspicious = false` (first login)

### Test 8.3: Login from Same Device
Login again with same User-Agent and IP:
```
POST http://localhost:8080/api/auth/login
User-Agent: [same as Test 8.1]
```

**Expected:** 
- Session updated (`last_used_at` refreshed)
- `is_suspicious = false`

### Test 8.4: Suspicious Login Detection
Login from **different device AND different location**:
```
POST http://localhost:8080/api/auth/login
User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 17_0) Mobile Safari/604.1
X-Forwarded-For: 203.0.113.45 (different IP)
```

**Expected:**
- New device session created
- `is_suspicious = true`
- `suspicion_reason: "New device and new location"`
- User should receive email notification (if implemented)

### Test 8.5: Check Login Locations
```sql
SELECT ip_address, country, city, login_count, 
       first_seen_at, last_seen_at, is_trusted 
FROM login_locations 
WHERE user_id = '[user_id]';
```

**Expected:** Multiple location records tracking login patterns

### Test 8.6: Logout from Specific Device
```
POST http://localhost:8080/api/auth/logout
Authorization: Bearer [token]
```

**Expected:** Current device session marked inactive

### Test 8.7: Logout from All Devices
```
POST http://localhost:8080/api/auth/logout-all
Authorization: Bearer [token]
```

**Expected:** All device sessions for user marked inactive

---

## ✅ PHASE 9: CORS Protection

### Test 9.1: Allowed Origin (from Postman)
```
GET http://localhost:8080/api/health
Origin: http://localhost:3000
```

**Expected:** 
- Response includes `Access-Control-Allow-Origin: http://localhost:3000`
- Request succeeds

### Test 9.2: Disallowed Origin
```
GET http://localhost:8080/api/health
Origin: http://evil-site.com
```

**Expected:** 
- No `Access-Control-Allow-Origin` header
- Browser would block (Postman might still show response)

### Test 9.3: Preflight Request
```
OPTIONS http://localhost:8080/api/auth/register
Origin: http://localhost:3000
Access-Control-Request-Method: POST
Access-Control-Request-Headers: Content-Type
```

**Expected:** 
- 204 No Content
- `Access-Control-Allow-Methods: POST, GET, OPTIONS, PUT, DELETE`
- `Access-Control-Allow-Headers: ...`

---

## ✅ PHASE 10: Database Security

### Test 10.1: Verify Restricted User Permissions
```sql
-- Try to DROP table as refyne_app_user (should FAIL)
\c refyneDB refyne_app_user
DROP TABLE users;
-- Expected: ERROR: permission denied

-- Try to CREATE table (should FAIL)
CREATE TABLE test_table (id INT);
-- Expected: ERROR: permission denied

-- Try to SELECT (should SUCCEED)
SELECT COUNT(*) FROM users;
-- Expected: Success

-- Try to INSERT (should SUCCEED)
INSERT INTO audit_logs (event_type, user_id, ip_address) 
VALUES ('test_event', NULL, '127.0.0.1');
-- Expected: Success
```

### Test 10.2: Verify Connection Pooling
```sql
-- Check active connections
SELECT 
    count(*) as total_connections,
    count(*) FILTER (WHERE state = 'active') as active,
    count(*) FILTER (WHERE state = 'idle') as idle
FROM pg_stat_activity 
WHERE datname = 'refyneDB';
```

**Expected:** 
- Total connections ≤ 10 (dev setting)
- Healthy mix of active/idle

### Test 10.3: Query Timeout Test
Try to execute a long-running query:
```sql
-- This should timeout after 30 seconds
SELECT pg_sleep(60);
```

**Expected:** Query cancelled after 30s (statement_timeout)

### Test 10.4: Connection Timeout
Stop the database temporarily, then try to connect:
```bash
docker stop refyne_db
[Try to start application]
```

**Expected:** Connection attempt times out after 10s

---

## 📊 Testing Summary Checklist

After completing all tests, verify:

- [ ] ✅ **Health Checks** - All 4 endpoints working
- [ ] ✅ **Security Headers** - 10+ headers present on all responses
- [ ] ✅ **Input Validation** - XSS/SQL injection blocked
- [ ] ✅ **Rate Limiting** - 429 after 100 requests
- [ ] ✅ **Account Lockout** - Locked after 5 failures
- [ ] ✅ **Audit Logging** - All events captured in database
- [ ] ✅ **Error Handling** - Consistent responses with request IDs
- [ ] ✅ **Token Invalidation** - Old tokens fail after password change
- [ ] ✅ **Device Fingerprinting** - Sessions tracked, suspicious logins detected
- [ ] ✅ **CORS** - Only allowed origins permitted
- [ ] ✅ **Database Security** - Restricted user, timeouts, pooling working

---

## 🐛 Troubleshooting

### Issue: Rate limit not triggering
- Check Redis is running: `docker ps | grep redis`
- Verify rate limit window in code: `internal/api/middlewares/rate_limit.go`

### Issue: Account not locking
- Check `account_lockouts` table exists
- Verify `MAX_FAILED_ATTEMPTS = 5` in code
- Check system time is correct

### Issue: Audit logs not created
- Verify migration 000006 applied: `SELECT version FROM schema_migrations;`
- Check `audit_logs` table exists
- Verify audit service is initialized

### Issue: Tokens still work after password change
- Check `token_version` field in `users` table
- Verify JWT middleware checks token version
- Clear old tokens from client

### Issue: Device sessions not tracking
- Check migration 000008 applied
- Verify `device_sessions` and `login_locations` tables exist
- Check User-Agent header is present in requests

---

## 📝 Notes

1. **Testing Environment**: All tests assume development environment (`APP_ENV=development`)
2. **Database**: Ensure fresh database or clean test data between test runs
3. **Rate Limits**: May need to wait for rate limit windows to reset
4. **Postman Collections**: Consider creating a Postman collection with all these tests
5. **Automation**: Can be automated with Newman (Postman CLI) or Jest/Mocha

---

**All 10 Security Features Implemented and Ready for Testing! 🎉**
