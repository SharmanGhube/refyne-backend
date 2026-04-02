# Railway Deployment Setup Guide

## Overview

This guide explains how to deploy Refyne Backend to Railway with automatic environment variable synchronization from linked services (PostgreSQL, Redis).

## Prerequisites

- Railway account (https://railway.app)
- GitHub repository with Refyne Backend pushed
- Paddle Sandbox credentials (from https://sandbox-vendors.paddle.com)
- Gmail SMTP credentials (for email notifications)

## Step 1: Create Railway Project

1. Go to https://railway.app
2. Click "New Project"
3. Select "Deploy from GitHub" → Authorize → Select `refyne-backend` repository
4. Railway will auto-detect `railway.json` configuration

## Step 2: Add PostgreSQL Service

1. In Railway dashboard, click "New" in your project
2. Add plugin: **PostgreSQL**
3. Railway auto-creates:
   - `PGHOST` (database host, e.g., `postgres.railway.internal`)
   - `PGPORT` (e.g., `5432`)
   - `PGUSER` (auto-generated username)
   - `PGPASSWORD` (auto-generated password)
   - `PGDATABASE` (auto-generated database name)

4. Link PostgreSQL to the Refyne service:
   - Click PostgreSQL service
   - Click "Connect" → Select Refyne service
   - This makes all `PGHOST`, `PGPORT`, etc. available to the app

## Step 3: Add Redis Service

1. In Railway dashboard, click "New"
2. Add plugin: **Redis**
3. Railway auto-creates:
   - `REDIS_HOST` (e.g., `redis.railway.internal`)
   - `REDIS_PORT` (e.g., `6379`)
   - `REDIS_PASSWORD` (auto-generated password)

4. Link Redis to the Refyne service:
   - Click Redis service
   - Click "Connect" → Select Refyne service

## Step 4: Configure Environment Variables

### Option A: Auto-Sync with Service References (Recommended)

1. Click Refyne service → Variables tab
2. Click "Add" and import variables from `railway.env.template`:
   - All `${{Postgres.PGHOST}}` references auto-populate when services are linked
   - All `${{Redis.REDIS_HOST}}` references auto-populate when services are linked
   - Manual variables (API keys, secrets) need to be entered manually

3. **Required manual variables:**

```env
# Application
APP_ENV=production
APP_PORT=8080

# JWT (generate with: openssl rand -base64 64)
JWT_SECRET=<your-generated-secret>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Frontend URL
FRONTEND_URL=https://your-frontend.com

# Email (Gmail SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password-here
SMTP_FROM=noreply@refyne.com

# Paddle Sandbox
PAYMENT_MODE=sandbox
PADDLE_SANDBOX_API_KEY=<your-key-from-sandbox-vendors.paddle.com>
PADDLE_SANDBOX_WEBHOOK_SECRET=<your-webhook-secret>
PADDLE_SANDBOX_PRODUCT_ID_STARTER=<pricer-id>
PADDLE_SANDBOX_PRODUCT_ID_PROFESSIONAL=<pricer-id>
PADDLE_SANDBOX_PRODUCT_ID_BUSINESS=<pricer-id>
PADDLE_SANDBOX_PRODUCT_ID_ENTERPRISE=<pricer-id>

# Other
RATE_LIMIT_ENABLED=true
RATE_LIMIT_STORE=redis
LOG_LEVEL=info
AUTO_MIGRATE=true
```

### Option B: Manual Environment Variables

If auto-sync isn't working:

1. Get PostgreSQL credentials from Railway:
   - PostgreSQL service → Variables tab → Copy all `PG*` variables

2. Get Redis credentials from Railway:
   - Redis service → Variables tab → Copy all `REDIS_*` variables

3. Manually enter them in Refyne service → Variables tab using exact names:
   - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`
   - `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`

## Step 5: Verify Deployment

### Check Application Logs
1. Click Refyne service → Deployments tab
2. Select latest deployment → View logs
3. Look for:
   ```
   Database connection pool initialized successfully
   Redis connection established successfully
   Running in sandbox payment mode
   Server started successfully on :8080
   ```

### Health Check Endpoint
Once deployed, verify the app is running:
```bash
curl https://your-refyne-service.railway.app/api/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": "2026-04-02T10:30:00Z"
}
```

## Step 6: Configure Domain & Custom URL (Optional)

1. Click Refyne service → Settings
2. Under "Networking" → "Public Networking"
3. Click "Generate Domain" to get a Railway domain
4. Or add custom domain with CNAME pointing to Railway domain

## Troubleshooting

### Environment Variables Not Auto-Syncing

**Problem:** `${{Postgres.PGHOST}}` type variables not being replaced
- **Solution:** Ensure PostgreSQL and Redis services are **linked** to Refyne service
  - Go to PostgreSQL service → Click "Connect" → Select Refyne service
  - Go to Redis service → Click "Connect" → Select Refyne service

### Connection Refused Error

**Problem:** `dial tcp [::1]:6379: connect: connection refused`
- **Root cause:** Using localhost instead of Railway service hostname
- **Solution:** Ensure `REDIS_HOST=${{Redis.REDIS_HOST}}` (not hardcoded `localhost`)
- **Verification:** Check Refyne service logs show `redis.railway.internal` in connection string

### Database Migration Errors

**Problem:** App crashes with migration errors
- **Solution:** Ensure `AUTO_MIGRATE=true` is set and database is initialized
- **Debug:** SSH into Railway container or check logs for SQL errors

### App Shows "Production Off"

**Problem:** App was running but is now showing as offline
- **Solution:** Check recent deployments for errors
  - Deployments tab → Select recent deployment → View logs
- **Common causes:** Missing env vars, database connection timeout, Redis unavailable

## Next Steps

After successful deployment:

1. **Test authentication flow** - See `docs/TESTING.md`
2. **Monitor with Grafana** - See `docs/MONITORING.md`
3. **Set up Paddle webhooks** - See `docs/PADDLE_SANDBOX_SETUP.md`

## Reference: Variable Mapping

When services are linked, Railway automatically maps:

| Railway Service Variable | Refyne Environment Variable |
|---|---|
| `Postgres.PGHOST` | `DB_HOST` |
| `Postgres.PGPORT` | `DB_PORT` |
| `Postgres.PGUSER` | `DB_USER` |
| `Postgres.PGPASSWORD` | `DB_PASSWORD` |
| `Postgres.PGDATABASE` | `DB_NAME` |
| `Redis.REDIS_HOST` | `REDIS_HOST` |
| `Redis.REDIS_PORT` | `REDIS_PORT` |
| `Redis.REDIS_PASSWORD` | `REDIS_PASSWORD` |

These are automatically injected when you use `${{ServiceName.VARIABLE}}` syntax in `railway.env.template`.
