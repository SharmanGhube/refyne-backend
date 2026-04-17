# 📚 Refyne Backend Documentation Index

**Last Updated:** April 17, 2026  
**Status:** ✅ Complete & Production-Ready

---

## 👋 Start Here

If you're new to Refyne Backend:
1. **Quick Overview** → Read this file
2. **Your Role?** → Jump to section below
3. **Specific Question?** → Use the table of contents

---

## 👥 Choose Your Path

### I'm a **Frontend Developer**

Start here → [`FRONTEND_CONFIG_GUIDE.md`](FRONTEND_CONFIG_GUIDE.md)

Then read:
- [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) - Keep handy for quick lookups
- [`FRONTEND_API_INTEGRATION.md`](FRONTEND_API_INTEGRATION.md) - Deep dive into all endpoints

**Why this order:**
1. FRONTEND_CONFIG_GUIDE shows authentication & payment flows with code examples
2. QUICK_REFERENCE has URLs, credentials, and common commands
3. FRONTEND_API_INTEGRATION is exhaustive reference (2000+ lines)

---

### I'm a **Backend Developer / DevOps**

Start here → [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md)

Then read specific sections:
- Local Development Setup (Docker, PostgreSQL, Redis)
- Railway Deployment (how production is set up)
- Paddle Integration (payment processing)
- Monitoring & Debugging (logs, metrics, troubleshooting)

**Keep handy:**
- [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) - URLs, credentials, commands
- [`RAILWAY_SETUP.md`](RAILWAY_SETUP.md) - Railway-specific details

---

### I'm **Deploying to Production**

1. Backend deployment → [`RAILWAY_SETUP.md`](RAILWAY_SETUP.md)
2. Payment processing → [`PADDLE_SANDBOX_SETUP.md`](PADDLE_SANDBOX_SETUP.md)
3. Monitoring setup → [`MONITORING.md`](MONITORING.md)
4. CI/CD pipeline → [`CI_CD_SETUP.md`](CI_CD_SETUP.md)

---

### I'm **Troubleshooting an Issue**

1. Check [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](#troubleshooting) → Troubleshooting section
2. Check [`QUICK_REFERENCE.md`](#-debugging-commands) → Debugging commands
3. View logs: `make run` (local) or Railway dashboard (production)
4. Check backend health: `curl http://localhost:8080/api/health/detailed`

---

## 📋 Complete Documentation Map

### 🚀 Deployment & Infrastructure

| Document | For | Purpose |
|----------|-----|---------|
| **INFRASTRUCTURE_COMPLETE_GUIDE.md** | Backend/DevOps | Everything about deployment, databases, Redis, Railway, Paddle |
| **RAILWAY_SETUP.md** | Backend/DevOps | Railway-specific deployment steps |
| **PADDLE_SANDBOX_SETUP.md** | Backend/DevOps | Paddle payment integration setup |
| **DEPLOYMENT.md** | DevOps | Original deployment guide |
| **CI_CD_SETUP.md** | DevOps | GitHub Actions pipeline configuration |

### 🔌 API & Frontend Integration

| Document | For | Purpose |
|----------|-----|---------|
| **FRONTEND_CONFIG_GUIDE.md** | Frontend Dev | Quick start + authentication + subscriptions + error handling |
| **FRONTEND_API_INTEGRATION.md** | Frontend Dev | Complete API reference (all endpoints, errors, examples) |
| **QUICK_REFERENCE.md** | Everyone | URLs, credentials, commands (quick lookup) |

### 📊 Monitoring & Operations

| Document | For | Purpose |
|----------|-----|---------|
| **MONITORING.md** | DevOps/Backend | Prometheus + Grafana setup |
| **GRAFANA_CLOUD_SETUP.md** | DevOps | Cloud monitoring configuration |
| **ALERTING.md** | DevOps | Alert rules and notifications |

### 🏗️ Architecture & Reference

| Document | For | Purpose |
|----------|-----|---------|
| **DATA_MODELS.md** | Backend Dev | Database schema & relationships |
| **DEVELOPMENT_WORKFLOW.md** | Backend Dev | Local development setup & workflow |
| **DATABASE_SECURITY.md** | Backend Dev | Database security best practices |

### 🎯 Feature-Specific

| Document | For | Purpose |
|----------|-----|---------|
| **INSTAGRAM_INTEGRATION.md** | Backend Dev | Instagram API integration |
| **USER_JOURNEY_COMPLETE.md** | Designer/Product | User workflows and experience |
| **TESTING.md** | Backend Dev | End-to-end test suite |

---

## 🎯 Common Tasks - Which Document?

### "I need to connect frontend to backend"
→ [`FRONTEND_CONFIG_GUIDE.md`](FRONTEND_CONFIG_GUIDE.md) + [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md)

### "How do I test OTP login?"
→ [`FRONTEND_API_INTEGRATION.md`](FRONTEND_API_INTEGRATION.md) section 1.4-1.5 + [`FRONTEND_CONFIG_GUIDE.md`](FRONTEND_CONFIG_GUIDE.md) auth flow

### "How do I set up Paddle for testing?"
→ [`PADDLE_SANDBOX_SETUP.md`](PADDLE_SANDBOX_SETUP.md) or [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) "Paddle Integration" section

### "How do I deploy to Railway?"
→ [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) "Railway Deployment" section or [`RAILWAY_SETUP.md`](RAILWAY_SETUP.md)

### "Backend won't start - help!"
→ [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) "Troubleshooting" section

### "What API endpoints are available?"
→ [`FRONTEND_API_INTEGRATION.md`](FRONTEND_API_INTEGRATION.md) (complete) or [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) (quick summary)

### "What are the database tables?"
→ [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) "Database Architecture" or [`DATA_MODELS.md`](DATA_MODELS.md)

### "How do I debug production issues?"
→ [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) "Monitoring & Debugging" section

### "What credentials do I need?"
→ [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) "Local Development Credentials" section

### "How do I run tests?"
→ [`TESTING.md`](TESTING.md)

---

## 📍 Key Locations Reference

### Development URLs (Local)

```
Backend:             http://localhost:8080
Frontend:            http://localhost:3000
PostgreSQL:          localhost:5432
Redis:               localhost:6379
PgAdmin:             http://localhost:5050
RedisInsight:        http://localhost:5540
```

See [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) for credentials.

### Production URLs (Railway)

```
Backend:      https://your-refyne-service.railway.app
Database:     Private Railway PostgreSQL
Redis:        Private Railway Redis
Frontend:     https://your-domain.com
```

### Important File Paths

```
Documentation:  docs/               (you are here)
Backend config: .env               (main config)
Environment:    railway.env.template (Railroad deployment)
Database:       internal/database/migrations/sql/
DI Setup:       internal/dependencies/
API Routes:     internal/api/handlers/
```

---

## 🔐 Authentication & Security

### Token Management

See [`FRONTEND_CONFIG_GUIDE.md`](#-token-lifecycle) for token lifecycle details.

```
Access Token:    15 minutes expiry (JWT)
Refresh Token:   7 days expiry (UUID)
Rate Limiting:   100 requests/minute globally
Login Lockout:   5 failed attempts = 15 min lock
```

### API Key Management

All API keys should be stored in `.env` (local) or Railway Variables (production).

**Never commit secrets to Git.** The `.gitignore` already excludes `.env`.

---

## 🐳 Local Development

### Start Services

```bash
# Start Docker Compose (PostgreSQL, Redis, etc.)
docker-compose up -d

# Verify services running
docker-compose ps

# Start backend
make run
```

**Full guide:** [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) → Local Development Setup

---

## 🚀 Deployment

### Railway Deployment

1. Push to `main` branch
2. Railway auto-deploys
3. View status in Railway dashboard

**Full guide:** [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) → Railway Deployment

### Local Docker

```bash
# Build image
docker build -t refyne-backend .

# Run container
docker run -p 8080:8080 --env-file .env refyne-backend
```

---

## 💳 Payment Processing

### Paddle Integration

Current setup: **Sandbox mode for testing, Production mode for live**

**To set up Paddle:**
→ [`PADDLE_SANDBOX_SETUP.md`](PADDLE_SANDBOX_SETUP.md) or [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) "Paddle Integration" section

**Subscription tier:** Single "Pro" tier ($0 or custom price)

---

## 📊 Monitoring

### Health Checks

```bash
# Basic
curl http://localhost:8080/api/health

# Detailed
curl http://localhost:8080/api/health/detailed

# Readiness
curl http://localhost:8080/api/health/ready

# Liveness
curl http://localhost:8080/api/health/live
```

**Full guide:** [`MONITORING.md`](MONITORING.md) + [`GRAFANA_CLOUD_SETUP.md`](GRAFANA_CLOUD_SETUP.md)

---

## 🔍 API Endpoints (Quick Summary)

### Authentication
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/refresh
```

### Subscriptions
```
GET    /api/subscription/status
POST   /api/subscription/checkout
POST   /api/subscription/portal
```

### User Profile
```
GET    /api/user/profile
PUT    /api/user/profile
GET    /api/user/settings
PUT    /api/user/settings
```

**Complete list:** [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) → "Critical API Endpoints" or [`FRONTEND_API_INTEGRATION.md`](FRONTEND_API_INTEGRATION.md) for exhaustive reference

---

## ❓ FAQ

### "Which document should I read?"

**If you're:**
- **Frontend Dev** → Start with [`FRONTEND_CONFIG_GUIDE.md`](FRONTEND_CONFIG_GUIDE.md)
- **Backend Dev** → Start with [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md)
- **DevOps** → Start with [`RAILWAY_SETUP.md`](RAILWAY_SETUP.md) + [`CI_CD_SETUP.md`](CI_CD_SETUP.md)
- **Deploying** → Follow [`DEPLOYMENT.md`](DEPLOYMENT.md) or [`RAILWAY_SETUP.md`](RAILWAY_SETUP.md)

### "I don't know what to read, where do I start?"

→ [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) is the fastest way to get essential info

### "Which API endpoint should I use?"

→ Check [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) → "Critical API Endpoints" or [`FRONTEND_API_INTEGRATION.md`](FRONTEND_API_INTEGRATION.md) for details

### "How do I debug locally?"

→ [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) → "Troubleshooting" section

### "What's the database structure?"

→ [`DATA_MODELS.md`](DATA_MODELS.md) or [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) → "Database Architecture"

---

## 📞 Support & Questions

**Before asking:**
1. Check [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) for quick lookup
2. Search relevant guide for your question
3. Check troubleshooting section
4. View logs: `make run` or Railway dashboard

**Need help?** Contact backend team with:
- What you're trying to do
- Error message (exact text)
- Steps to reproduce
- Environment (local/testing/production)

---

## 🎓 Learning Path

**New to Refyne Backend? Follow this order:**

1. **Read:** [`FRONTEND_CONFIG_GUIDE.md`](FRONTEND_CONFIG_GUIDE.md) (15 min)
   - Overview of how backend works
   - Authentication flow
   - Payment flow

2. **Bookmark:** [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md)
   - Keep this open while coding
   - URLs, credentials, commands

3. **Deep Dive:** [`FRONTEND_API_INTEGRATION.md`](FRONTEND_API_INTEGRATION.md) or [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md)
   - Depending on your role

4. **Experiment:** Set up locally
   - Start Docker Compose
   - Run backend: `make run`
   - Test API endpoints

5. **Deploy:** Follow deployment guide
   - [`RAILWAY_SETUP.md`](RAILWAY_SETUP.md) for production
   - [`CI_CD_SETUP.md`](CI_CD_SETUP.md) for automation

---

## 📋 Document Versions

| Document | Version | Last Updated | Status |
|----------|---------|--------------|--------|
| INFRASTRUCTURE_COMPLETE_GUIDE.md | 1.0 | Apr 17, 2026 | ✅ Complete |
| FRONTEND_CONFIG_GUIDE.md | 1.0 | Apr 17, 2026 | ✅ Complete |
| QUICK_REFERENCE.md | 1.0 | Apr 17, 2026 | ✅ Complete |
| FRONTEND_API_INTEGRATION.md | 1.0 | Nov 30, 2025 | ✅ Complete |
| RAILWAY_SETUP.md | 1.0 | Apr 2, 2026 | ✅ Complete |
| PADDLE_SANDBOX_SETUP.md | 1.0 | Apr 3, 2026 | ✅ Complete (needs tier update) |
| MONITORING.md | 1.0 | Apr 4, 2026 | ✅ Complete |
| TESTING.md | 1.0 | Apr 4, 2026 | ✅ Complete |
| DATA_MODELS.md | 1.0 | Latest | ✅ Current |
| DEPLOYMENT.md | 1.0 | Latest | ✅ Current |

---

## 🎯 Next Steps

1. **Choose your path** (see above - Frontend/Backend/DevOps)
2. **Read relevant guide** (takes 15-30 minutes)
3. **Try it locally** (start Docker, run backend)
4. **Test API endpoints** (use curl or Postman)
5. **Ask questions** if stuck

---

**Status:** ✅ All documentation complete and production-ready  
**Last Reviewed:** April 17, 2026  
**Backend Version:** 1.0.0

---

## 📚 Full Document List

- [`INFRASTRUCTURE_COMPLETE_GUIDE.md`](INFRASTRUCTURE_COMPLETE_GUIDE.md) - NEW ⭐ Start here for complete overview
- [`FRONTEND_CONFIG_GUIDE.md`](FRONTEND_CONFIG_GUIDE.md) - NEW ⭐ Frontend developer guide
- [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md) - NEW ⭐ Quick lookup card
- [`FRONTEND_API_INTEGRATION.md`](FRONTEND_API_INTEGRATION.md) - Complete API reference
- [`RAILWAY_SETUP.md`](RAILWAY_SETUP.md) - Railway deployment
- [`PADDLE_SANDBOX_SETUP.md`](PADDLE_SANDBOX_SETUP.md) - Payment integration
- [`DEVELOPMENT_WORKFLOW.md`](DEVELOPMENT_WORKFLOW.md) - Local dev setup
- [`DEPLOYMENT.md`](DEPLOYMENT.md) - General deployment
- [`CI_CD_SETUP.md`](CI_CD_SETUP.md) - GitHub Actions pipeline
- [`MONITORING.md`](MONITORING.md) - Prometheus + Grafana
- [`GRAFANA_CLOUD_SETUP.md`](GRAFANA_CLOUD_SETUP.md) - Cloud monitoring
- [`ALERTING.md`](ALERTING.md) - Alert rules
- [`TESTING.md`](TESTING.md) - E2E tests
- [`DATA_MODELS.md`](DATA_MODELS.md) - Database schema
- [`DATABASE_SECURITY.md`](DATABASE_SECURITY.md) - Security best practices
- [`INSTAGRAM_INTEGRATION.md`](INSTAGRAM_INTEGRATION.md) - Instagram API
- [`USER_JOURNEY_COMPLETE.md`](USER_JOURNEY_COMPLETE.md) - User workflows

---

**👉 Ready? Pick a document and dive in!**
