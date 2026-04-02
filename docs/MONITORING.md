# Monitoring Guide - Prometheus & Grafana

## Overview

This guide covers setting up production monitoring for Refyne Backend using Prometheus for metrics collection and Grafana for visualization.

**What's Monitored:**
- HTTP request rates, latency (P95/P99), and error rates
- Database connection pool health
- Redis operation performance
- Authentication metrics (login attempts, failures)
- Email job processing
- Paddle subscription API calls
- Rate limiting exceeded events

---

## Part 1: Local Development Setup

### Prerequisites

- Docker & Docker Compose
- `docker-compose-monitoring.yml` configured
- Prometheus and Grafana images available

### Quick Start

**Start the monitoring stack:**

```bash
cd d:/Refyne/refyne-backend
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d
```

This starts:
- Refyne Backend: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

**Verify all services are running:**

```bash
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml ps
```

Expected output:
```
NAME          STATUS
refyne        Up (healthy)
db            Up (healthy)
redis         Up (healthy)
prometheus    Up (healthy)
grafana       Up (healthy)
```

**Generate traffic to collect metrics:**

```bash
# E2E tests will generate realistic traffic
go test ./tests -v

# Or manually
for i in {1..100}; do
  curl -s http://localhost:8080/api/health > /dev/null
done
```

---

## Part 2: Prometheus

### Access Prometheus UI

Open: http://localhost:9090

### Querying Metrics

**Example 1: HTTP Request Rate (requests per second)**

```promql
rate(refyne_http_requests_total[5m])
```

**Example 2: Request Latency (P95)**

```promql
histogram_quantile(0.95, rate(refyne_http_request_duration_seconds_bucket[5m]))
```

**Example 3: Error Rate**

```promql
rate(refyne_http_requests_total{status=~"5..|4.."}[5m])
```

**Example 4: Database Connections**

```promql
refyne_db_connections_active
```

**Example 5: Authentication Failures**

```promql
rate(refyne_auth_login_failures_total[5m])
```

### Metrics Endpoint

Raw metrics available at: http://localhost:8080/metrics

Expected format:
```
# HELP refyne_http_requests_total Total number of HTTP requests received
# TYPE refyne_http_requests_total counter
refyne_http_requests_total{endpoint="/api/health",method="GET",status="200"} 42
```

### Configuration

**File:** `monitoring/prometheus.yml`

**Key Settings:**
- `scrape_interval: 15s` - Collect metrics every 15 seconds
- `scrape_timeout: 10s` - Timeout individual scrape requests at 10 seconds
- `targets: ['localhost:8080']` - Scrape backend at :8080/metrics

**Edit configuration:**

```bash
# Edit prometheus.yml
nano monitoring/prometheus.yml

# Restart Prometheus to apply changes
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml restart prometheus
```

---

## Part 3: Grafana

### Access Grafana UI

Open: http://localhost:3000

**Default credentials:**
- Username: `admin`
- Password: `admin`

### Import Dashboard

1. Click "+" → "Import"
2. Select `monitoring/grafana-dashboard.json`
3. Choose Prometheus data source
4. Click "Import"

**Dashboard Panels:**

| Panel | Metric | Description |
|---|---|---|
| Request Rate | `rate(refyne_http_requests_total[5m])` | HTTP requests/second by endpoint |
| Latency (P95/P99) | `histogram_quantile(0.95, ...)` | Request latency percentiles |
| Error Rate | `rate(refyne_http_requests_total{status=~"[45].."}[5m])` | 4xx and 5xx errors per second |
| DB Connections | `refyne_db_connections_active/_used` | Active and in-use connections |
| Redis Operations | `rate(refyne_redis_operations_total[5m])` | Redis ops/sec by operation |
| Email Jobs | `rate(refyne_email_jobs_processed_total[5m])` | Email delivery rate and failures |
| Paddle API | `rate(refyne_paddle_api_calls_total[5m])` | Paddle API call rate |
| Subscriptions | `refyne_subscriptions_by_tier` | Active subscriptions by tier |

### Creating Custom Dashboards

**Create new dashboard:**

1. Click "+" → "Dashboard"
2. Click "Add panel"
3. Enter Prometheus query (e.g., `refyne_http_requests_total`)
4. Configure visualization (Graph, Gauge, Stat, etc.)
5. Save dashboard

**Example Custom Panel:**

```promql
# Top error endpoints (last 5 minutes)
rate(refyne_http_requests_total{status=~"5.."}[5m])
```

### Setting Alerts (Optional)

1. Open dashboard panel
2. Click "Alert" tab
3. Configure alert condition (e.g., "if P95 latency > 100ms for 5m")
4. Configure notification channel (email, Slack, etc.)

---

## Part 4: Metrics Reference

### HTTP Metrics

```promql
# Total requests
refyne_http_requests_total{method, endpoint, status}

# Request duration (latency)
refyne_http_request_duration_seconds{method, endpoint}
  Buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 seconds
```

**Usage Examples:**

```promql
# Requests/sec by status
rate(refyne_http_requests_total[1m])

# P95 latency
histogram_quantile(0.95, rate(refyne_http_request_duration_seconds_bucket[5m]))

# 5xx error rate
rate(refyne_http_requests_total{status=~"5.."}[5m])
```

### Database Metrics

```promql
# Active connections in pool
refyne_db_connections_active{pool}

# Connections currently in use
refyne_db_connections_used{pool}
```

### Redis Metrics

```promql
# Operations per second
refyne_redis_operations_total{operation, status}

# Errors
refyne_redis_errors_total{operation, error_type}
```

### Authentication Metrics

```promql
# Login attempts by method (otp, password, refresh)
refyne_auth_login_attempts_total{method}

# Failed logins by reason
refyne_auth_login_failures_total{reason}

# Tokens generated
refyne_auth_tokens_generated_total{token_type}
```

### Subscription Metrics

```promql
# Active subscriptions by tier and status
refyne_subscriptions_by_tier{tier, status}

# Paddle API calls
refyne_paddle_api_calls_total{operation, status}

# Paddle API errors
refyne_paddle_api_errors_total{operation, error_code}
```

### Email Metrics

```promql
# Processed email jobs
refyne_email_jobs_processed_total{email_type, status}

# Failed email jobs
refyne_email_jobs_failures_total{email_type, reason}
```

### Rate Limiting Metrics

```promql
# Requests rejected by rate limiter
refyne_rate_limit_exceeded_total{endpoint, client_ip}
```

---

## Part 5: Production Deployment

### Deploy to Railway

Prometheus and Grafana are optional in production (Railway provides basic monitoring).

**Option 1: Railway Native Monitoring** (Recommended)
- Railway dashboard shows logs, CPU, memory
- Built-in, no setup required
- Limited to basic metrics

**Option 2: Self-Hosted Prometheus + Grafana**
- Deploy Prometheus and Grafana as separate Railway services
- Requires more resources
- Full control and customization

### Enable Monitoring in Production

1. Ensure `prometheus` dependency is included:
   ```bash
   go get github.com/prometheus/client_golang
   ```

2. Metrics endpoint is always exposed at `/metrics`
   - Authentication: No auth required for `/metrics`
   - Consider restricting via firewall/proxy in production

3. Initialize monitoring on app start:
   ```go
   monitoring.Initialize()
   ```

### Sampling Issues in Production

**Problem:** Too many metrics causes storage bloat

**Solution:** Use metric sampling/scraping intervals

```yaml
# In prometheus.yml
global:
  scrape_interval: 30s    # Increase from 15s to 30s
  scrape_timeout: 15s
```

### High-Cardinality Metrics

**At-risk metrics:**

```promql
# This can explode with unique client IPs
refyne_rate_limit_exceeded_total{endpoint, client_ip}

# This varies by endpoint
refyne_http_requests_total{method, endpoint, status}
```

**Solution:** Aggregate in Prometheus

```yaml
# Add relabeling to limit cardinality
metric_relabelings:
  - source_labels: [client_ip]
    regex: '.*'
    target_label: client_ip
    replacement: 'masked'
```

---

## Part 6: Troubleshooting

### Prometheus not scraping metrics

**Check:**
1. Refyne app is running: `curl http://localhost:8080/api/health`
2. Metrics endpoint responds: `curl http://localhost:8080/metrics`
3. Prometheus can reach app: Check Prometheus UI → Status → Targets

**Fix:**

```bash
# Restart Prometheus
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml restart prometheus

# View Prometheus logs
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml logs prometheus
```

### Grafana not showing data

**Check:**
1. Prometheus has data: Visit http://localhost:9090 and query a metric
2. Datasource is configured: Grafana → Configuration → Data Sources → Prometheus
3. Dashboard queries are valid: Click panel → Edit → Check PromQL query

**Fix:**

```bash
# Restart Grafana
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml restart grafana

# View Grafana logs
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml logs grafana
```

### High memory usage

**Problem:** Too many metrics or long retention

**Solution:**
```yaml
# In prometheus.yml
global:
  scrape_interval: 30s  # Increase to 30-60s

# Retention
command:
  - '--storage.tsdb.retention.time=7d'  # Reduce from 30d
```

### Metrics endpoint slow

**Problem:** Generating metrics takes too long

**Solution:** Reduce instrumentation or use sampling

```go
// In metrics recording
if rand.Intn(100) < 10 {  // Record 10% of events
    metrics.RecordEvent(...)
}
```

---

## Part 7: Maintenance

### Backup Grafana Dashboards

**Export dashboard:**
1. Click dashboard → "Dashboard settings" (gear icon)
2. Click "More" → "Export"
3. Save JSON to version control

### Clean Prometheus Storage

```bash
# Delete old data (keep last 7 days)
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml exec prometheus \
  promtool query instant 'time() - 7*24*60*60' 300
```

### Monitor Monitoring (Meta!)

```promql
# Prometheus memory usage
process_resident_memory_bytes{job="prometheus"}

# Grafana memory usage
process_resident_memory_bytes{job="grafana"}

# Prometheus scrape latency
scrape_duration_seconds
```

---

## Next Steps

1. **Set up alerts** for critical metrics (high error rate, latency, connection pool exhaustion)
2. **Enable persistent storage** for historical analysis
3. **Integrate with incident management** (PagerDuty, Opsgenie)
4. **Configure log aggregation** (ELK, Loki) to correlate with metrics
5. **Track SLOs** using metrics (99.9% uptime, <100ms P95 latency)

---

## Summary

| Tool | Purpose | URL | User |
|---|---|---|---|
| Prometheus | Metrics collection & storage | http://localhost:9090 | N/A |
| Grafana | Visualization & dashboards | http://localhost:3000 | admin / admin |
| `/metrics` | Raw Prometheus format | http://localhost:8080/metrics | N/A |
