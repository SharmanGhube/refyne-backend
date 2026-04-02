# Monitoring Guide - Refyne Backend

Production observability stack with Prometheus metrics collection and Grafana visualization.

## Quick Start

### Start Monitoring Stack

```bash
# Start all services (backend + monitoring)
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d

# Or start monitoring alongside running backend
docker-compose -f docker-compose-monitoring.yml up -d
```

### Access Dashboards

- **Grafana:** http://localhost:3000
  - Username: `admin`
  - Password: `admin`
  - Dashboard: "Refyne Backend - Production Metrics"

- **Prometheus:** http://localhost:9090
  - Query interface at `/graph`
  - Targets at `/targets`

## Architecture

```
Backend (8080)
    ↓
    /metrics endpoint
         ↓
    Prometheus (9090)
    - Scrapes every 10s
    - Stores metrics (15 days)
         ↓
    Grafana (3000)
    - Visualizes data
    - Real-time dashboards
```

## Metrics Collected

### HTTP Requests
- `refyne_http_requests_total` - Counter: total requests by method/endpoint/status
- `refyne_http_request_duration_seconds` - Histogram: request duration (P50/P95/P99)
- `refyne_rate_limit_exceeded_total` - Counter: rate limit violations

### Database
- `refyne_db_connections_active` - Gauge: active connections in pool
- `refyne_db_connections_used` - Gauge: connections currently in use

### Redis
- `refyne_redis_operations_total` - Counter: Redis operations (get/set/etc)
- `refyne_redis_errors_total` - Counter: Redis errors by type

### Authentication
- `refyne_auth_login_attempts_total` - Counter: login attempts by method
- `refyne_auth_login_failures_total` - Counter: failed logins by reason
- `refyne_auth_tokens_generated_total` - Counter: tokens created (access/refresh)

### Subscriptions (Paddle)
- `refyne_paddle_api_calls_total` - Counter: Paddle API calls by operation
- `refyne_paddle_api_errors_total` - Counter: Paddle errors
- `refyne_subscriptions_by_tier` - Gauge: active subscriptions by tier/status

### Email
- `refyne_email_jobs_processed_total` - Counter: email jobs sent
- `refyne_email_jobs_failures_total` - Counter: failed email jobs

## Dashboard Panels

### 1. Request Rate (top-left)
**Shows:** Requests per second averaged over 5 minutes
- Green: < 100 req/s
- Red: > 100 req/s (should scale horizontally)

### 2. Request Latency (top-right)
**Shows:** P95 and P99 latency in seconds
- Green: < 100ms
- Yellow: 100-500ms
- Red: > 500ms

### 3. Error Rate (middle-left)
**Shows:** 4xx and 5xx errors
- Green: < 10 errors/5min
- Red: > 10 errors

### 4. Database Connections (middle-right)
**Shows:** Active vs used connections in pool
- Yellow threshold: 15/20 (75% used)
- Red threshold: 19/20 (95% used)

### 5. Redis Operations (bottom-left)
**Shows:** Redis command rate (get/set/del/etc)
- Indicates cache hit rate health

### 6. Auth Activity (middle-bottom)
**Shows:** Login attempts vs failures
- Spike in failures = possible attack or auth issues

### 7. Paddle Activity (bottom-right)
**Shows:** Payment API calls and errors
- Indicates subscription processing health

### 8. Rate Limit Violations (far-bottom)
**Shows:** How many requests blocked by rate limiting
- Normal: 0 violations
- Alert if > 100/10min (users hitting limits)

## Key Queries

### Request Rate by Endpoint
```promql
rate(refyne_http_requests_total[5m])
```

### Error Rate (5xx only)
```promql
rate(refyne_http_requests_total{status=~"5.."}[5m])
```

### P99 Latency
```promql
histogram_quantile(0.99, rate(refyne_http_request_duration_seconds_bucket[5m]))
```

### Database Connection Pool Usage
```promql
refyne_db_connections_used / refyne_db_connections_active
```

### Failed Login Rate
```promql
rate(refyne_auth_login_failures_total[5m])
```

### Subscription Status
```promql
refyne_subscriptions_by_tier{status="active"}
```

## Alerting Rules (Future)

Create `monitoring/alerting_rules.yml`:

```yaml
groups:
  - name: refyne_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(refyne_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(refyne_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        annotations:
          summary: "P95 latency > 1s"

      - alert: DatabasePoolExhausted
        expr: refyne_db_connections_used / refyne_db_connections_active > 0.9
        for: 2m
        annotations:
          summary: "Database connection pool 90% full"

      - alert: RedisDown
        expr: up{job="redis"} == 0
        for: 1m
        annotations:
          summary: "Redis service down"

      - alert: PaddleAPIErrors
        expr: rate(refyne_paddle_api_errors_total[5m]) > 0.01
        for: 5m
        annotations:
          summary: "Paddle API errors exceeding threshold"
```

## Troubleshooting

### Prometheus not scraping metrics
```bash
# Check target status
curl http://localhost:9090/api/v1/targets

# Check prometheus logs
docker logs refyne-prometheus

# Verify backend metrics endpoint
curl http://localhost:8080/metrics
```

### Grafana not showing data
```bash
# Check if Prometheus datasource is connected
# Grafana → Configuration → Data Sources → Prometheus
# Test connection button

# Verify queries in Prometheus first
# http://localhost:9090/graph
```

### Metrics not updating
- Increase scrape frequency: Edit `monitoring/prometheus.yml`
- Default: 10 seconds → Change to 5 seconds for faster updates
- Restart Prometheus: `docker restart refyne-prometheus`

### Dashboard showing "No data"
- Backend not running: `make run`
- Metrics endpoint not responding: `curl http://localhost:8080/metrics`
- Wait 30+ seconds for first scrape (initial data collection)

## Production Setup

### Remote Prometheus Storage

For long-term storage, configure remote storage:

```yaml
# Add to prometheus.yml
remote_write:
  - url: "https://prometheus.example.com/write"
    basic_auth:
      username: 'user'
      password: 'pass'
```

### High Availability Setup

Run multiple Prometheus instances behind a load balancer:
- Each scrapes independently
- Redundancy if one fails
- Use AlertManager for deduplication

### Retention Policy

Default: 15 days of metrics

Adjust in `docker-compose-monitoring.yml`:
```yaml
command:
  - '--storage.tsdb.retention.time=30d'  # Keep 30 days
```

## Performance Impact

- Metrics collection: < 5ms per request
- Memory overhead: ~100MB for Prometheus
- Disk usage: ~5GB per month (at 100 req/s)

## Security

### Secure Prometheus

Add basic auth:
```yaml
# prometheus.yml
global:
  external_labels:
    environment: 'production'

http_sd_configs:
  - basic_auth:
      username: 'prometheus'
      password: 'secure_password'
```

### Secure Grafana

- Change default admin password immediately
- Configure LDAP/OAuth if available
- Restrict dashboard access by role
- Enable audit logging

## Maintenance

### Daily
- Check error rate dashboard
- Verify latency < 500ms
- Check database connection usage

### Weekly
- Review failed login attempts
- Check Paddle API error rate
- Verify Redis connection health

### Monthly
- Analyze traffic patterns
- Plan capacity increases
- Review alerting rules effectiveness

## Capacity Planning

Based on current metrics:

- **At 100 req/s:** 1 PostgreSQL replica needed
- **At 1000 req/s:** Add Prometheus remote storage + HAProxy
- **At 10,000 req/s:** Multiple Prometheus instances, distributed database

Current setup comfortable for: **500-1000 req/s**

## Next Steps

1. Monitor for 24 hours to establish baseline
2. Set alert thresholds based on baseline
3. Create runbooks for common alerts
4. Test failover scenarios
5. Document on-call procedures
