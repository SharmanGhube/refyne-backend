# Refyne Backend - Complete Infrastructure & Deployment Guide

**Last Updated:** April 17, 2026  
**Environment:** Go 1.23 + Railway (Production) + Docker Compose (Local)  
**Status:** ✅ Production-Ready

---

## 📋 Quick Navigation

- [Local Development Setup](#local-development-setup)
- [Database Architecture](#database-architecture)
- [Redis Setup](#redis-setup)
- [Railway Deployment](#railway-deployment)
- [Paddle Integration](#paddle-integration)
- [Frontend Integration](#frontend-integration)
- [Monitoring & Debugging](#monitoring--debugging)
- [Troubleshooting](#troubleshooting)

---

## Local Development Setup

### Prerequisites

- **Docker Desktop** (https://www.docker.com/products/docker-desktop)
- **Go 1.23+** (optional, for local building)
- **PostgreSQL Client** (optional, `psql` for CLI access)
- **Redis CLI** (optional, for CLI access)

### Step 1: Start Local Infrastructure

All services run in Docker Compose:

```bash
# Navigate to project root
cd /path/to/refyne-backend

# Start all services (PostgreSQL, PgAdmin, Redis, RedisInsight)
docker-compose up -d

# Verify services are running
docker-compose ps
```

**Expected Output:**
```
NAME                    STATUS          PORTS
refyne_db               Up 2 minutes    0.0.0.0:5432->5432/tcp
pgadmin4_container      Up 2 minutes    0.0.0.0:5050->80/tcp
refyne_redis           Up 2 minutes    0.0.0.0:6379->6379/tcp
redisinsight_container Up 2 minutes    0.0.0.0:5540->5540/tcp
```

### Step 2: Access Local Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **PgAdmin** | http://localhost:5050 | Email: `sharmanghube@gmail.com`, Password: `Goobs@123` |
| **RedisInsight** | http://localhost:5540 | No login required |
| **PostgreSQL** | localhost:5432 | User: `root`, Password: `Goobs@123`, DB: `refyneDB` |
| **Redis** | localhost:6379 | Password: `crashed` |

### Step 3: Configure Environment

Copy environment variables to your `.env` file:

```bash
# Copy from .env template and update values
cp .env.template .env

# Key local values already set:
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=root
export DB_PASSWORD=Goobs@123
export DB_NAME=refyneDB

export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=crashed
export REDIS_DB=0

export APP_ENV=development
export APP_PORT=8080
export FRONTEND_URL=http://localhost:3000
```

### Step 4: Start Backend Server

```bash
# Install dependencies
go mod download

# Generate Wire DI code
cd cmd
wire
cd ..

# Run migrations and start server
APP_ENV=development make run

# Or manually:
go run ./cmd/main.go
```

**Expected Output:**
```
Database connection pool initialized successfully
Redis connection established successfully
Server started successfully on :8080
```

### Useful Commands

```bash
# View logs from specific service
docker-compose logs -f refyne_db
docker-compose logs -f refyne_redis

# Access PostgreSQL CLI
docker-compose exec db psql -U root -d refyneDB

# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ deletes data)
docker-compose down -v

# Restart specific service
docker-compose restart refyne_db
```

---

## Database Architecture

### PostgreSQL Overview

**Location:**
- **Local:** Container: `refyne_db`, Port: `5432`
- **Railroad:** Railway PostgreSQL plugin (auto-provisioned)
- **Storage:** Local volume: `./DBdata/`

**Connection Pool (Production):**
```
Max Connections:    20
Idle Connections:   10
Connection Lifetime: 15 minutes
Idle Timeout:       5 minutes
SSL Mode:           require (Railway), disable (Local)
```

### Database Schema

The database is auto-migrated on startup (when `AUTO_MIGRATE=true`).

**Current Tables (20+ migrations):**

```
users
├─ id (UUID primary key)
├─ email (unique)
├─ password_hash
├─ full_name
├─ username (optional)
├─ onboarding_completed
├─ subscription_status (free/active/paused/canceled)
├─ subscription_tier (pro, or null)
├─ paddle_customer_id
├─ paddle_subscription_id
├─ subscription_started_at
├─ subscription_ends_at
├─ created_at
├─ updated_at
└─ deleted_at (soft delete)

user_settings
├─ id (UUID primary key)
├─ user_id (foreign key → users)
├─ language (default: en)
├─ timezone (default: UTC)
├─ email_notifications (boolean)
└─ theme (light/dark)

verification_tokens
├─ id (UUID)
├─ user_id (foreign key)
├─ token_hash
├─ expires_at

password_reset_tokens
├─ id (UUID)
├─ user_id (foreign key)
├─ token_hash
├─ expires_at

account_security
├─ id (UUID)
├─ user_id (foreign key)
├─ failed_attempts (integer)
├─ locked_until (timestamp)
├─ last_login (timestamp)

audit_logs
├─ id (UUID)
├─ user_id (foreign key)
├─ action (login/logout/register/password_reset)
├─ ip_address
├─ user_agent
├─ created_at

device_sessions
├─ id (UUID)
├─ user_id (foreign key)
├─ refresh_token_hash
├─ device_name
├─ last_activity
├─ created_at
├─ expires_at

workspaces
├─ id (UUID)
├─ user_id (foreign key → users)
├─ name
├─ created_at
├─ updated_at

workspace_members
├─ id (UUID)
├─ workspace_id (foreign key → workspaces)
├─ user_id (foreign key → users)
├─ role (owner/member)
├─ joined_at

otto_conversations
├─ id (UUID)
├─ user_id (foreign key)
├─ title
├─ created_at
├─ updated_at

otto_messages
├─ id (UUID)
├─ conversation_id (foreign key)
├─ role (user/assistant)
├─ content
├─ created_at

instagram_accounts
├─ id (UUID)
├─ user_id (foreign key)
├─ instagram_user_id
├─ access_token
├─ connected_at

... (and more for media, analytics, etc.)
```

### Viewing Database Structure

**Via PgAdmin (GUI):**
1. Go to http://localhost:5050
2. Login with email/password
3. Add Server → Connect to `refyne_db:5432`
4. Browse tables → Right-click → View Data

**Via SQL:**
```sql
-- Connect to database
psql -U root -d refyneDB -h localhost

-- List all tables
\dt

-- View table structure
\d users

-- View all users
SELECT * FROM users;

-- Count records in each table
SELECT table_name, COUNT(*) FROM information_schema.tables GROUP BY table_name;
```

### Migrations

**Location:** `internal/database/migrations/sql/`

Migrations run automatically on startup. To add a new migration:

```bash
# Migration files have format: NNNNNN_description.{up,down}.sql

# Example structure:
# 000001_create_users_table.up.sql (runs first)
# 000001_create_users_table.down.sql (runs on rollback)
```

---

## Redis Setup

### Redis Overview

**Location:**
- **Local:** Container: `refyne_redis`, Port: `6379`
- **Railway:** Railway Redis plugin (auto-provisioned)
- **Storage:** Local volume: `./redis_data/` (with AOF persistence)

**Configuration:**
```
Port:      6379
Password:  crashed (local), random (Railway)
Database:  0
Mode:      Standalone
Persistence: Append-Only File (AOF)
```

### Redis Use Cases in Refyne

1. **Rate Limiting** - in-memory request tracking (100 req/min per IP)
2. **Session Management** - token blacklist (logout functionality)
3. **Caching** - future implementation
4. **Job Queue Coordination** - River queue coordination

### Accessing Redis

**Via RedisInsight (GUI):**
1. Go to http://localhost:5540
2. Click "Add Redis Database"
3. Host: `localhost`, Port: `6379`, Password: `crashed`

**Via Redis CLI:**
```bash
# Connect to local Redis
redis-cli -h localhost -p 6379 -a crashed

# Common commands
PING                          # Check connection
KEYS *                        # List all keys
GET key_name                  # Get value
SET key_name value           # Set value
DEL key_name                 # Delete key
FLUSHDB                      # Clear all data (⚠️ destructive)

# Check rate limiting data
KEYS *rate_limit*
GET rate_limit:ip:192.168.1.1
```

### Redis Persistence

Data is saved to disk automatically using AOF (Append-Only File):

```
Local: ./redis_data/appendonly.aof (synced every second)
Railway: Managed by Railway (automatic backups)
```

---

## Railway Deployment

### Prerequisites

- **Railway Account** (https://railway.app)
- **GitHub Account** with refyne-backend pushed
- **Domain/DNS** (optional, Railway provides default domain)

### Architecture Diagram

```
┌─────────────────────────────────────────┐
│           Railway Project                │
├─────────────────────────────────────────┤
│                                         │
│  ┌──────────────────────────────┐      │
│  │  Refyne Backend Service      │      │
│  │  (Go Application)            │      │
│  │  - Port: 8080                │      │
│  │  - Auto-deploy on git push   │      │
│  └──────────────────────────────┘      │
│                │                        │
│    ┌───────────┼───────────┐           │
│    │           │           │           │
│    ▼           ▼           ▼           │
│  ┌─────┐   ┌─────────┐  ┌──────┐     │
│  │ DB  │   │ Redis   │  │ Email│     │
│  │ Pg  │   │ Cache   │  │ SMTP │     │
│  └─────┘   └─────────┘  └──────┘     │
│                                         │
└─────────────────────────────────────────┘
         ↕ HTTP/HTTPS ↕
┌─────────────────────────────────────────┐
│      Frontend Application               │
│  (Next.js/React on Vercel/Netlify)     │
│  - http://refyne.com (your domain)     │
└─────────────────────────────────────────┘
```

### Step-by-Step Deployment

#### Step 1: Create Railway Project

1. Go to https://railway.app
2. Click "New Project"
3. Select "Deploy from GitHub"
4. Authorize OAuth → Select `refyne-backend` repository
5. Railway detects `railway.json` and auto-configures

#### Step 2: Add PostgreSQL Service

1. In Railway dashboard, click "New" (top-right)
2. Select "Database" → "PostgreSQL"
3. Railway creates database with auto-generated credentials
4. Link to Refyne service:
   - Click PostgreSQL card → "Connect"
   - Select Refyne service
   - Variables now available: `${{Postgres.PGHOST}}`, `${{Postgres.PGUSER}}`, etc.

#### Step 3: Add Redis Service

1. Click "New" (top-right)
2. Select "Database" → "Redis"
3. Railway creates Redis instance
4. Link to Refyne service:
   - Click Redis card → "Connect"
   - Select Refyne service
   - Variables now available: `${{Redis.REDIS_HOST}}`, `${{Redis.REDIS_PORT}}`, etc.

#### Step 4: Configure Environment Variables

1. Click Refyne service → "Variables" tab
2. Add all required variables from `railway.env.template`:

**Database Variables (auto-populated):**
```env
DB_HOST=${{Postgres.PGHOST}}
DB_PORT=${{Postgres.PGPORT}}
DB_USER=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
DB_NAME=${{Postgres.PGDATABASE}}
DB_SSL_MODE=require
```

**Redis Variables (auto-populated):**
```env
REDIS_HOST=${{Redis.REDIS_HOST}}
REDIS_PORT=${{Redis.REDIS_PORT}}
REDIS_PASSWORD=${{Redis.REDIS_PASSWORD}}
REDIS_DB=0
```

**Application Configuration (manual):**
```env
APP_ENV=production
APP_PORT=8080
AUTO_MIGRATE=true

JWT_SECRET=YOUR_GENERATED_SECRET_HERE
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

FRONTEND_URL=https://your-frontend-domain.com
```

**Email Configuration (manual):**
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@refyne.com
```

**Paddle Configuration (manual):**
```env
PAYMENT_MODE=sandbox
PADDLE_SANDBOX_API_KEY=pdl_sandbox_apikey_...
PADDLE_SANDBOX_WEBHOOK_SECRET=ntfset_...
PADDLE_SANDBOX_PRODUCT_ID_PRO=pri_...
```

#### Step 5: Push to Deploy

```bash
# Push to main branch
git push origin main

# Railway automatically:
# 1. Detects changes
# 2. Builds Docker image
# 3. Runs migrations
# 4. Deploys to production
# 5. Shows deployment status
```

#### Step 6: Verify Deployment

```bash
# Check health
curl https://your-refyne-service.railway.app/api/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2026-04-17T10:30:00Z"
}
```

### Railway Dashboard Features

**Deployments Tab:**
- View deployment history
- Rollback to previous version
- View logs in real-time
- Cancel active deployment

**Variables Tab:**
- Edit environment variables
- Auto-sync with linked services
- Save and redeploy automatically

**Settings Tab:**
- Custom domain setup
- Health check configuration
- Resource limits
- Auto-scaling rules

**Metrics Tab:**
- CPU usage
- Memory usage
- Request count
- Error rate

### Railway Environment Variables Reference

| Variable | Source | Auto-Populated |
|----------|--------|---|
| `${{Postgres.PGHOST}}` | PostgreSQL Service | ✅ After linking |
| `${{Postgres.PGPORT}}` | PostgreSQL Service | ✅ After linking |
| `${{Postgres.PGUSER}}` | PostgreSQL Service | ✅ After linking |
| `${{Postgres.PGPASSWORD}}` | PostgreSQL Service | ✅ After linking |
| `${{Postgres.PGDATABASE}}` | PostgreSQL Service | ✅ After linking |
| `${{Redis.REDIS_HOST}}` | Redis Service | ✅ After linking |
| `${{Redis.REDIS_PORT}}` | Redis Service | ✅ After linking |
| `${{Redis.REDIS_PASSWORD}}` | Redis Service | ✅ After linking |

**Manual Variables (must be entered):**
- `JWT_SECRET` (generate with `openssl rand -base64 64`)
- `PADDLE_SANDBOX_API_KEY`
- `PADDLE_SANDBOX_WEBHOOK_SECRET`
- `SMTP_PASSWORD` (Gmail app password)
- All API keys

---

## Paddle Integration

### Overview

Refyne uses **Paddle** for subscription payments (payment processor similar to Stripe).

**Current Setup:**
- **Payment Mode:** Sandbox (testing) + Production (live)
- **Subscription Tier:** Single "Pro" tier only ($0 - customer decides)
- **Billing Cycle:** Monthly
- **Customer Management:** Via Paddle customer portal

### Paddle Configuration

**File:** `internal/domains/subscription/config/paddle_config.go`

Refyne supports three payment modes:

| Mode | Use Case | API Keys | Environment |
|------|----------|----------|---|
| `mock` | Unit tests | None | development |
| `sandbox` | Testing real API | Sandbox keys | staging |
| `production` | Live payments | Live keys | production |

### Step 1: Create Paddle Account

1. Go to https://www.paddle.com
2. Create account (not sandbox account yet)
3. Complete business verification
4. Save login credentials

### Step 2: Create Sandbox Account

1. Go to https://sandbox-vendors.paddle.com
2. Create new sandbox account
3. Verify email
4. Save sandbox credentials separately

### Step 3: Obtain Sandbox API Key

**In Paddle Sandbox Dashboard:**

1. Click "Developers" → "API credentials"
2. Click "Generate key"
3. Copy API key (starts with `pdl_`)

**Example:**
```
pdl_sandbox_apikey_01kmg59k2jfaj5aqfev0m590dr_CsjSN419gAPSNqNjCVjD7J_AGa
```

### Step 4: Create Webhook Secret

**In Paddle Sandbox Dashboard:**

1. Click "Developers" → "Webhooks"
2. Click "Create webhook"
3. **Destination URL:** `https://your-railway-domain.railway.app/api/webhook/paddle`
4. **Events to subscribe:**
   - `subscription.created`
   - `subscription.updated`
   - `subscription.canceled`
   - `transaction.completed`
   - `transaction.updated`
5. Copy webhook signing secret

**Example:**
```
ntfset_01kmg5p43qm4py8d9rqganehx9
```

### Step 5: Create Product Price

**For Pro Tier:**

1. Go to "Products" in Sandbox Dashboard
2. Create product: "Refyne Subscriptions"
3. Create price:
   - **Name:** "Refyne Pro"
   - **Billing:** Monthly
   - **Price:** $0 (or test amount, e.g., $99/month)
   - **Copy Pricer ID:** e.g., `pri_01kb65b3gzy2xn21nh0zw922yn`

### Step 6: Configure Environment

**Local Development (`.env`):**
```env
PAYMENT_MODE=sandbox
PADDLE_SANDBOX_API_KEY=pdl_sandbox_apikey_01kmg59k2jfaj5aqfev0m590dr_CsjSN419gAPSNqNjCVjD7J_AGa
PADDLE_SANDBOX_WEBHOOK_SECRET=ntfset_01kmg5p43qm4py8d9rqganehx9
PADDLE_SANDBOX_PRODUCT_ID_PRO=pri_01kb65b3gzy2xn21nh0zw922yn
```

**Railway (via Variables tab):**
```
PAYMENT_MODE=sandbox
PADDLE_SANDBOX_API_KEY=pdl_sandbox_apikey_...
PADDLE_SANDBOX_WEBHOOK_SECRET=ntfset_...
PADDLE_SANDBOX_PRODUCT_ID_PRO=pri_...
```

### Step 7: Test Paddle Integration

**Local Testing:**

```bash
# Start backend
make run

# Check logs for:
# "Running in sandbox payment mode"
# "Paddle configuration validated successfully"

# Test subscription creation
curl -X POST http://localhost:8080/api/subscription/checkout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tier":"pro"}'

# Returns:
{
  "status": "success",
  "data": {
    "checkout_url": "https://sandbox-checkout.paddle.com/...",
    "transaction_id": "txn_...",
    "expires_in": 3600
  }
}
```

**Test Card Numbers (Paddle Sandbox):**

| Card | Number | Exp | CVC |
|------|--------|-----|-----|
| Visa (Success) | 4111 1111 1111 1111 | 12/25 | 123 |
| Visa (Decline) | 4000 0000 0000 0002 | 12/25 | 123 |
| Mastercard | 5555 5555 5555 4444 | 12/25 | 123 |

### Webhook Flow

```
1. User completes payment on Paddle checkout
2. Paddle sends transaction.paid webhook to backend
3. Backend receives at POST /api/webhook/paddle
4. Verifies webhook signature (HMAC-SHA256)
5. Identifies subscription event
6. Updates user in database:
   - subscription_status = "active"
   - subscription_tier = "pro"
   - paddle_customer_id = "ctm_..."
   - paddle_subscription_id = "sub_..."
7. Returns 200 OK to Paddle
8. Frontend polls GET /api/subscription/status
9. Once status is "active", redirects to dashboard
```

### Switching to Production

When ready for live payments:

1. Create production Paddle account (https://vendors.paddle.com)
2. Obtain production credentials:
   - `PADDLE_LIVE_API_KEY`
   - `PADDLE_LIVE_WEBHOOK_SECRET`
   - `PADDLE_LIVE_PRODUCT_ID_PRO`

3. Update environment variables:
   ```env
   PAYMENT_MODE=production
   PADDLE_LIVE_API_KEY=pdl_live_apikey_...
   PADDLE_LIVE_WEBHOOK_SECRET=ntfset_...
   PADDLE_LIVE_PRODUCT_ID_PRO=pri_...
   ```

4. Update webhook URL to production domain
5. Deploy to Railway
6. Verify logs show: `Running in production payment mode`

---

## Frontend Integration

### Base URLs

**Development (Local):**
```
Backend: http://localhost:8080
Frontend: http://localhost:3000
```

**Production (Railway):**
```
Backend: https://your-refyne-service.railway.app
Frontend: https://your-domain.com
```

### CORS Configuration

**Allowed Origins (.env):**
```env
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,https://your-domain.com
```

The backend sends these headers on all responses:
```
Access-Control-Allow-Origin: <matched-origin>
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
```

### Critical API Endpoints for Frontend

#### Authentication Flow

```typescript
// 1. Register
POST /api/auth/register
Body: { email, password, full_name }

// 2. Verify Email
POST /api/auth/verify
Body: { token }  // from email link

// 3. Request OTP
POST /api/auth/request-otp
Body: { email }

// 4. Login
POST /api/auth/login
Body: { email, otp }
Response: { access_token, refresh_token, user }

// 5. Refresh Token
POST /api/auth/refresh
Body: { refresh_token }
Response: { access_token }

// 6. Logout
POST /api/auth/logout
Headers: { Authorization: Bearer access_token }

// 7. Get Protected Endpoint
GET /api/protected/me
Headers: { Authorization: Bearer access_token }
```

#### Subscription Flow

```typescript
// 1. Get Subscription Status
GET /api/subscription/status
Headers: { Authorization: Bearer access_token }
Response: { subscription_status, subscription_tier, ... }

// 2. Create Checkout
POST /api/subscription/checkout
Headers: { Authorization: Bearer access_token }
Body: { tier: "pro" }
Response: { checkout_url, transaction_id }

// 3. Get Customer Portal URL
POST /api/subscription/portal
Headers: { Authorization: Bearer access_token }
Response: { portal_url }
```

### Environment Variables (Frontend)

```typescript
// .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080  # development
NEXT_PUBLIC_API_URL=https://your-refyne-service.railway.app  # production
```

### Token Management

```typescript
// Store tokens after login
localStorage.setItem('access_token', response.access_token)
localStorage.setItem('refresh_token', response.refresh_token)

// Add to all authenticated requests
fetch('/api/protected/me', {
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('access_token')}`
  }
})

// Refresh token when 401 received
if (response.status === 401) {
  const newToken = await refreshAccessToken(refresh_token)
  localStorage.setItem('access_token', newToken)
  // Retry original request
}
```

### Error Handling

```typescript
// Standard error response
{
  "error": "Human-readable message",
  "code": "ERROR_CODE",
  "details": { "field": "Specific error" }
}

// HTTP Status Codes
200/201  → Success
400      → Validation error (show details to user)
401      → Unauthorized (redirect to login)
409      → Conflict (e.g., email exists)
429      → Rate limited (show retry timer)
500      → Server error (log to monitoring)
```

---

## Monitoring & Debugging

### Logs

**Local Development:**
```bash
# View backend logs
make run

# View specific service logs
docker-compose logs -f refyne_db
docker-compose logs -f refyne_redis
```

**Railway:**
1. Click Refyne service → "Deployments" tab
2. Select deployment → "View logs"
3. Real-time streaming of production logs

### Health Checks

```bash
# Basic health
curl https://your-refyne-service.railway.app/api/health
# { "status": "healthy", "timestamp": "..." }

# Detailed health
curl https://your-refyne-service.railway.app/api/health/detailed
# { "status": "healthy", "checks": { "database": "healthy", "redis": "healthy" } }

# Readiness probe (for Kubernetes)
curl https://your-refyne-service.railway.app/api/health/ready

# Liveness probe (for Kubernetes)
curl https://your-refyne-service.railway.app/api/health/live
```

### Metrics

**Via Prometheus endpoint:**
```bash
# Scrape metrics
curl https://your-refyne-service.railway.app/metrics

# Returns Prometheus text format with metrics:
# - http_request_duration_seconds (latency)
# - http_requests_total (request count)
# - db_connections_active
# - redis_commands_executed
# - auth_login_attempts_total
# - paddle_transactions_total
# - rate_limit_hits_total
```

**Via Grafana Cloud:**
1. Configure scrape job pointing to `/metrics` endpoint
2. Grafana pulls metrics every 15-30 seconds
3. View dashboards automatically

### Database Debugging

**Connect to PostgreSQL directly:**

```bash
# Local via Docker
docker-compose exec db psql -U root -d refyneDB

# Remote via Railway
psql -h [PGHOST] -U [PGUSER] -d [PGDATABASE] -W

# Useful queries
\dt                          # List tables
\d users                     # User table structure
SELECT COUNT(*) FROM users;  # User count
SELECT * FROM users WHERE email = 'test@example.com';
```

**Query patterns:**

```sql
-- Active subscriptions
SELECT id, email, subscription_tier, subscription_started_at 
FROM users 
WHERE subscription_status = 'active';

-- Failed login attempts
SELECT user_id, failed_attempts, locked_until 
FROM account_security 
WHERE failed_attempts > 0;

-- Recent logins
SELECT user_id, action, ip_address, created_at 
FROM audit_logs 
WHERE action = 'login' 
ORDER BY created_at DESC 
LIMIT 10;
```

### Redis Debugging

**Connect to Redis:**

```bash
# Local via CLI
redis-cli -h localhost -p 6379 -a crashed

# Check Redis health
PING                         # Should return PONG

# View rate limiting data
KEYS *rate_limit*
GET rate_limit:ip:192.168.1.1
TTL rate_limit:ip:192.168.1.1

# View token blacklist
KEYS *blacklist*
GET blacklist:token_uuid

# Check memory usage
INFO memory
DBSIZE                       # Total keys
```

---

## Troubleshooting

### Backend Won't Start

**Problem:** `Connection refused` error

```
database/sql: driver: bad connection
```

**Solution:**
```bash
# Check PostgreSQL is running
docker-compose ps refyne_db

# If not running, start it
docker-compose up -d refyne_db

# Check logs
docker-compose logs refyne_db

# Verify connection string in .env
DB_HOST=localhost (not 127.0.0.1)
DB_PORT=5432
DB_USER=root
DB_PASSWORD=Goobs@123
```

---

### Database Migration Failed

**Problem:** `Migration error` during startup

```
error: current version does not match expected version
```

**Solution:**
```bash
# Check migrations applied
docker-compose exec db psql -U root -d refyneDB \
  -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 10;"

# If dirty=true, manual fix required:
UPDATE schema_migrations SET dirty = false WHERE version = XXXX;

# Or reset database (⚠️ destructive)
docker-compose down -v
docker-compose up -d
```

---

### Redis Connection Error

**Problem:** `dial tcp: connection refused` for Redis

```
error: failed to connect to Redis on localhost:6379
```

**Solution:**
```bash
# Check Redis is running
docker-compose ps refyne_redis

# If not running, start it
docker-compose up -d refyne_redis

# Test connection
redis-cli -h localhost -p 6379 -a crashed ping

# Check .env has correct credentials
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=crashed
```

---

### Email Not Sending

**Problem:** OTP/verification email not received

**Solution:**
```bash
# Check SMTP configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password  # NOT regular password!

# For Gmail:
# 1. Enable 2FA: https://myaccount.google.com/security
# 2. Create app password: https://myaccount.google.com/apppasswords
# 3. Copy 16-char password to SMTP_PASSWORD

# Check logs for email sending errors
make run | grep -i "email\|smtp"

# Test email locally (no actual send)
APP_ENV=development make run
# Send OTP request → check logs for debug info
```

---

### Paddle Webhook Not Received

**Problem:** Payment completed but subscription not updated

**Solution:**

1. **Verify webhook URL:**
   ```
   Paddle Dashboard → Webhooks → check URL matches your Railway domain
   ```

2. **Check logs for webhook errors:**
   ```bash
   # Railway dashboard → View logs
   # Search for "webhook" or "paddle"
   ```

3. **Verify webhook secret:**
   ```bash
   # In .env, check:
   PADDLE_SANDBOX_WEBHOOK_SECRET=ntfset_...
   # Must match Paddle dashboard exactly
   ```

4. **Check firewall:**
   ```
   Ensure /api/webhook/paddle is publicly accessible (not behind auth)
   ```

5. **Test webhook locally:**
   ```bash
   # Use Paddle's webhook testing feature in dashboard
   # Or use ngrok to expose local endpoint
   ngrok http 8080
   # Update webhook URL in Paddle to ngrok URL
   ```

---

### Rate Limiting Too Strict

**Problem:** User getting 429 (Too Many Requests)

**Solution:**
```env
# Adjust rate limit (requests per minute)
RATE_LIMIT_ENABLED=true
RATE_LIMIT_STORE=redis
RATE_LIMIT_PER_MINUTE=100  # Default, adjust as needed

# Or disable for development
RATE_LIMIT_ENABLED=false
```

---

### Token Doesn't Validate

**Problem:** `Invalid token` error when accessing protected endpoints

**Solution:**
```bash
# Check JWT_SECRET is set
JWT_SECRET=<very-long-random-string>

# Generate new secret if needed:
openssl rand -base64 64

# For Railway, ensure consistency:
# 1. Generate secret locally
# 2. Set in Rails Variables tab
# 3. Redeploy
# 4. All existing tokens will be invalid (users must re-login)
```

---

## Contact & Support

**Questions about this guide?** Check:
1. The specific subcategory above
2. Backend logs (local or Railway)
3. Docker Compose health

**Found a bug?**
Include:
- Error message (exact text)
- Steps to reproduce
- Environment (local/Railway)
- Relevant logs

---

**Last Updated:** April 17, 2026  
**Backend Version:** 1.0.0  
**Production Status:** ✅ Live on Railway  
**Maintenance:** Actively maintained
