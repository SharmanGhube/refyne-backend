# Instagram Integration - Troubleshooting Guide

**Last Updated:** April 13, 2026  
**Severity Levels:** Critical 🔴 | Major 🟠 | Minor 🟡 | Info ⚪

## Quick Diagnostics

### Is the service running?

```bash
# Check if backend is accepting connections
curl -i http://localhost:8080/health

# Expected: 200 OK with { "status": "ok" }
```

### Are database connections healthy?

```bash
# Check PostgreSQL connection pool
psql $DATABASE_URL -c "SELECT datname, usename, count(*) FROM pg_stat_activity GROUP BY datname, usename;"

# Expected: Connection count < 50 (most configurations)
```

### Is Redis accessible?

```bash
# Test Redis connectivity
redis-cli PING

# Expected: PONG
```

### Are River jobs being processed?

```bash
# Check River job queue status
psql $DATABASE_URL -c "
SELECT state, COUNT(*) as count FROM river_job 
WHERE kind LIKE 'instagram_%' 
GROUP BY state;"

# Expected: Most jobs should have state='completed'
```

---

## Critical Issues 🔴

### Issue: Webhooks Not Received from Instagram

**Symptoms:**
- No webhook events appearing in logs
- River queue has no `instagram_webhook` jobs
- Instagram app status shows "Many warnings"

**Diagnosis:**

```bash
# 1. Verify webhook URL is accessible
curl -v https://api.yourapp.com/api/instagram/webhooks

# Should NOT return 404 or connection refused
# Should return 403 (no signature on test)

# 2. Check webhook logs for recent activity
tail -100f /var/log/refyne/instagram.log | grep webhook

# 3. Verify webhook is registered in Instagram app
# Go to Meta Dashboard > Settings > Webhooks
# Should show: Callback URL: https://api.yourapp.com/api/instagram/webhooks
```

**Solutions:**

1. **Wrong Callback URL:**
   - Go to Meta Developer Dashboard
   - Settings > Basic > Copy your App ID
   - Webhooks section: Update Callback URL
   - Verify: `https://api.yourapp.com/api/instagram/webhooks` (no trailing slash)

2. **Verify Token Mismatch:**
   ```bash
   # Check configured token
   echo $INSTAGRAM_WEBHOOK_VERIFY_TOKEN
   
   # Go to Meta Dashboard > Webhooks
   # Click "Verify and Save"
   # Paste the token shown there into INSTAGRAM_WEBHOOK_VERIFY_TOKEN env var
   ```

3. **Network Connectivity Issue:**
   ```bash
   # Test if Instagram can reach your server
   # From your server, test connectivity to Instagram
   curl -I https://graph.instagram.com/
   
   # If behind firewall, whitelist Instagram IP ranges:
   # See: https://developers.facebook.com/docs/graph-api/webhooks
   ```

4. **SSL Certificate Problem:**
   ```bash
   # Verify SSL certificate is valid
   openssl s_client -connect api.yourapp.com:443
   
   # Certificate must:
   # - Not be self-signed
   # - Have valid expiry date
   # - Match domain name
   ```

**Resolution Steps:**
1. ✓ Fix callback URL in Instagram app settings
2. ✓ Test with `curl` to confirm reachability
3. ✓ Wait 5 minutes for Instagram to re-enable webhooks
4. ✓ Check logs for webhook delivery
5. ✓ Monitor River job queue

---

### Issue: Rate Limited by Instagram (429 Errors) 🔴

**Symptoms:**
```
ERROR: Failed to fetch media from Instagram
Response Status: 429
Message: Rate limit reached
```

**Diagnosis:**

```bash
# Check current API call count in Redis
redis-cli ZCARD instagram:api_calls:{account_id}

# View sliding window of calls
redis-cli ZRANGE instagram:api_calls:{account_id} 0 -1 WITHSCORES | tail -10

# Check how many calls used in last hour
redis-cli ZCOUNT instagram:api_calls:{account_id} $(date -d '1 hour ago' +%s) +inf
```

**Root Causes:**

1. **Too Many Sync Jobs Running Simultaneously**
   ```
   Cause: Multiple accounts syncing at same time
   Limit: 200 calls/hour per account
   ```

2. **Gemini API Also Making Calls**
   ```
   Each media analysis might trigger additional API calls
   ```

3. **Retry Loop Without Backoff**
   ```
   Failed requests retried immediately, consuming quota
   ```

**Solutions:**

**Immediate (Stop the Bleeding):**

```bash
# Pause media sync jobs temporarily
psql $DATABASE_URL -c "
UPDATE river_job SET state='discarded' 
WHERE kind='instagram_sync_media' AND state='scheduled';"

# Monitor: Check if error rate decreases
tail -f /var/log/refyne/instagram.log | grep "Rate limit"

# Wait 1 hour for rate limit window to slide
sleep 3600

# Resume one account at a time
# Manually restart sync via dashboard
```

**Prevent Recurrence:**

1. **Reduce concurrent syncs:**
   ```bash
   # Edit config: Max 1 sync per account per 30 minutes
   INSTAGRAM_SYNC_INTERVAL_MINUTES=30
   
   # Stagger accounts: Account A at 0min, B at 10min, C at 20min
   ```

2. **Reduce calls per sync:**
   ```bash
   # Fetch only new media, not full history
   # Use insights fields that don't require extra calls
   ```

3. **Implement circuit breaker:**
   ```go
   // In rate_limiter.go - stop calling if approaching limit
   if callsInWindow > 180 {
       // Return early, don't make API call
       return nil, errRateLimited
   }
   ```

4. **Exponential backoff on 429:**
   ```
   1st attempt: Immediate
   2nd attempt: Retry after 5s
   3rd attempt: Retry after 10s
   4th attempt: Retry after 20s
   etc.
   ```

**Long-term Monitoring:**

```bash
# Create alert when > 150 calls used
# Create dashboard panel showing call usage trend
# Set budget: Max 150 calls per account per day
```

---

### Issue: Webhook Verification Fails (Invalid Signature) 🔴

**Symptoms:**
```
WARNING: Invalid webhook signature
Response: 401 Unauthorized
Logs show repeated signature validation failures
```

**Diagnosis:**

```bash
# 1. Check webhook secret used for signing
echo $INSTAGRAM_WEBHOOK_SECRET

# 2. Check what Instagram is sending
# (Log the signature header)

# 3. Verify our calculation logic
grep -A 10 "VerifyWebhookSignature" internal/domains/instagram/services/webhook_service.go
```

**Root Causes:**

1. **Wrong Webhook Secret:**
   - Webhook signing secret ≠ webhook verify token
   - Signing secret from app Dashboard > Settings > Basic

2. **Timing Issues:**
   - Webhook timestamp > 5 minutes old
   - Server clock skew

3. **Body Modification:**
   - Body processed/decoded before signature check
   - Whitespace changes in JSON

**Solutions:**

```bash
# 1. Get correct secret from Meta Dashboard
# Settings > Basic > Webhook Secret (NOT Verify Token)
# Update: INSTAGRAM_WEBHOOK_SECRET=<new_secret>

# 2. Verify time sync
ntpdate -q pool.ntp.org
hwclock --set --date="$(date)"

# 3. Restart service with new secret
systemctl restart refyne

# 4. Test webhook from Instagram
# Meta Dashboard > Webhooks > Send Test Event
```

---

## Major Issues 🟠

### Issue: Media Sync Jobs Failing

**Symptoms:**
- Sync jobs show 'errored' state in River
- Media not updating in past 6+ hours
- User sees stale media in dashboard

**Diagnosis:**

```bash
# View failed sync jobs
psql $DATABASE_URL -c "
SELECT id, kind, state, attempt, last_error, created_at 
FROM river_job 
WHERE kind='instagram_sync_media' AND state='errored'
ORDER BY created_at DESC LIMIT 10;"

# Check error details
psql $DATABASE_URL -c "
SELECT last_error FROM river_job 
WHERE id='<job_id>';"
```

**Common Error Messages & Solutions:**

**Error: "Failed to retrieve account access token"**
```
Cause: Token no longer in database
Solution:
1. Verify account exists: SELECT * FROM instagram_accounts WHERE id='<id>';
2. If missing, account was deleted
3. Re-connect account via OAuth
```

**Error: "No access token for account"**
```
Cause: Account connected but no token stored
Solution:
1. Check token expiry: SELECT token_expires_at FROM instagram_accounts WHERE id='<id>';
2. If expired > 60 days, token invalid
3. Re-authenticate account
```

**Error: "Database sync failed"**
```
Cause: Upsert query error
Solution:
1. Check database logs: SELECT * FROM pg_stat_statements WHERE query LIKE '%instagram_media%' ORDER BY total_time DESC;
2. Check for constraint violations
3. Verify media schema is correct
```

**Resolution:**

```bash
# Check failed job details
psql $DATABASE_URL -c "
SELECT id, kind, args, last_error 
FROM river_job 
WHERE kind='instagram_sync_media' 
AND state='errored' 
LIMIT 1;"

# Identify root cause from error message
# Fix underlying issue (see above)
# Manually retry job (optional)
```

### Issue: AI Analysis Not Generating Recommendations 🟠

**Symptoms:**
- AI analysis jobs queued but not completing
- No recommendations appear in MediaRecommendations endpoint
- Jobs in 'errored' state with Gemini API errors

**Diagnosis:**

```bash
# Check AI job status
psql $DATABASE_URL -c "
SELECT state, COUNT(*) as count FROM river_job 
WHERE kind='instagram_process_ai' 
GROUP BY state;"

# View recent errors
tail -50 /var/log/refyne/instagram.log | grep "gemini\|ai_process"
```

**Common Issues:**

**Error: Invalid Gemini API Key**
```
Cause: Expired or incorrect GEMINI_API_KEY
Solution:
1. Log into Google Cloud console
2. Verify API key exists and not rate limited
3. Confirm billing enabled
4. Update .env: GEMINI_API_KEY=<new_key>
5. Restart service
```

**Error: Gemini Rate Limited (429)**
```
Cause: Exceeded 60000 tokens/minute
Solution:
1. Check Gemini usage: Google Cloud console > Vision API > Quotas
2. Reduce AI processing jobs (batch them)
3. Implement queue backoff
```

**Error: Image Download Failed (Image URL Invalid)**
```
Cause: Instagram media URL expired (24h TTL)
Solution:
1. Store image locally when fetched
2. Or refresh media URL from Instagram before analyzing
```

**Solutions:**

```bash
# Validate Gemini API key
curl -X POST https://generativelanguage.googleapis.com/v1/models/gemini-2.0-flash:generateContent \
  -H "Content-Type: application/json" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -d '{"contents": [{"parts": [{"text": "Hello"}]}]}'

# Should return a response, not 401

# If fails: Update key and restart
```

---

## Minor Issues 🟡

### Issue: Token Expiration Refresh Not Working

**Symptoms:**
- Token expires without refresh
- Accounts require re-authentication after 60 days
- No `instagram_refresh_token` jobs appearing

**Diagnosis:**

```bash
# Check token expiry dates
psql $DATABASE_URL -c "
SELECT id, token_expires_at, NOW() 
FROM instagram_accounts 
WHERE token_expires_at < NOW() + INTERVAL 7 days
ORDER BY token_expires_at;"

# Check for refresh jobs
psql $DATABASE_URL -c "
SELECT * FROM river_job 
WHERE kind='instagram_refresh_token' 
ORDER BY created_at DESC LIMIT 10;"
```

**Solutions:**

```bash
# Manually trigger refresh for account
# (Implement admin endpoint or use direct job insertion)

INSERT INTO river_job (kind, state, args, created_at, scheduled_at)
VALUES ('instagram_refresh_token', 'available', 
  '{"account_id":"<id>"}', NOW(), NOW());

# Verify refresh job completes
SELECT * FROM river_job WHERE kind='instagram_refresh_token' ORDER BY created_at DESC LIMIT 1;
```

### Issue: High Cache Miss Rate

**Symptoms:**
- Media endpoint slow (not using cache)
- Redis query latency high
- Too many Instagram API calls

**Diagnosis:**

```bash
# Check cache hit ratio
redis-cli INFO stats | grep -E "hits|misses"

# Sample: hits=1000, misses=1000 = 50% hit rate (needs improvement)

# Check what's in cache
redis-cli --pattern "instagram:media:*" | wc -l

# Should be > 20 if 20+ accounts
```

**Root Causes:**

1. **Cache TTL Too Short:** 1h might be too fast expiry
2. **Cache Key Mismatch:** Query using wrong format
3. **Cache Bypass:** Calls not caching results

**Solutions:**

```bash
# Verify cache is configured
grep -r "CacheMedia\|SetEx" internal/domains/instagram/services/

# Increase TTL for media cache
# Change: time.Hour → 6*time.Hour (if acceptable)

# Monitor hit rate improvement
redis-cli INFO stats | grep -E "hits|misses"
```

---

### Issue: Slow Analytics Queries

**Symptoms:**
- `/api/instagram/analytics` endpoint slow (> 1s)
- Database CPU spike when fetching analytics
- High query latency in logs

**Diagnosis:**

```bash
# Check query performance
psql $DATABASE_URL -c "
SELECT query, mean_exec_time, calls 
FROM pg_stat_statements 
WHERE query LIKE '%instagram_insights%' 
ORDER BY mean_exec_time DESC LIMIT 5;"

# View slow query log
grep "duration:" /var/log/postgresql/postgresql.log | tail -20
```

**Solutions:**

```bash
# Add index on common query pattern
CREATE INDEX idx_insights_account_date 
ON instagram_insights(account_id, metric_date DESC);

# Reanalyze query plan
EXPLAIN ANALYZE 
SELECT * FROM instagram_insights 
WHERE account_id='<id>' AND metric_date > NOW() - INTERVAL 30 days;

# Consider materialized view for frequently aggregated data
CREATE MATERIALIZED VIEW instagram_insights_daily AS
SELECT account_id, DATE(metric_date) as date, 
  AVG(impressions), AVG(reach), AVG(follower_count)
FROM instagram_insights
GROUP BY account_id, DATE(metric_date);
```

---

## Info & Best Practices ⚪

### Webhook Event Types Explained

**feed (Most Common)**
```json
{
  "value": {
    "media_id": "17123456789",
    "status": "PUBLISHED"  // or EXPIRED
  }
}
```
Action: Fetch this media + analytics

**story**
```json
{
  "value": {
    "story_id": "...",
    "status": "PUBLISHED"  // 24h expiry
  }
}
```
Action: Optional - stories expire quickly

**messages**
```json
{
  "value": {
    "from": "123456",
    "to": "987654",
    "message_id": "m_123",
    "text": "Hello!"
  }
}
```
Action: Store message, optional auto-response

---

### Performance Tuning Tips

**Reduce API Call Latency:**

```bash
# 1. Use connection pooling (already done in code)
# 2. Set timeouts appropriately
MEDIA_FETCH_TIMEOUT_SECONDS=15
INSIGHTS_FETCH_TIMEOUT_SECONDS=10

# 3. Parallelize independent requests
# (Already done in batch media sync)
```

**Improve Webhook Processing:**

```bash
# 1. Increase River workers for webhooks
RIVER_WEBHOOK_WORKERS=10  # from default 2

# 2. Use goroutine pool for deduplication checks
# (Reduce Redis round trips)

# 3. Async notification delivery
# (Already implemented)
```

**Database Optimization:**

```bash
# 1. Enable query parallelization
SET max_parallel_workers_per_gather = 4;

# 2. Increase shared buffers
shared_buffers = 4GB  # 25% of system RAM

# 3. Tune cost settings
random_page_cost = 1.1  # For SSD (not HDD)
```

---

### Monitoring Checklist

**Daily:**
- [ ] Error rate < 0.1%
- [ ] Webhook delivery success > 99%
- [ ] No stuck jobs in River queue
- [ ] Token refresh completed for expiring tokens

**Weekly:**
- [ ] API call usage < 180/hour per account
- [ ] Cache hit ratio > 95%
- [ ] Database indexes healthy (no bloat)
- [ ] Gemini API usage within budget

**Monthly:**
- [ ] Review slow query log
- [ ] Analyze error patterns
- [ ] Test disaster recovery
- [ ] Validate backup integrity

---

## Getting Help

**Before Contacting Support:**

1. ✓ Reproduce issue with steps
2. ✓ Check logs: `tail -100 /var/log/refyne/instagram.log`
3. ✓ Run diagnostics from "Quick Diagnostics" section
4. ✓ Check this guide for similar issues
5. ✓ Check known limitations below

**Report With:**

- [ ] Error message verbatim
- [ ] Steps to reproduce
- [ ] Relevant logs (anonymized)
- [ ] Environment (staging/prod, service version)
- [ ] When it started

**Contact:**

- Slack: #refyne-help
- Email: backend-team@refyne.dev
- On-Call: @backend-oncall

---

## Known Limitations

1. **Instagram API:**
   - 200 API calls/hour per account (hard limit)
   - 60-second webhook event delivery window
   - Media URLs expire after 24 hours

2. **Gemini API:**
   - 60,000 tokens/minute (can request increase)
   - 4096 tokens per request maximum
   - Vision API supports: PNG, JPEG, WebP, HEIC

3. **Rate Limiting:**
   - In-memory in single instance (use Redis for multi-instance)
   - Sliding window algorithm has edge cases around hour boundaries

4. **Token Storage:**
   - Tokens encrypted but stored in database (key compromise = keys compromised)
   - No automatic rotation (manual intervention required)

---

**Document Version:** 1.0  
**Last Updated:** April 13, 2026  
**Feedback:** Please update guide when resolving issues not covered above
