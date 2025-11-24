# Database Security Documentation

## Overview
This document outlines the comprehensive database security measures implemented in the Refyne backend application. These measures protect against common database vulnerabilities including SQL injection, connection exhaustion, long-running queries, and unauthorized access.

---

## 1. Restricted Database User (Principle of Least Privilege)

### Implementation
**Migration:** `000009_create_restricted_user.up.sql`

A dedicated application user `refyne_app_user` has been created with minimal necessary privileges.

### Permissions Granted
- ✅ **SELECT** - Read data from tables
- ✅ **INSERT** - Add new records
- ✅ **UPDATE** - Modify existing records
- ✅ **DELETE** - Remove records
- ✅ **USAGE, SELECT** on sequences - For auto-increment IDs

### Permissions Explicitly Denied
- ❌ **DROP** - Cannot delete tables/schemas
- ❌ **CREATE** - Cannot create new tables
- ❌ **ALTER** - Cannot modify table structure
- ❌ **TRUNCATE** - Cannot empty tables
- ❌ **SUPERUSER** - No administrative privileges
- ❌ **CREATEDB** - Cannot create databases
- ❌ **CREATEROLE** - Cannot create users/roles
- ❌ **REPLICATION** - Cannot replicate data
- ❌ **BYPASSRLS** - Cannot bypass Row Level Security

### Security Timeouts
```sql
ALTER ROLE refyne_app_user SET statement_timeout = '30s';
ALTER ROLE refyne_app_user SET lock_timeout = '10s';
ALTER ROLE refyne_app_user SET idle_in_transaction_session_timeout = '60s';
```

### Production Setup
1. Run migration 000009 with a superuser account
2. Change the default password in the migration before running
3. Update environment variables:
   ```bash
   DB_USER=refyne_app_user
   DB_PASSWORD=<secure_password>
   ```
4. Restart the application

---

## 2. Connection Pooling & Limits

### Configuration
**File:** `internal/database/init.go`, `internal/database/pool.go`

Connection pooling prevents connection exhaustion and improves performance.

### Production Settings (sqlx.DB)
```go
MaxOpenConns:        20          // Maximum concurrent connections
MaxIdleConns:        10          // Idle connections to keep open
ConnMaxLifetime:     15 minutes  // Connection reuse limit
ConnMaxIdleTime:     5 minutes   // Idle connection timeout
```

### Development Settings
```go
MaxOpenConns:        10          // More conservative
MaxIdleConns:        5
ConnMaxLifetime:     1 hour      // Longer for dev
ConnMaxIdleTime:     10 minutes
```

### Environment Variables
```bash
# Production recommended values
DB_MAX_CONNECTIONS=20
DB_MAX_IDLE_CONNECTIONS=10
DB_CONN_MAX_LIFETIME=15m
DB_CONN_MAX_IDLE_TIME=5m
```

### Benefits
- ✅ Prevents connection pool exhaustion
- ✅ Reduces connection overhead through reuse
- ✅ Automatically closes stale connections
- ✅ Limits concurrent database load

---

## 3. Connection & Query Timeouts

### Connection Timeout
**Default:** 10 seconds

Prevents hanging connection attempts.

```bash
DB_CONNECT_TIMEOUT=10  # seconds
```

### Statement Timeout
**Default:** 30 seconds

Automatically terminates long-running queries to prevent resource exhaustion.

```bash
DB_STATEMENT_TIMEOUT=30000  # milliseconds
```

### Implementation
```go
// Connection string includes timeouts
dsn := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%s statement_timeout=%s",
    host, port, user, password, database, sslMode, connectTimeout, statementTimeout
)
```

### Benefits
- ✅ Prevents database resource exhaustion
- ✅ Protects against denial-of-service attacks
- ✅ Ensures predictable response times
- ✅ Automatic cleanup of hung queries

---

## 4. Secure Query Execution with QueryExecutor

### Overview
**File:** `internal/database/executor.go`

A specialized `QueryExecutor` provides secure, context-aware query execution with automatic timeouts and logging.

### Key Features

#### Context-Based Timeouts
Every query method includes context timeout (default 30s):
```go
ctx, cancel := context.WithTimeout(ctx, qe.timeout)
defer cancel()
```

#### Slow Query Logging
Queries taking >1 second are automatically logged:
```go
if duration > time.Second {
    qe.logger.Warn("Slow query detected", 
        zap.String("query", query),
        zap.Duration("duration", duration))
}
```

#### Available Methods
All methods use context and enforce timeouts:

- `ExecContext()` - INSERT, UPDATE, DELETE operations
- `QueryContext()` - SELECT queries returning multiple rows
- `QueryRowContext()` - SELECT queries returning single row
- `GetContext()` - SELECT with struct scanning
- `SelectContext()` - SELECT with slice scanning
- `BeginTxContext()` - Transaction management
- `PrepareContext()` - Prepared statements
- `NamedExecContext()` - Named parameter execution
- `NamedQueryContext()` - Named parameter queries

### Usage Example
```go
// Inject via Wire
type UserRepository struct {
    executor *database.QueryExecutor
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    var user User
    query := `SELECT * FROM users WHERE id = $1`
    
    // Automatic timeout + slow query logging
    err := r.executor.GetContext(ctx, &user, query, id)
    return &user, err
}
```

### Benefits
- ✅ Automatic query timeout enforcement
- ✅ Slow query detection and logging
- ✅ Consistent error handling
- ✅ Performance monitoring built-in
- ✅ Prevents long-running query attacks

---

## 5. Prepared Statements

### Implementation
The `QueryExecutor.PrepareContext()` method supports prepared statements:

```go
stmt, err := executor.PrepareContext(ctx, query)
```

### Benefits
- ✅ **SQL Injection Prevention** - Parameters are never interpolated into SQL
- ✅ **Performance** - Query plan is cached and reused
- ✅ **Type Safety** - Database validates parameter types

### All Query Methods Use Parameterized Queries
```go
// ✅ Safe - uses parameterized query
executor.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", userName)

// ❌ Unsafe - string concatenation (DON'T DO THIS)
executor.ExecContext(ctx, "INSERT INTO users (name) VALUES ('" + userName + "')")
```

---

## 6. Additional Security Layers

### Input Validation Middleware
**File:** `internal/api/middlewares/input_validation.go`

All incoming requests are validated for:
- XSS attempts (script tags, event handlers, javascript: protocol)
- SQL injection patterns (UNION, SELECT, DROP, etc.)
- Maximum request size (10MB limit)

### SQL Injection Detection
**File:** `internal/shared/validation/validator.go`

Regex patterns detect common SQL injection attempts:
```go
SQLInjectionRegex = `(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|onerror|onload)`
```

### Audit Logging
**Migration:** `000006_create_audit_logs.up.sql`

All database-modifying operations are logged:
- User login/logout
- Password changes
- Data modifications
- Suspicious activity

---

## 7. Security Best Practices

### ✅ DO
- Use the `QueryExecutor` for all database operations
- Always pass context to database methods
- Use parameterized queries ($1, $2, etc.)
- Set appropriate statement timeouts
- Monitor slow query logs
- Use the restricted `refyne_app_user` in production
- Enable SSL mode in production (`sslmode=require`)
- Regularly rotate database passwords
- Use strong passwords (min 16 characters)
- Keep connection pool sizes appropriate for your load

### ❌ DON'T
- Never concatenate user input into SQL strings
- Don't use the superuser account for application connections
- Don't set unlimited timeouts
- Don't skip input validation
- Don't expose database errors to users (sanitize error messages)
- Don't store passwords in plain text in code
- Don't commit database credentials to version control

---

## 8. Monitoring & Alerting

### Slow Query Monitoring
Queries exceeding 1 second are logged:
```
WARN  Slow query detected  query="SELECT * FROM large_table" duration=2.3s
```

### Connection Pool Monitoring
Available via health check endpoints:
```
GET /api/health/detailed
```

Returns:
```json
{
  "database": {
    "status": "healthy",
    "max_open_connections": 20,
    "open_connections": 5,
    "in_use": 2,
    "idle": 3
  }
}
```

### Recommendations
1. **Set up alerts** for:
   - Slow queries (>1s)
   - Connection pool exhaustion (>80% utilization)
   - Query timeouts
   - Failed login attempts (from audit logs)

2. **Regular reviews**:
   - Weekly slow query analysis
   - Monthly audit log review
   - Quarterly security assessment

---

## 9. Environment Variables Reference

### Required
```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=refyne
DB_USER=refyne_app_user          # Use restricted user in production
DB_PASSWORD=<secure_password>
DB_SSL_MODE=require              # Use 'require' in production
```

### Optional (with defaults)
```bash
# Connection pooling
DB_MAX_CONNECTIONS=20            # Max concurrent connections
DB_MAX_IDLE_CONNECTIONS=10       # Idle connections to maintain
DB_CONN_MAX_LIFETIME=15m         # Connection reuse duration
DB_CONN_MAX_IDLE_TIME=5m         # Idle connection timeout

# Timeouts
DB_CONNECT_TIMEOUT=10            # Connection timeout (seconds)
DB_STATEMENT_TIMEOUT=30000       # Query timeout (milliseconds)
```

---

## 10. Testing Database Security

### Test Connection Limits
```bash
# Simulate many concurrent connections
for i in {1..50}; do
  curl http://localhost:8080/api/user/me &
done
```
Expected: Requests should queue gracefully, none should fail

### Test Query Timeout
```sql
-- This should timeout after 30s
SELECT pg_sleep(60);
```
Expected: Query cancelled after 30 seconds

### Test Restricted User
```sql
-- Connect as refyne_app_user and try to:
DROP TABLE users;  -- Should FAIL
CREATE TABLE test (id int);  -- Should FAIL
SELECT * FROM users;  -- Should SUCCEED
INSERT INTO users (...) VALUES (...);  -- Should SUCCEED
```

### Test SQL Injection Prevention
```bash
# Try SQL injection in API
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"' OR '1'='1"}'
```
Expected: Returns 400 Bad Request (blocked by input validation)

---

## 11. Incident Response

### If SQL Injection is Detected
1. Check audit logs for the attack pattern
2. Block the IP address if necessary
3. Review all recent queries for anomalies
4. Verify no data was exfiltrated
5. Update input validation rules if needed

### If Query Timeout Occurs
1. Check the slow query log
2. Analyze the query execution plan
3. Add appropriate indexes if needed
4. Consider increasing timeout if legitimate
5. Optimize the query

### If Connection Pool Exhausted
1. Check active connection count
2. Look for connection leaks (unclosed connections)
3. Review recent code changes
4. Consider increasing pool size temporarily
5. Add monitoring alerts

---

## 12. Compliance & Standards

This implementation follows:
- ✅ **OWASP Top 10** - SQL Injection prevention (#3)
- ✅ **CWE-89** - SQL Injection mitigation
- ✅ **CWE-400** - Uncontrolled Resource Consumption prevention
- ✅ **PCI DSS** - Database security requirements
- ✅ **GDPR** - Data access logging and audit trails
- ✅ **Principle of Least Privilege** - Restricted user permissions

---

## Summary

The database security implementation provides **defense in depth** with multiple layers:

1. **Network Layer** - SSL/TLS encryption
2. **Authentication Layer** - Restricted user with minimal privileges
3. **Connection Layer** - Pooling limits + timeouts
4. **Query Layer** - Context timeouts + prepared statements
5. **Application Layer** - Input validation + XSS prevention
6. **Monitoring Layer** - Slow query logging + audit trails

These measures collectively prevent:
- SQL injection attacks
- Connection exhaustion attacks
- Long-running query DoS
- Unauthorized data access
- Data exfiltration
- Schema manipulation

**All 10 security features are now complete! 🎉**
