# Instagram Integration Documentation

**Date:** April 13, 2026  
**Version:** 1.0  
**Status:** Production Ready  
**Maintainers:** Refyne Backend Team

## 1. Overview

Complete Instagram Graph API integration for Refyne, enabling:
- OAuth 2.0 connection & token management
- Real-time webhook processing (feed, stories, messages)
- Scheduled media & insights synchronization
- AI-powered content analysis (Gemini Vision API)
- Analytics dashboard with trend analysis

## 2. System Architecture

### 2.1 Component Stack

```
Graph API (Instagram)
    ↓
┌─────────────────────────────────┐
│ OAuth Service                    │ Handles token exchange & refresh
│ Media Service                    │ Fetches posts, metrics
│ Insights Service                 │ Account analytics
│ Webhook Service                  │ Event validation & parsing
│ Gemini Service                   │ AI content analysis
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ River Job Queue                  │ Async processing
│ • Webhook Processing             │
│ • Media Sync (30m interval)      │
│ • Insights Fetch (6h interval)   │
│ • AI Analysis (per new media)    │
│ • Token Refresh (55 days)        │
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ Cache Layer (Redis)              │
│ • Media (1h TTL)                 │
│ • Insights (6h TTL)              │
│ • AI Results (24h TTL)           │
│ • Webhook Events (24h)           │
│ • Access Tokens (60d)            │
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ PostgreSQL Database              │
│ • Accounts & connections         │
│ • Media catalog                  │
│ • Insights history               │
│ • AI recommendations             │
└─────────────────────────────────┘
```

### 2.2 Data Flow

**Webhook Real-time Flow:**
```
Instagram Webhook → POST /webhooks
    → Signature Verification
    → Deduplication Check (Redis)
    → River Job Queue
    → Async Processing
    → Database Write
    → 200 OK Response (immediate)
```

**Polling Sync Flow:**
```
River Scheduler (every 30m)
    ↓
SyncMediaArgs Job
    ↓
Rate Limit Check
    ↓
Instagram API Call
    ↓
Dedup Against Existing
    ↓
Database Upsert
    ↓
Queue AI Processing
    ↓
Cache Update
```

## 3. Environment Configuration

### 3.1 Instagram App Setup

**Requirements:**
- Instagram Business Account
- Meta App Developer Account
- Instagram Graph API access

**Environment Variables:**

```bash
# Instagram OAuth
INSTAGRAM_CLIENT_ID=your_app_id
INSTAGRAM_CLIENT_SECRET=your_secret
INSTAGRAM_REDIRECT_URL=https://api.yourapp.com/api/instagram/auth/callback

# Webhook Verification
INSTAGRAM_WEBHOOK_VERIFY_TOKEN=secure_random_token_32_chars_min

# Request Signing (for security)
INSTAGRAM_WEBHOOK_SECRET=webhook_secret_from_app_settings

# Rate Limiting
INSTAGRAM_RATE_LIMIT_PER_HOUR=200  # Instagram's hard limit

# OAuth Token Expiry
INSTAGRAM_TOKEN_REFRESH_DAYS=55    # Refresh at day 55 of 60-day validity
```

### 3.2 Gemini API Setup

```bash
# Google Gemini API
GEMINI_API_KEY=your_gemini_api_key
GEMINI_MODEL=gemini-2.0-flash
GEMINI_MAX_TOKENS=2048
GEMINI_TEMPERATURE=0.7
GEMINI_TOP_P=0.95
GEMINI_TOP_K=64
GEMINI_TIMEOUT_SECONDS=30
```

### 3.3 Redis Configuration

```bash
# For webhook deduplication, token caching, rate limiting
REDIS_URL=redis://localhost:6379/0
REDIS_MAX_RETRIES=3
REDIS_TIMEOUT_SECONDS=5
```

## 4. Deployment Guide

### 4.1 Pre-Production Checklist

- [ ] Instagram app created in Meta Developer Console
- [ ] OAuth redirect URL configured in app settings
- [ ] Webhook callback URL registered
- [ ] Webhook verify token configured
- [ ] Gemini API key obtained and billing enabled
- [ ] Environment variables set in deployment platform
- [ ] Database migrations applied
- [ ] River queue tables created
- [ ] Redis instance provisioned
- [ ] SSL/TLS certificates configured

### 4.2 Railway Deployment

**Update** `railway.env.template`:

```bash
# Instagram Domain
INSTAGRAM_CLIENT_ID=${{secrets.INSTAGRAM_CLIENT_ID}}
INSTAGRAM_CLIENT_SECRET=${{secrets.INSTAGRAM_CLIENT_SECRET}}
INSTAGRAM_REDIRECT_URL=https://api-${RAILWAY_ENVIRONMENT_NAME}.up.railway.app/api/instagram/auth/callback
INSTAGRAM_WEBHOOK_VERIFY_TOKEN=${{secrets.INSTAGRAM_WEBHOOK_VERIFY_TOKEN}}
INSTAGRAM_WEBHOOK_SECRET=${{secrets.INSTAGRAM_WEBHOOK_SECRET}}

# Gemini API
GEMINI_API_KEY=${{secrets.GEMINI_API_KEY}}
GEMINI_MODEL=gemini-2.0-flash

# Service Configuration
RIVER_MAX_WORKERS=12  # For Instagram job queues
```

**Instagram Webhook URL Configuration:**
```
Callback URL: https://api-production.up.railway.app/api/instagram/webhooks
Verify Token: (from INSTAGRAM_WEBHOOK_VERIFY_TOKEN env var)
Subscribed Events: feed, story, messages, message_template_status_update
```

### 4.3 Database Migrations

All required tables are auto-created on startup. Key tables:

```sql
-- Account connections
CREATE TABLE instagram_accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    instagram_user_id VARCHAR(255) UNIQUE NOT NULL,
    access_token TEXT,  -- encrypted
    token_expires_at TIMESTAMP,
    connected_at TIMESTAMP DEFAULT NOW(),
    sync_status VARCHAR(32),  -- syncing/idle/error
    sync_error TEXT,
    last_sync_at TIMESTAMP,
    last_insights_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Media catalog
CREATE TABLE instagram_media (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES instagram_accounts(id),
    instagram_media_id VARCHAR(255) UNIQUE NOT NULL,
    media_type VARCHAR(32),  -- PHOTO/VIDEO/CAROUSEL/REELS
    caption TEXT,
    media_url TEXT,
    posted_at TIMESTAMP,
    like_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    synced_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Analytics data
CREATE TABLE instagram_insights (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES instagram_accounts(id),
    metric_date DATE NOT NULL,
    impressions INT,
    reach INT,
    profile_visits INT,
    follower_count INT,
    engagement_rate DECIMAL(5,2),
    growth_rate DECIMAL(5,2),
    collected_at TIMESTAMP DEFAULT NOW()
);

-- AI recommendations
CREATE TABLE instagram_ai_recommendations (
    id UUID PRIMARY KEY,
    media_id UUID NOT NULL REFERENCES instagram_media(id),
    account_id UUID NOT NULL REFERENCES instagram_accounts(id),
    analysis JSONB,  -- sentiment, themes, quality
    suggestions JSONB,  -- captions, hashtags
    strategy JSONB,  -- posting times, predictions
    confidence_score DECIMAL(3,2),
    generated_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP
);
```

## 5. Operation Guide

### 5.1 Monitoring

**Key Metrics to Track:**

```
API Calls (per account per hour):
├── Threshold: 180/200 (90% utilization warning)
├── Alert: >200 calls/hour (rate limited)
└── Grafana Dashboard: Instagram API Rate Limit

Webhook Processing:
├── Success Rate: >99%
├── Latency: <500ms (p95)
├── Duplicate Events: <0.1%
└── Alert: >1% failure rate

Job Queue Health:
├── Pending Jobs: <1000
├── Failed Jobs: <10 per day
├── Processing Time: <5m (median)
└── Alert: >100 failed jobs in 1h

AI Processing:
├── Confidence Score: >0.8 (high confidence)
├── Latency: <30s per media
├── Cost: Monitor Gemini API usage
└── Alert: >5% of requests fail
```

**Grafana Dashboard Panels:**

Create panels for:
1. Instagram API call rate (by endpoint)
2. Webhook event volume (by type)
3. Sync job success/failure rate
4. AI processing latency distribution
5. Cache hit ratio (media, tokens, insights)
6. Database connection pool utilization
7. River job queue depth
8. Rate limit consumption per account

### 5.2 Troubleshooting

**Webhook Events Not Being Processed:**

```bash
# Check webhook delivery logs
SELECT * FROM river_job WHERE kind = 'instagram_webhook' 
ORDER BY created_at DESC LIMIT 20;

# Verify signature validation is enabled
grep -r "VerifyWebhookSignature" internal/domains/instagram/

# Test webhook manually
curl -X POST http://localhost:8080/api/instagram/webhooks \
  -H "X-Hub-Signature-256: sha256=..." \
  -d '{"object":"instagram","entry":[...]}'
```

**Rate Limiting Issues:**

```bash
# Check current API call count in Redis
redis-cli ZCARD instagram:api_calls:{account_id}

# View rate limit sliding window
redis-cli ZRANGE instagram:api_calls:{account_id} 0 -1 WITHSCORES

# Reset rate limit (emergency only)
redis-cli DEL instagram:api_calls:{account_id}
```

**Media Sync Not Running:**

```bash
# Check River job queue
SELECT * FROM river_job WHERE kind = 'instagram_sync_media' 
ORDER BY created_at DESC LIMIT 10;

# Check last successful sync
SELECT account_id, last_sync_at, sync_status FROM instagram_accounts;

# Manually trigger sync job
# (Query River client or use admin endpoint)
```

**Token Expiration During Processing:**

```bash
# Check token expiry
SELECT account_id, token_expires_at FROM instagram_accounts;

# Force token refresh job
SELECT * FROM river_job WHERE kind = 'instagram_refresh_token'
ORDER BY scheduled_at DESC LIMIT 5;
```

### 5.3 Maintenance Tasks

**Daily:**
- Monitor error logs for Instagram domain
- Check webhook delivery success rate
- Verify media sync jobs completed

**Weekly:**
- Review rate limit utilization trends
- Check Gemini API costs
- Validate database indexes
- Review slow query logs

**Monthly:**
- Analyze media sync latency
- Review cache hit ratios
- Plan token refresh schedule
- Update API quota estimates

**Quarterly:**
- Review Instagram Graph API changelog
- Test OAuth flow end-to-end
- Load test with synthetic media
- Security audit of token storage

## 6. Rate Limiting Strategy

### 6.1 Instagram API Limits

```
Endpoint                    Per Account Limit
────────────────────────────────────────────
GET /me                     10 calls/minute
GET /me/media               5 calls/minute
GET /media/{id}             10 calls/minute
GET /media/{id}/insights    10 calls/minute
POST /media                 1 call/minute
─────────────────────────────────────────
GLOBAL TOTAL                200 calls/hour
```

### 6.2 Our Rate Limiting

```
Strategy: Sliding Window Counter (Redis)
├── Window: 1 hour
├── Target: <180 calls/hour (90% utilization)
├── Backoff: Exponential retry on 429 (Too Many Requests)
└── Circuit Breaker: Pause syncs if >200 calls in rolling hour

Job Prioritization:
├── HIGH: Token refresh (prevent auth failure)
├── MEDIUM: Media sync (30m cadence)
├── MEDIUM: Insights fetch (6h cadence)
└── LOW: AI analysis (parallelizable)
```

## 7. Security Best Practices

### 7.1 Token Management

```
✅ DO:
- Store tokens encrypted in database (AES-256)
- Rotate encryption keys regularly
- Use separate keys per environment
- Set token TTL (60 days for long-lived)
- Implement token refresh before expiry

❌ DON'T:
- Log or expose access tokens
- Store in Redis unencrypted
- Include in error messages
- Use same key across environments
- Rely solely on expiry time
```

### 7.2 Webhook Security

```
✅ DO:
- Verify X-Hub-Signature-256 header on every request
- Validate timestamp (prevent replay attacks)
- Deduplicate by event ID (Redis 24h TTL)
- Return 200 OK before processing (async)
- Log all verification failures

❌ DON'T:
- Accept webhooks without signature verification
- Process stale events (>1 hour old)
- Re-process duplicate event IDs
- Block on webhook processing
- Accept webhooks from unverified IPs
```

### 7.3 API Key Management

```
Gemini API Key:
├── Stored in environment variables only
├── Rotated every 90 days
├── Monitored for unusual usage
├── Rate-limited per key
└── Scoped to minimal permissions

Instagram App Secret:
├── Never exposed to frontend
├── Stored in secure secret manager
├── Used only for server-to-server calls
├── Rotated on security incidents
└── Each environment has separate keys
```

## 8. Scaling Considerations

### 8.1 Current Capacity

```
With 4 Sync Workers + 6 AI Workers:
├── Media sync: 200 media/minute
├── AI analysis: 33 media/minute (Gemini limited)
├── Webhook processing: 1000 events/min
└── Supports: ~50 active accounts
```

### 8.2 Scaling Strategy

**Vertical Scaling (easier first):**
- Increase worker pool size in River
- Increase Redis memory allocation
- Increase database connection pool

**Horizontal Scaling (if needed):**
- Multiple River service instances
- Redis cluster for cache
- Database read replicas for analytics queries
- CDN for media thumbnails

**Database Optimization:**
```sql
-- Add indexes for common queries
CREATE INDEX idx_instagram_media_posted 
ON instagram_media(account_id, posted_at DESC);

CREATE INDEX idx_instagram_insights_date 
ON instagram_insights(account_id, metric_date DESC);

-- Partition large tables by date
ALTER TABLE instagram_media 
PARTITION BY RANGE (YEAR(posted_at));
```

## 9. Rollback Procedures

### 9.1 Database Rollback

```bash
# If migration fails, previous version available
# Schema is backward compatible for non-destructive changes

# Checklist:
- [ ] Verify no active queries on instagram_* tables
- [ ] Create backup before rollback
- [ ] Disable Instagram service temporarily
- [ ] Run rollback migration
- [ ] Verify all tests pass
- [ ] Re-enable service
```

### 9.2 Code Rollback

```bash
# Revert to previous version
git revert <commit_hash>
git push

# Or use Railway deployment history
railway rollback <deployment_id>

# Verify:
- [ ] Webhook still queues correctly
- [ ] Sync jobs run without errors
- [ ] Access tokens still valid
```

## 10. Compliance & Compliance

### 10.1 Data Retention

```
Instagram Data Minimization:
├── Media URLs: Keep 7 days (cache busting)
├── Engagement Metrics: Keep 90 days (analytics)
├── Account Connections: Keep indefinitely (user data)
├── Webhook Events: Keep 24 hours (dedup only)
└── AI Analysis: Keep 30 days (recommendations)
```

### 10.2 GDPR Compliance

```
User Rights:
├── Access: Export connected accounts data
├── Deletion: Remove all account data on disconnect
├── Portability: Export media metadata
└── Objection: Disable tracking via settings

Consent:
├── Explicit consent for OAuth connection
├── Clear privacy policy on data usage
├── Transparent about Gemini API processing
└── Option to opt-out of AI analysis
```

## 11. Quick Reference

### 11.1 Common Commands

```bash
# View recent Instagram job activity
psql $DATABASE_URL -c "
SELECT kind, state, attempt, created_at 
FROM river_job 
WHERE kind LIKE 'instagram_%' 
ORDER BY created_at DESC LIMIT 20;"

# Check webhook verification token
echo $INSTAGRAM_WEBHOOK_VERIFY_TOKEN

# Test media sync trigger
curl -X POST http://localhost:8080/api/instagram/media/sync \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"account_id": "acc_123"}'

# View rate limit status
redis-cli ZCARD instagram:api_calls:{account_id}

# Clear stuck jobs in River
psql $DATABASE_URL -c "
DELETE FROM river_job 
WHERE kind = 'instagram_sync_media' 
AND state = 'errored' AND attempt > 5;"
```

### 11.2 Emergency Procedures

**Webhook Delivery Stopped:**
1. Verify webhook URL is correct in Instagram app settings
2. Check signature verification keys match
3. Restart River job processor: `systemctl restart refyne`
4. Monitor logs: `tail -f /var/log/refyne/instagram.log`
5. If persistent, manually re-subscribe webhook

**Rate Limited by Instagram:**
1. Check current call count: `redis-cli ZCARD instagram:api_calls:*`
2. Pause media sync jobs temporarily
3. Only allow critical token refresh jobs
4. Wait 1 hour for rate limit window to reset
5. Resume normal operations

**Database Connection Lost:**
1. Verify database is accessible
2. Check connection pool exhaustion: `SHOW max_connections`
3. Restart application server gracefully
4. Wait for existing connections to drain
5. Resume service

---

## Document Control

- **Version:** 1.0
- **Last Updated:** April 13, 2026
- **Next Review:** July 13, 2026
- **Contact:** Backend Team @refyne-dev on Slack
