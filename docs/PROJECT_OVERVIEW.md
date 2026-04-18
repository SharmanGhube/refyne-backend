# Refyne - Community Growth Platform: Project Overview

## Executive Summary

**Refyne** is an AI-powered SaaS application designed to help social media creators and businesses manage, analyze, and grow their communities across platforms. The backend is a production-ready REST API built with Go, deployed on Railway, and integrated with Instagram, AI analysis tools, and payment processing.

**Status:** ✅ Live in Production  
**Deployment:** Railway (PostgreSQL, Redis, Email Services)  
**Language:** Go 1.23  
**Framework:** Gin HTTP Framework  

---

## What Does Refyne Do?

Refyne empowers users to:

### 1. **Authenticate & Manage Accounts**
   - Secure user registration and login (OTP, password, social auth ready)
   - JWT-based access & refresh tokens (15min / 7-day expiry)
   - Multi-factor security (account lockout, password hashing, audit logging)
   - Account deletion and data management compliance

### 2. **Connect Social Media Platforms**
   - OAuth-based Instagram account connection
   - Media synchronization (photos, videos, metadata)
   - Real-time webhook handling for Instagram events
   - Analytics and engagement metrics extraction

### 3. **AI-Powered Content Analysis**
   - AI conversation interface for community insights
   - Message-based Q&A with AI assistant (Otto AI)
   - Feedback system for AI response quality improvement
   - Context-aware analysis using connected media and data

### 4. **Manage Teams & Permissions**
   - Workspace creation and management
   - Role-based access control (Owner, Member)
   - Team member invitations via email
   - Workspace-level subscription management

### 5. **Subscribe & Payment Processing**
   - Paddle-based subscription management
   - Single "Pro" tier with feature access
   - Webhook-driven subscription lifecycle (new, renew, cancel, churn recovery)
   - Tax-compliant payment handling

### 6. **Receive Notifications & Alerts**
   - Email notifications for key events (invitations, subscription updates)
   - Structured logging and audit trails
   - Monitoring and alerting (Prometheus + Grafana Cloud)

---

## Architecture Overview

### Domain-Driven Design (DDD) Structure

Refyne uses **Domain-Driven Design** with Google Wire for dependency injection. Each domain encapsulates a business capability:

```
├── auth/
│   ├── handler/        (HTTP endpoints)
│   ├── service/        (business logic)
│   ├── repository/     (data access)
│   └── model/          (domain entities)
│
├── user/               (Profile, settings, onboarding)
├── subscription/       (Paddle integration, tiers)
├── email/              (SMTP service, background jobs)
├── workspace/          (Team management, roles)
├── otto/               (AI assistant conversations)
├── instagram/          (OAuth, media sync, webhooks)
├── ai/                 (AI assistant registry)
├── context/            (Context documents for AI)
└── notification/       (User alerts & events)
```

Each domain follows the same layered architecture:
- **HTTP Handler:** Request/response handling, validation
- **Service:** Core business logic, orchestration
- **Repository:** Database operations
- **Model:** Domain entities and value objects

### API Response Format

All endpoints follow a standardized **response envelope**:

```json
{
  "success": true,
  "code": 200,
  "message": "Success",
  "data": { /* actual response data */ },
  "error": null,
  "meta": {
    "timestamp": "2026-04-18T10:30:00Z",
    "request_id": "abc123"
  }
}
```

Error responses include:
```json
{
  "success": false,
  "code": 400,
  "message": "Bad Request",
  "data": null,
  "error": { "field": "email", "message": "Invalid email format" },
  "meta": { /* ... */ }
}
```

---

## Tech Stack

| Component | Technology | Details |
|-----------|-----------|---------|
| **Language** | Go 1.23 | High-performance, concurrent |
| **Framework** | Gin | RESTful HTTP with middleware support |
| **Database** | PostgreSQL | pgx driver v5, migrations auto-run |
| **Cache/Session** | Redis | Rate limiting, session management |
| **Job Queue** | River | PostgreSQL-based async jobs |
| **DI/Wiring** | Google Wire | Compile-time dependency injection |
| **Authentication** | JWT | Access + refresh tokens, OTP |
| **Payment** | Paddle | Subscription management & webhooks |
| **Email** | SMTP | Background job-driven via River |
| **Logging** | Zap | Structured logging with levels |
| **Monitoring** | Prometheus + Grafana Cloud | Metrics collection, visualization |
| **Deployment** | Docker + Railway | Alpine-based images, auto-deploy CI/CD |

---

## Key Features

### ✅ Authentication & Security
- User registration with email verification
- OTP login (secure, email-only)
- Password reset with token validation
- JWT token management (automatic refresh)
- Account lockout (5 failed attempts = 15-min block)
- Audit logging for all auth events
- Token blacklist for logout

### ✅ User Management
- Profile management (name, username, avatar)
- Settings management (language, timezone, notifications)
- Onboarding flow with completion tracking
- Account deletion (soft delete for data retention)
- User preferences persistence

### ✅ Team & Workspace
- Unlimited workspace creation per user
- Role-based access (Owner, Member)
- Member invitation system (email-based)
- Workspace-level subscription awareness
- Audit logging of workspace changes

### ✅ Instagram Integration
- OAuth 2.0 connection flow
- Media synchronization (feed, stories, reels)
- Webhook handling (real-time updates)
- Engagement metrics (likes, comments, shares)
- AI-powered content analysis
- **17 endpoints** for Instagram operations

### ✅ AI Assistant (Otto)
- Conversation-based interface
- Message threading and context awareness
- Feedback mechanism for response quality
- **11 endpoints** for AI operations
- Context documents for personalized responses

### ✅ Subscription Management
- Paddle webhook integration
- Single "Pro" tier (simplified model)
- Subscription lifecycle handling (new, renew, cancel)
- Churn recovery campaigns
- Invoice webhook processing

### ✅ Email Service
- SMTP integration (Gmail with app passwords)
- Background job queue (River)
- Templated emails (invitations, password reset, notifications)
- Retry logic for failed sends

### ✅ Rate Limiting & DDoS Protection
- In-memory + Redis-backed rate limiting
- Global limit: 100 requests/minute (configurable)
- Per-endpoint overrides available
- Account lockout for auth attempts
- Automatic cleanup of expired limits

### ✅ Monitoring & Observability
- **Prometheus metrics** (HTTP, database, auth, payment)
- **Grafana Cloud** integration for dashboards
- **Health check endpoints** (basic, detailed, ready, live)
- **Structured logging** with request IDs and tracing
- **Alert rules** with 5 severity levels (critical, warning, info, etc.)

### ✅ CI/CD & Deployment
- **GitHub Actions** CI (linting, testing, security scan)
- **Automated Railway deployment** on main branch push
- **Multi-stage Docker build** (80.8MB production image)
- **Database migrations** (auto-run on startup if `AUTO_MIGRATE=true`)
- **Comprehensive test coverage** (unit + E2E)

---

## API Endpoints Summary

### Authentication (20+ endpoints)
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - Password login
- `POST /api/auth/otp/send` - Send OTP
- `POST /api/auth/otp/verify` - Verify OTP and login
- `POST /api/auth/refresh` - Refresh access token
- `POST /api/auth/logout` - Logout (token blacklist)
- `POST /api/auth/password/reset/request` - Request password reset
- `POST /api/auth/password/reset/confirm` - Confirm password reset
- `POST /api/auth/verify/email/resend` - Resend verification email

### User Profile & Settings (8 endpoints)
- `GET /api/user/profile` - Get user profile
- `PUT /api/user/profile` - Update profile
- `GET /api/user/settings` - Get user settings
- `PUT /api/user/settings` - Update settings
- `POST /api/user/onboarding/complete` - Mark onboarding done
- `DELETE /api/user/account` - Delete account

### Workspaces & Teams (8 endpoints)
- `POST /api/workspaces` - Create workspace
- `GET /api/workspaces` - List workspaces
- `GET /api/workspaces/:id` - Get workspace
- `PUT /api/workspaces/:id` - Update workspace
- `DELETE /api/workspaces/:id` - Delete workspace
- `POST /api/workspaces/:id/members` - Invite member
- `GET /api/workspaces/:id/members` - List members
- `DELETE /api/workspaces/:id/members/:user_id` - Remove member

### Instagram Integration (17 endpoints)
- `GET /api/instagram/auth/url` - Get OAuth URL
- `POST /api/instagram/auth/callback` - Handle OAuth callback
- `GET /api/instagram/media` - List user media
- `GET /api/instagram/media/:id` - Get media details
- `POST /api/instagram/media/sync` - Sync media
- `GET /api/instagram/analytics` - Get analytics
- `POST /api/instagram/webhooks` - Webhook receiver
- And 10+ more for media details, engagement, etc.

### AI Assistant (Otto) (11 endpoints)
- `POST /api/otto/conversations` - Create conversation
- `GET /api/otto/conversations` - List conversations
- `GET /api/otto/conversations/:id` - Get conversation
- `POST /api/otto/conversations/:id/messages` - Send message
- `GET /api/otto/conversations/:id/messages` - Get messages
- `POST /api/otto/conversations/:id/feedback` - Provide feedback
- And 5+ more for analysis and context

### Subscription (6 endpoints)
- `POST /api/subscription/checkout` - Create checkout session
- `GET /api/subscription/status` - Get subscription status
- `POST /api/subscription/cancel` - Cancel subscription
- `POST /api/subscription/webhooks/paddle` - Paddle webhook handler
- And more for subscription history and invoices

### Health & Monitoring (5 endpoints)
- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed health status
- `GET /health/ready` - Readiness probe (for K8s)
- `GET /health/live` - Liveness probe (for K8s)
- `GET /metrics` - Prometheus metrics (JSON + text format)

---

## Database Schema

### Core Tables
- **users** - User accounts, credentials, subscription info
- **verification_tokens** - Email verification
- **password_reset_tokens** - Password reset
- **user_settings** - User preferences
- **account_security** - Failed login attempts, lockouts
- **auth_audit_logs** - Audit trail

### Workspace & Team
- **workspaces** - Per-user workspace (support for future multi-workspace)
- **workspace_members** - Team members with roles
- **member_invitations** - Email-based invitations

### Instagram Integration
- **instagram_accounts** - Connected accounts with OAuth tokens
- **instagram_media** - Synced media metadata
- **instagram_analytics** - Engagement metrics

### AI & Context
- **conversations** - AI conversation threads
- **messages** - Individual messages in conversations
- **context_documents** - Context docs for AI personalization
- **feedback** - User feedback on AI responses

### Subscription & Payment
- **subscriptions** - Active subscriptions per user
- **subscription_history** - Subscription lifecycle events
- **invoices** - Payment invoices from Paddle

---

## Environment Configuration

All configuration via `.env` file:

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/refyne

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-long-random-secret-key

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Paddle (Sandbox or Live)
PAYMENT_MODE=sandbox
PADDLE_SANDBOX_API_KEY=...
PADDLE_SANDBOX_WEBHOOK_SECRET=...
PADDLE_SANDBOX_PRODUCT_ID_PRO=...

# Instagram OAuth
INSTAGRAM_CLIENT_ID=...
INSTAGRAM_CLIENT_SECRET=...
INSTAGRAM_REDIRECT_URI=...

# Deployment
PORT=8080
APP_ENV=production
AUTO_MIGRATE=true
CORS_ORIGINS=https://app.refyne.io,https://web.refyne.io
```

---

## Deployment & Operations

### Local Development
```bash
# Start services
docker-compose up

# Generate DI wiring
cd cmd && wire

# Run server
make run

# Run tests
make test

# Run E2E tests
APP_ENV=test make run  # Terminal 1
go test ./tests -v     # Terminal 2
```

### Production Deployment (Railway)
```bash
# Environment variables auto-synced
# Database: Railway PostgreSQL
# Cache: Railway Redis
# Email: Gmail SMTP

# Deploy trigger: Push to main branch
git push origin main
# → GitHub Actions CI runs
# → Auto-deploy to Railway
```

### Monitoring
- **Logs:** Railway dashboard + Grafana Cloud
- **Metrics:** Prometheus `/metrics` endpoint
- **Dashboards:** Grafana Cloud (15-panel dashboard)
- **Alerts:** AlertManager with Slack integration
- **On-call:** Alert routing to teams (@database-oncall, #refyne-critical)

---

## Security Posture

✅ **Authentication**
- JWT tokens with automatic refresh
- OTP login (no password stored in email)
- Password hashing with bcrypt

✅ **Data Protection**
- All sensitive data encrypted in transit (HTTPS/TLS)
- Database user credentials in .env (never in code)
- Refresh tokens stored securely

✅ **Access Control**
- Role-based access (Owner, Member)
- Workspace-level isolation
- Audit logging for all actions

✅ **Rate Limiting & DDoS**
- Global rate limit (100 req/min)
- Account lockout (5 failed attempts)
- Automatic cleanup

✅ **Compliance**
- Soft delete for data retention
- Audit logs for 90+ days
- GDPR-compliant account deletion
- Paddle tax compliance

---

## Performance & Scalability

| Metric | Value | Details |
|--------|-------|---------|
| **Response Time (p95)** | < 200ms | Most endpoints < 100ms |
| **Throughput** | 1000+ RPS | Tested on Railway |
| **Database Pool** | 20 connections (10 idle) | Configured for concurrency |
| **Rate Limit** | 100 req/min per user | Protects against abuse |
| **Cache Hit Rate** | 70%+ | Redis-backed rate limiting |
| **Image Size** | 80.8MB | Alpine multi-stage Docker |
| **Memory Usage** | ~100MB | Lean Go runtime |

---

## Roadmap & Future Enhancements

### Phase 2 (Q3 2026)
- [ ] Multi-platform support (TikTok, YouTube, LinkedIn)
- [ ] Advanced analytics dashboard
- [ ] Scheduled posting and automation
- [ ] Content calendar view

### Phase 3 (Q4 2026)
- [ ] AI-powered content recommendations
- [ ] Competitor analysis tools
- [ ] Team collaboration features (comments, reviews)
- [ ] White-label SaaS for agencies

### Phase 4 (2027)
- [ ] Mobile app (iOS/Android)
- [ ] Video editing tools
- [ ] Influencer marketplace
- [ ] Enterprise features (SSO, audit trails)

---

## Support & Documentation

- **Backend Docs:** `/docs` directory (Railway, Testing, Monitoring, Alerting)
- **API Reference:** Postman collection available
- **Database:** Schema documented with migrations
- **Deployment:** Railway-specific setup guide
- **Monitoring:** Grafana dashboard configuration

---

## Contact & Contribution

- **Repository:** [refyne-backend](https://github.com/refyne/refyne-backend)
- **Main Branch:** `main` (production)
- **Issue Tracking:** GitHub Issues
- **Deployment:** Railway (production-ready)

---

**Last Updated:** 2026-04-18  
**Version:** 1.0  
**Status:** ✅ Production Ready
