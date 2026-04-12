# Alerting Guide - Refyne Backend Production Monitoring

Complete guide for setting up, configuring, and managing production alerts using Prometheus and AlertManager.

## Quick Start (Local Development)

### Start Monitoring Stack with Alerts

```bash
# Start with alerting enabled
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d

# Access services:
# - Prometheus: http://localhost:9090
# - AlertManager: http://localhost:9093
# - Grafana: http://localhost:3000 (admin/admin)
```

### Access AlertManager Dashboard

```
http://localhost:9093
```

Shows:
- Active alerts
- Alert groups
- Silences
- Status

## Architecture

```
Backend (8080)
    ↓
/metrics endpoint
    ↓
Prometheus (9090)
    - Evaluates alert rules
    - Stores metrics
    ↓
AlertManager (9093)
    - Groups alerts
    - Routes notifications
    ↓
Notification Channels:
    - Slack
    - Email (future)
    - PagerDuty (future)
```

## Alert Rules

All alert rules are defined in `monitoring/alerting_rules.yml`.

### Alert Levels

#### 🔴 CRITICAL (Severity: critical)
- Immediate action required
- On-call engineer notified
- Sent to `#refyne-critical` Slack channel
- Email to escalation contact
- 10 second group wait, 1 hour re-send

Examples:
- `HighErrorRate`: HTTP 5xx error rate > 5% for 5 minutes
- `DatabaseConnectionPoolExhausted`: DB connections > 90% for 2 minutes
- `ServiceUnresponsive`: Zero requests for 5 minutes
- `FailedLoginSpike`: > 10 failed logins per second

#### ⚠️ WARNING (Severity: warning)
- Investigation recommended
- Sent to `#refyne-warnings` Slack channel
- 1 minute group wait, 4 hour re-send

Examples:
- `HighLatency`: P95 latency > 1 second for 5 minutes
- `HighFailedLoginRate`: > 1 failed login per second for 3 minutes
- `EmailJobFailureRate`: Email failure rate > 10% for 5 minutes

#### ℹ️ INFO (Severity: info)
- Informational only
- Sent to `#refyne-info` Slack channel
- 5 minute group wait, 24 hour re-send

## Setting Up Slack Integration

### Step 1: Create Slack Webhook

1. Go to your Slack workspace
2. Create a new Incoming Webhook:
   - Visit: https://slack.com/apps
   - Search: "Incoming WebHooks"
   - Click "Add to Slack"
   - Choose channel: `#refyne-alerts`
   - Authorize

3. Copy the webhook URL (looks like):
   ```
   https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX
   ```

### Step 2: Configure AlertManager

#### Option A: Environment Variable (Development)

```bash
export SLACK_WEBHOOK_URL='https://hooks.slack.com/services/T00000000/...'
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d
```

#### Option B: .env File

```bash
# Add to .env
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/T00000000/...

# Start containers
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d
```

#### Option C: Direct Edit (Production)

Edit `monitoring/alertmanager/alertmanager.yml`:

```yaml
global:
  slack_api_url: 'https://hooks.slack.com/services/T00000000/...'
```

Reload AlertManager:

```bash
# Send reload signal
curl -X POST http://localhost:9093/-/reload
```

### Step 3: Verify Integration

1. Trigger a test alert:
   ```bash
   # Send test alert directly to AlertManager
   curl -X POST http://localhost:9093/api/v1/alerts \
     -H "Content-Type: application/json" \
     -d '[{
       "labels": {
         "alertname": "TestAlert",
         "severity": "critical"
       },
       "annotations": {
         "summary": "This is a test alert",
         "description": "Slack integration test"
       }
     }]'
   ```

2. Check Slack - you should see the message in `#refyne-alerts`

### Step 4: Configure Alert Channels

Edit `monitoring/alertmanager/alertmanager.yml` to customize channels:

```yaml
receivers:
  - name: 'critical'
    slack_configs:
      - channel: '#refyne-critical'
        # ... customizations
```

## Alert Categories & Response

### Database Alerts

#### DatabaseConnectionPoolExhausted (CRITICAL)
- **Severity:** Critical
- **Threshold:** > 90% connections used for 2 minutes
- **Impact:** New requests will be rejected
- **Response:**
  1. Check for slow queries: `SELECT * FROM pg_stat_activity;`
  2. Increase connection pool (temporary): Update `DB_POOL_MAX` in config
  3. Scale database horizontally (permanent)
  4. Review slow logs for long-running queries

#### DatabaseUnhealthy (CRITICAL)
- **Severity:** Critical
- **Impact:** Service cannot function
- **Response:**
  1. Check database service: `docker logs refyne_db`
  2. Verify connectivity: `psql -U root -h localhost -d refyneDB`
  3. Restart PostgreSQL: `docker restart refyne_db`
  4. Check disk space: `df -h /var/lib/postgresql`

### Authentication/Security Alerts

#### FailedLoginSpike (CRITICAL)
- **Severity:** Critical
- **Threshold:** > 10 failed logins per second
- **Impact:** Possible brute force attack
- **Response:**
  1. Check recent login attempts:
     ```sql
     SELECT * FROM audit_logs 
     WHERE event_type = 'LOGIN_FAILED' 
     ORDER BY created_at DESC LIMIT 100;
     ```
  2. Enable additional rate limiting
  3. Block suspicious IPs
  4. Review for credential compromise

#### HighFailedLoginRate (WARNING)
- **Severity:** Warning
- **Threshold:** > 1 failed login per second for 3 minutes
- **Impact:** Users may be locked out
- **Response:**
  1. Review failed login reasons (invalid credentials, account locked, etc.)
  2. Notify users of potential outages
  3. Check email system (OTP delivery)

### Service Performance Alerts

#### HighErrorRate (CRITICAL)
- **Severity:** Critical
- **Threshold:** 5xx error rate > 5% for 5 minutes
- **Impact:** Service is degraded
- **Response:**
  1. Check logs: `docker logs refyne-backend`
  2. Review error types: `curl http://localhost:9090/metrics | grep http_requests`
  3. Scale application instances
  4. Check external service dependencies

#### HighLatency (WARNING)
- **Severity:** Warning
- **Threshold:** P95 latency > 1 second
- **Impact:** Users experiencing slow responses
- **Response:**
  1. Check database query performance
  2. Review slow endpoint in Grafana
  3. Implement caching for slow endpoints
  4. Scale application

### Subscription/Payment Alerts

#### PaddleAPIErrors (WARNING)
- **Severity:** Warning
- **Threshold:** > 1% API errors for 5 minutes
- **Impact:** Subscription operations may fail
- **Response:**
  1. Check Paddle API status
  2. Verify API credentials in config
  3. Review error logs for specific failures
  4. Contact Paddle support if issue persists

#### PaddleAPIDown (CRITICAL)
- **Severity:** Critical
- **Impact:** Cannot process subscriptions
- **Response:**
  1. Check Paddle service status
  2. Verify network connectivity to Paddle
  3. Test API manually: `curl https://api.paddle.com/v1/health`
  4. Review firewall rules

### Email Service Alerts

#### EmailJobFailureRate (WARNING)
- **Severity:** Warning
- **Threshold:** > 10% jobs failing for 5 minutes
- **Impact:** Users not receiving notifications
- **Response:**
  1. Check email queue: `SELECT COUNT(*) FROM river_job WHERE state = 'failed';`
  2. Review error logs for specific failures
  3. Restart email worker: `docker restart river-email-worker`
  4. Check SMTP configuration
  5. Verify sender email is not rate-limited

## Alert Silencing

### Temporary Silences

Silence an alert while performing maintenance:

1. **Via AlertManager UI:**
   - Go to http://localhost:9093
   - Click "Silences"
   - Click "Create Silence"
   - Select labels (e.g., `alertname=HighLatency`)
   - Set duration (e.g., 30 minutes)
   - Add comment (e.g., "Deploying v2.0")

2. **Via API:**
   ```bash
   curl -X POST http://localhost:9093/api/v1/silences \
     -H "Content-Type: application/json" \
     -d '{
       "matchers": [
         {
           "name": "alertname",
           "value": "HighLatency",
           "isRegex": false
         }
       ],
       "startsAt": "2026-04-12T10:00:00Z",
       "endsAt": "2026-04-12T10:30:00Z",
       "createdBy": "DevOps",
       "comment": "Database maintenance"
     }'
   ```

### Inhibition Rules

Inhibition rules suppress certain alerts when others are firing.

Example in `alertmanager.yml`:
```yaml
inhibit_rules:
  # Don't send warning if critical is firing
  - source_match:
      severity: critical
    target_match:
      severity: warning
    equal: ['alertname', 'component']
```

This prevents alert noise during outages.

## Testing Alerts

### Test Alert Rule Evaluation

```bash
# Check if a rule would fire
curl 'http://localhost:9090/api/v1/query' \
  --data-urlencode 'query=rate(refyne_http_requests_total{status=~"5.."}[5m]) > 0.05'

# View all active alerts
curl http://localhost:9090/api/v1/alerts
```

### Trigger Test Alert

```bash
# Send test alert to AlertManager
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "critical",
      "component": "test"
    },
    "annotations": {
      "summary": "Test alert from API",
      "description": "This is a test"
    },
    "startsAt": "2026-04-12T10:00:00Z",
    "endsAt": "0001-01-01T00:00:00Z"
  }]'
```

### Cause Alert by Load Testing

```bash
# Trigger HighErrorRate alert
ab -n 10000 -c 100 http://localhost:8080/api/health

# Trigger HighLatency alert
for i in {1..1000}; do
  curl -s http://localhost:8080/api/health &
done
```

## Production Deployment

### Configure Slack Webhook (Railway)

1. Get webhook URL from Slack (as above)
2. Add to Railway environment variables:
   ```
   SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
   ```
3. Restart AlertManager

### High Availability Setup

For production, run multiple AlertManager instances:

```bash
# AlertManager 1
docker run -d \
  --name alertmanager-1 \
  -p 9093:9093 \
  -v alertmanager-data-1:/alertmanager \
  -e SLACK_WEBHOOK_URL='https://...' \
  prom/alertmanager

# AlertManager 2
docker run -d \
  --name alertmanager-2 \
  -p 9094:9093 \
  -v alertmanager-data-2:/alertmanager \
  -e SLACK_WEBHOOK_URL='https://...' \
  prom/alertmanager

# Configure Prometheus with both instances
# In prometheus.yml:
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - localhost:9093
            - localhost:9094
```

## Monitoring AlertManager

### AlertManager Metrics

AlertManager exposes metrics at `http://localhost:9093/metrics`:

```bash
# View AlertManager metrics
curl http://localhost:9093/metrics

# Key metrics:
# alertmanager_alerts - Total alerts
# alertmanager_alerts_received_total - Alerts received
# alertmanager_notifications_total - Notifications sent
# alertmanager_notifications_failed_total - Failed notifications
```

### Health Checks

```bash
# Check AlertManager is running
curl http://localhost:9093/-/healthy

# Status code 200 = healthy
```

## Troubleshooting

### Alerts Not Firing

1. **Check Prometheus rule evaluation:**
   ```bash
   curl 'http://localhost:9090/api/v1/rules'
   ```

2. **Verify rule syntax:**
   ```bash
   promtool check rules monitoring/alerting_rules.yml
   ```

3. **Check AlertManager logs:**
   ```bash
   docker logs refyne-alertmanager
   ```

4. **Verify webhook URL:**
   ```bash
   curl -X POST https://hooks.slack.com/services/... \
     -H "Content-Type: application/json" \
     -d '{"text":"Test message"}'
   ```

### Notifications Not Sending

1. **Check AlertManager configuration:**
   ```bash
   docker exec refyne-alertmanager cat /etc/alertmanager/alertmanager.yml
   ```

2. **Verify Slack webhook:**
   ```bash
   curl -X POST $SLACK_WEBHOOK_URL \
     -H "Content-Type: application/json" \
     -d '{"text":"Test from AlertManager"}'
   ```

3. **Check AlertManager logs for delivery errors:**
   ```bash
   docker logs refyne-alertmanager | grep -i "slack\|error"
   ```

4. **Test with verbose logging:**
   ```bash
   docker exec refyne-alertmanager \
     alertmanager --log.level=debug
   ```

### Too Many Alerts

1. **Adjust alert thresholds** in `alerting_rules.yml`
2. **Add inhibition rules** to suppress lower-priority alerts
3. **Increase group_wait** and `repeat_interval` in `alertmanager.yml`

## Runbooks

For each critical alert, create a runbook with troubleshooting steps.

Example runbook structure:
```markdown
# High Error Rate Runbook

## Alert: HighErrorRate

### Symptoms
- Error rate > 5% for 5 minutes
- Slack notification in #refyne-critical

### Diagnosis
1. Check error logs
2. Identify failing endpoint
3. Check recent deployments

### Resolution
1. Scale application instances
2. Rollback bad deployment
3. Enable circuit breaker for failing service
```

## Future Improvements

1. **PagerDuty Integration** - For on-call rotation
2. **Email Notifications** - For secondary channel
3. **Webhook Integration** - For custom scripts
4. **SMS Alerts** - For critical incidents
5. **Incident Auto-Remediation** - Automatic scaling, restarts, etc.

## Useful Commands

```bash
# Reload Prometheus config
curl -X POST http://localhost:9090/-/reload

# Reload AlertManager config
curl -X POST http://localhost:9093/-/reload

# View active alerts
curl http://localhost:9090/api/v1/alerts | jq '.data.alerts | length'

# Query metric
curl 'http://localhost:9090/api/v1/query' \
  --data-urlencode 'query=up'

# View all rules
curl http://localhost:9090/api/v1/rules | jq '.data.groups[0].rules | length'

# Check AlertManager status
curl http://localhost:9093/api/v1/status
```

## References

- [Prometheus Alerting](https://prometheus.io/docs/alerting/latest/overview/)
- [AlertManager Configuration](https://prometheus.io/docs/alerting/latest/configuration/)
- [Slack Incoming Webhooks](https://docs.slack.com/messaging/webhooks)
