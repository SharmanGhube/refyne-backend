# Quick Reference - Backend URLs, Credentials & Configs

**Created:** April 17, 2026  
**Updated:** Whenever infrastructure changes

---

## 🌐 Service URLs

### Development (Local)

```
Backend API:         http://localhost:8080
Frontend:            http://localhost:3000
Database (PgAdmin):  http://localhost:5050
Redis (RedisInsight):   http://localhost:5540
PostgreSQL:          localhost:5432
Redis:               localhost:6379
```

### Production (Railway)

```
Backend API:         https://your-refyne-service.railway.app
Database:            Private Railway PostgreSQL
Redis:               Private Railway Redis
Frontend:            https://your-domain.com
```

---

## 🔑 Local Development Credentials

### PostgreSQL

```
Host:     localhost
Port:     5432
User:     root
Password: Goobs@123
Database: refyneDB
SSL Mode: disable
```

### Redis

```
Host:     localhost
Port:     6379
Password: crashed
Database: 0
```

### PgAdmin

```
Email:    sharmanghube@gmail.com
Password: Goobs@123
```

### Test User

```
Email:    test@example.com
Password: TestPass123!
```

---

## 🚀 Key Environment Variables

### Backend (.env)

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASSWORD=Goobs@123
DB_NAME=refyneDB

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=crashed

# Application
APP_ENV=development
APP_PORT=8080
APP_VERSION=1.0.0

# JWT
JWT_SECRET=<generate-with-openssl-rand-base64-64>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Frontend
FRONTEND_URL=http://localhost:3000

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Payment (Paddle Sandbox)
PAYMENT_MODE=sandbox
PADDLE_SANDBOX_API_KEY=pdl_sandbox_apikey_...
PADDLE_SANDBOX_WEBHOOK_SECRET=ntfset_...
PADDLE_SANDBOX_PRODUCT_ID_PRO=pri_...

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

### Frontend (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Refyne
NEXT_PUBLIC_APP_VERSION=1.0.0
```

---

## 📊 Database Tables (20+ Tables)

```
✅ users                    - Main user account
✅ user_settings            - User preferences
✅ verification_tokens      - Email verification
✅ password_reset_tokens    - Password reset
✅ account_security         - Login attempts, lockouts
✅ audit_logs              - Login/logout history
✅ device_sessions         - Active device sessions
✅ workspaces              - User workspaces
✅ workspace_members       - Team members
✅ otto_conversations      - AI chat conversations
✅ otto_messages           - Chat messages
✅ instagram_accounts      - Instagram integration
✅ ... (more for media, context, etc.)
```

---

## 🔌 Critical API Endpoints (Frontend Must Implement)

### Authentication

```
POST   /api/auth/register              - User registration
POST   /api/auth/verify                - Email verification
POST   /api/auth/request-otp           - Request OTP
POST   /api/auth/login                 - Login with OTP
POST   /api/auth/refresh               - Refresh token
POST   /api/auth/logout                - Logout current session
POST   /api/auth/logout-all            - Logout all devices
POST   /api/auth/forgot-password       - Request password reset
POST   /api/auth/reset-password        - Reset password
```

### Subscription (Paddle)

```
GET    /api/subscription/status        - Get subscription status
POST   /api/subscription/checkout      - Create Paddle checkout
POST   /api/subscription/portal        - Get customer portal URL
POST   /api/webhook/paddle             - Paddle webhook (internal)
```

### User Profile

```
GET    /api/user/profile               - Get user profile
PUT    /api/user/profile               - Update profile
GET    /api/user/settings              - Get user settings
PUT    /api/user/settings              - Update settings
```

### Health Checks

```
GET    /api/health                     - Basic health check
GET    /api/health/detailed            - Health with components
GET    /api/health/ready               - Readiness probe
GET    /api/health/live                - Liveness probe
```

### Protected Endpoints

```
GET    /api/protected/me               - Current user info
```

---

## 🔐 Token Values & Expiry

```
Access Token:
- Type: JWT (JSON Web Token)
- Expiry: 15 minutes (900 seconds)
- Stored: localStorage in browser
- Header: Authorization: Bearer <token>

Refresh Token:
- Type: UUID
- Expiry: 7 days
- Stored: localStorage in browser
- Used to get new access token

JWT Secret:
- Generated: openssl rand -base64 64
- Stored: Backend .env (JWT_SECRET)
- Used to: Sign/verify JWT tokens
```

---

## 💳 Paddle Sandbox Configuration

### Sandbox URLs

```
Dashboard:         https://sandbox-vendors.paddle.com
API:               https://sandbox-api.paddle.com
Checkout:          https://sandbox-checkout.paddle.com
```

### Sandbox Credentials (Example)

```
API Key:           pdl_sandbox_apikey_01kmg59k2jfaj5aqfev0m590dr_CsjSN419gAPSNqNjCVjD7J_AGa
Webhook Secret:    ntfset_01kmg5p43qm4py8d9rqganehx9
Product ID (Pro):  pri_01kb65b3gzy2xn21nh0zw922yn
```

### Test Cards

```
Visa (Success):    4111 1111 1111 1111 (12/25, 123)
Visa (Decline):    4000 0000 0000 0002 (12/25, 123)
Mastercard:        5555 5555 5555 4444 (12/25, 123)
```

---

## 🐳 Docker Compose Commands

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View service status
docker-compose ps

# View logs
docker-compose logs -f refyne_db
docker-compose logs -f refyne_redis
docker-compose logs -f refyne_pgadmin

# Restart service
docker-compose restart refyne_db

# Connect to PostgreSQL CLI
docker-compose exec db psql -U root -d refyneDB

# Remove all data (⚠️ Destructive)
docker-compose down -v
```

---

## 📈 Performance Thresholds

```
Database:
- Max Connections:    20
- Idle Connections:   10
- Connection Timeout: 15 minutes

Redis:
- Used for:           Rate limiting, token blacklist
- Memory:             Configurable (default unlimited)
- Persistence:        AOF (Append-Only File)

Rate Limiting:
- Default:            100 requests/minute
- Per endpoint:       3 requests/15 minutes (OTP, verification)
- Lockout:            5 failed attempts = 15 min lock
```

---

## ✅ Health Check Status Codes

```
200  Healthy ✅
503  Unhealthy ❌

Quick health check:
curl http://localhost:8080/api/health
# { "status": "healthy" }

Detailed health:
curl http://localhost:8080/api/health/detailed
# Shows database, redis, queue status
```

---

## 🔍 Debugging Commands

```bash
# View backend logs
APP_ENV=development make run

# Check if Redis is running
redis-cli -h localhost -p 6379 -a crashed PING

# Check PostgreSQL connection
psql -h localhost -U root -d refyneDB -c "SELECT 1"

# View active connections
docker-compose ps

# Check specific service logs
docker-compose logs refyne_db -t -f
```

---

## 🚢 Railway Deployment Quick Steps

```
1. Push to main branch
   git push origin main

2. Railway auto-detects:
   - GitHub repository
   - Dockerfile
   - railway.json

3. Automatic deployment:
   - Build Docker image
   - Run migrations
   - Deploy to production
   - Show deployment status

4. View logs in Railway dashboard:
   - Click service → Deployments → View logs
```

---

## 📋 Subscription Tiers (Current)

```
Current Setup: SINGLE PRO TIER ONLY (as of Apr 17, 2026)

Tier: PRO
Price: Custom (customer decides)
Billing: Monthly
Features: All features
Product ID: pri_01kb65b3gzy2xn21nh0zw922yn (sandbox)
```

---

## 🔗 File Locations

```bash
# Backend root
/path/to/refyne-backend/

# Environment configs
.env                          # Local development
railway.env.template          # Railway template

# Database
internal/database/migrations/sql/  # Migration files

# Documentation
docs/INFRASTRUCTURE_COMPLETE_GUIDE.md
docs/FRONTEND_CONFIG_GUIDE.md
docs/FRONTEND_API_INTEGRATION.md
docs/RAILWAY_SETUP.md
docs/PADDLE_SANDBOX_SETUP.md

# Configuration
cmd/main.go                   # Entry point
internal/config/             # All configuration
internal/dependencies/        # Wire DI setup
```

---

## 🆚 Local vs Production

| Aspect | Local | Production |
|--------|-------|-----------|
| **Database** | Docker PostgreSQL | Railway PostgreSQL |
| **Redis** | Docker Redis | Railway Redis |
| **URL** | localhost:8080 | your-domain.railway.app |
| **SSL** | Disabled | Required |
| **Logging** | Console only | Railway dashboard + structured logs |
| **Monitoring** | None | Prometheus + Grafana |
| **Payment Mode** | sandbox | production |
| **CORS** | localhost:3000 | your-frontend-domain.com |

---

## 🎯 Common Tasks (Quick Reference)

### Start fresh development session

```bash
# 1. Start Docker services
docker-compose up -d

# 2. Verify services are running
docker-compose ps

# 3. Check database is accessible
psql -h localhost -U root -d refyneDB -c "SELECT 1"

# 4. Start backend
APP_ENV=development make run

# 5. In another terminal, start frontend
cd ../refyne-frontend
npm run dev
```

### Deploy to production

```bash
# 1. Ensure all changes are committed
git status

# 2. Push to main branch
git push origin main

# 3. Railway auto-deploys
# 4. Verify deployment
curl https://your-refyne-service.railway.app/api/health
```

### Check deployment status

```bash
# View recent deployments
# Go to Railway dashboard → Deployments tab

# View live logs
# Railway dashboard → Click latest deployment → View logs

# Check health
curl https://your-refyne-service.railway.app/api/health
```

### Access production database

```bash
# Get credentials from Railway → PostgreSQL service → Variables

psql -h <PGHOST> -U <PGUSER> -d <PGDATABASE> -W
# Enter password when prompted

# Common queries
SELECT COUNT(*) FROM users;
SELECT * FROM users WHERE email = 'test@example.com';
```

---

## 📞 Contact & Support

**Issues with:**
- Backend: Check backend logs (`make run`)
- Database: Check PostgreSQL connection
- Redis: Check Redis CLI (`redis-cli PING`)
- Deployment: Check Railway dashboard
- API: Check endpoint response with `curl`

**Generate new secrets:**

```bash
# New JWT secret
openssl rand -base64 64

# New UUID
python3 -c "import uuid; print(uuid.uuid4())"
```

---

**Last Updated:** April 17, 2026  
**Format:** Quick Reference Card (print-friendly)
