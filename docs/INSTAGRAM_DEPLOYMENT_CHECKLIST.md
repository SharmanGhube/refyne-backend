# Production Deployment Checklist - Instagram Integration

**Release Date:** April 13, 2026  
**Integration Status:** Complete & Ready for Production  
**Checklist Version:** 1.0

## Pre-Production Requirements

### Instagram App Configuration ✓
- [ ] Instagram Business Account created
- [ ] Meta Developer Account setup
- [ ] Instagram Graph API access approved
- [ ] App ID obtained: `${INSTAGRAM_CLIENT_ID}`
- [ ] App Secret obtained: `${INSTAGRAM_CLIENT_SECRET}`
- [ ] Webhook URL registered: `https://api.yourapp.com/api/instagram/webhooks`
- [ ] Webhook verify token generated (32+ chars)
- [ ] Webhook subscribed to: feed, story, messages
- [ ] OAuth redirect URL configured: `https://api.yourapp.com/api/instagram/auth/callback`
- [ ] App roles assigned (page access, insights)

### Gemini API Configuration ✓
- [ ] Google Cloud project created
- [ ] Generative AI API enabled
- [ ] Service account created
- [ ] API key generated: `${GEMINI_API_KEY}`
- [ ] Billing enabled and quota set
- [ ] Rate limits configured (60000 tokens/min)
- [ ] Mock testing completed

### Infrastructure & Services ✓

**Database (PostgreSQL):**
- [ ] PostgreSQL 14+ deployed
- [ ] Backup strategy implemented (daily automated)
- [ ] Replication configured (if HA required)
- [ ] Connection pooling enabled (PgBouncer/pgx)
- [ ] Slow query logging enabled
- [ ] Max connections: ≥50
- [ ] SSL/TLS for all connections

**Redis Cache:**
- [ ] Redis 7+ deployed
- [ ] Persistence enabled (RDB or AOF)
- [ ] Backup strategy implemented
- [ ] Memory allocation: ≥4GB (minimum)
- [ ] Eviction policy: allkeys-lru
- [ ] SSL/TLS for client connections
- [ ] ACL configured with passwords

**Job Queue (River):**
- [ ] Tables created: `river_job`, `river_queue`
- [ ] Workers assigned: 4 sync + 6 AI = 10 total
- [ ] Queue priorities configured
- [ ] Job retention policy set (14 days)
- [ ] Dead letter queue monitoring enabled

### Environment Variables ✓

**Required (will block startup if missing):**

```bash
# Instagram OAuth
✓ INSTAGRAM_CLIENT_ID=<your_app_id>
✓ INSTAGRAM_CLIENT_SECRET=<your_secret>
✓ INSTAGRAM_REDIRECT_URL=https://api.yourapp.com/api/instagram/auth/callback
✓ INSTAGRAM_WEBHOOK_VERIFY_TOKEN=<secure_token_32+_chars>
✓ INSTAGRAM_WEBHOOK_SECRET=<webhook_signing_secret>

# Gemini API
✓ GEMINI_API_KEY=<google_api_key>

# Database
✓ DATABASE_URL=postgres://user:pass@host:5432/refyne?sslmode=require

# Redis
✓ REDIS_URL=redis://user:pass@host:6379/0?tls=true
```

**Optional (with sensible defaults):**

```bash
# Instagram Configuration
INSTAGRAM_RATE_LIMIT_PER_HOUR=200        # Default: 200
INSTAGRAM_TOKEN_REFRESH_DAYS=55           # Default: 55

# Gemini Configuration
GEMINI_MODEL=gemini-2.0-flash             # Default: gemini-2.0-flash
GEMINI_MAX_TOKENS=2048                    # Default: 2048
GEMINI_TEMPERATURE=0.7                    # Default: 0.7
GEMINI_TOP_P=0.95                         # Default: 0.95
GEMINI_TOP_K=64                           # Default: 64
GEMINI_TIMEOUT_SECONDS=30                 # Default: 30

# River Job Queue
RIVER_MAX_WORKERS=10                      # Default: 10
RIVER_JOB_RETENTION_DAYS=14               # Default: 14
RIVER_POLL_INTERVAL_MS=100                # Default: 100

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100        # Default: 100
```

### Security & Encryption ✓
- [ ] Database passwords rotated and strong (32+ chars)
- [ ] Redis passwords configured and strong
- [ ] API keys stored in secure secret manager (not version control)
- [ ] Encryption keys for token storage generated
- [ ] SSL/TLS certificates valid (not self-signed)
- [ ] CORS headers configured correctly
- [ ] JWT secret configured (55+ chars)
- [ ] Rate limiting enabled on all endpoints
- [ ] Request signing for webhooks verified

### Testing & Validation ✓

**Unit Tests:**
- [ ] All services tested: `./internal/domains/instagram/services/*`
- [ ] All handlers tested: `./internal/domains/instagram/handlers/*`
- [ ] All jobs tested: `./internal/domains/instagram/jobs/*`

**Integration Tests:**
- [ ] E2E tests passing: `go test ./internal/domains/instagram/tests -v`
- [ ] All 21 tests passing
- [ ] Build succeeds: `go build ./...`
- [ ] No lint errors: `golangci-lint run ./internal/domains/instagram/...`

**Instagram Sandbox Testing:**
- [ ] OAuth flow tested end-to-end
- [ ] Token refresh working
- [ ] Webhook subscription verified
- [ ] Sample webhook event processed successfully
- [ ] Media sync completed without errors
- [ ] Insights fetched and stored
- [ ] AI analysis generated proper recommendations
- [ ] Rate limiting tested and working
- [ ] Error handling validated

**Load Testing:**
- [ ] 50 concurrent media syncs complete without errors
- [ ] 100 concurrent webhook events processed
- [ ] Database connection pool adequate
- [ ] Redis didn't hit memory limits
- [ ] Job queue processed without bottleneck

### Documentation & Runbooks ✓
- [ ] Deployment guide complete: `docs/INSTAGRAM_INTEGRATION.md`
- [ ] Troubleshooting guide created
- [ ] Emergency procedures documented
- [ ] Team trained on monitoring
- [ ] On-call procedures established
- [ ] Escalation contacts documented

### Monitoring & Alerting ✓
- [ ] Prometheus metrics exported at `/metrics`
- [ ] Grafana dashboard created with key panels:
  - [ ] Instagram API rate limit utilization
  - [ ] Webhook processing success rate
  - [ ] Media sync job duration
  - [ ] AI analysis confidence scores
  - [ ] Error rates by endpoint
  - [ ] Database connection pool usage
  - [ ] Redis memory utilization
  - [ ] River job queue depth

- [ ] Alerts configured:
  - [ ] Rate limit > 90% utilization
  - [ ] Webhook failure rate > 1%
  - [ ] Job queue depth > 1000
  - [ ] Database connections exhausted
  - [ ] Gemini API errors detected
  - [ ] Token expiry approaching (< 7 days)

- [ ] On-call rotation set up
- [ ] Escalation path defined
- [ ] Slack notifications configured

### Backup & Disaster Recovery ✓
- [ ] Database backups automated (daily)
- [ ] Redis backups automated (daily)
- [ ] Backup retention: 30 days
- [ ] Restore procedure tested
- [ ] RTO (Recovery Time Objective): < 4 hours
- [ ] RPO (Recovery Point Objective): 24 hours
- [ ] Disaster recovery drill scheduled

### Performance Baseline ✓
Document before going live for comparison:
```
Metric                          Target          Expected
─────────────────────────────────────────────────────────
OAuth callback latency          < 500ms         300ms
Webhook processing latency      < 500ms         400ms
Media sync (100 items)          < 5s            3s
Insights fetch latency          < 2s            1.2s
AI analysis per media           < 30s           20s
Cache hit ratio (media)         > 95%           97%
Database query p95              < 100ms         50ms
Error rate (all endpoints)      < 0.1%          0.02%
```

### Deployment Steps

**1. Pre-Deployment (1 day before):**
- [ ] Notify team of scheduled deployment
- [ ] Create database backup
- [ ] Create Redis backup
- [ ] Verify rollback procedure
- [ ] Schedule on-call coverage

**2. Deployment (during maintenance window):**
- [ ] Pull latest code: `git pull origin main`
- [ ] Build Docker image: `docker build -t refyne-backend:latest .`
- [ ] Push to registry: `docker push <registry>/refyne-backend:latest`
- [ ] Update deployment: `kubectl set image deployment/refyne-backend <image>`
- [ ] Monitor pod startup: `kubectl logs -f deployment/refyne-backend`

OR for Railway:

- [ ] Merge PR to main branch
- [ ] Railway auto-deploys from main
- [ ] Monitor deployment in Railway dashboard
- [ ] Verify health check passing

**3. Post-Deployment (30 mins after):**
- [ ] Check error logs for exceptions
- [ ] Verify webhook is receiving events
- [ ] Trigger manual sync: `curl -X POST .../api/instagram/media/sync ...`
- [ ] Check metrics in Grafana
- [ ] Verify all service endpoints responding
- [ ] Confirm backup jobs running
- [ ] Test OAuth flow with sandbox account

**4. Validation (1 hour after):**
- [ ] All metrics within baseline ranges
- [ ] No error rate spikes
- [ ] Queue processing normally
- [ ] Database performance acceptable
- [ ] Team confirms all systems healthy

**5. Communication:**
- [ ] Post deployment summary to #refyne-deploys
- [ ] Update status page
- [ ] Send team notification

### Rollback Procedures

**If Critical Issues Found:**

```bash
# Option 1: Revert Code & Redeploy
git revert HEAD
git push origin main
# Railway/K8s auto-deploys previous version

# Option 2: Use Railway Rollback
railway rollback <previous_deployment_id>

# Verification:
- [ ] Service is responding
- [ ] Error rate returned to baseline
- [ ] Webhooks processing normally
- [ ] Team confirms system stable
```

### Go-Live Approval

**Sign-off Required From:**

- [ ] Backend Lead: Backend infrastructure validated
- [ ] DevOps Engineer: Deployment proven & backed up
- [ ] Product Manager: Feature ready for users
- [ ] Security Lead: All security checks passed

**Sign-off Signatures:**

```
Backend Lead: __________________ Date: __________

DevOps Engineer: ________________ Date: __________

Product Manager: ________________ Date: __________

Security Lead: __________________ Date: __________
```

### Post-Launch Support (First Week)

**Daily Monitoring:**
- [ ] Check error logs first thing each morning
- [ ] Review error rate trends
- [ ] Verify webhook delivery healthy
- [ ] Check API rate limit usage
- [ ] Monitor database performance

**Weekly Review:**
- [ ] Analyze performance metrics
- [ ] Review failed job attempts
- [ ] Check token refresh completion
- [ ] Validate AI confidence scores
- [ ] Update runbooks if needed

## Contact & Escalation

**On-Call Engineer:** [Name/Contact]  
**Backend Lead:** [Name/Contact]  
**DevOps Lead:** [Name/Contact]  

**Communication Channels:**
- #refyne-critical (real-time alerts)
- #refyne-deploys (deployment notifications)
- @backend-oncall (urgent issues)

---

## Completed Milestones

✅ Phase 11: Service Layer Implementation  
✅ Phase 12: Job Worker Implementation  
✅ Phase 13: Real API Integration (Instagram + Gemini)  
✅ Phase 14: Handler Dependency Injection  
✅ Phase 15: Route Registration & Integration  
✅ Phase 16: Production Documentation & Deployment Plan  

**Status:** Ready for Production Deployment

**Last Updated:** April 13, 2026  
**Next Review:** Before production launch
