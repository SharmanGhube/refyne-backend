# Railway Monitoring Deployment Guide

This guide covers deploying the Prometheus + Grafana monitoring stack to Railway or other production environments.

## Overview

The monitoring stack for Refyne consists of:
- **Prometheus:** Time-series metrics database (scrapes `/metrics` endpoint)
- **Grafana:** Visualization and dashboarding (connects to Prometheus)
- **Backend API:** Exposes metrics via Prometheus client library

## Option 1: Grafana Cloud (Recommended for Cloud)

Grafana Cloud is the easiest cloud option with free tier (3 stacks, 10GB metrics storage).

### Setup Steps:

1. **Create Grafana Cloud Account**
   - Go to https://grafana.com/signup/
   - Sign up for free tier
   - Create new stack

2. **Get Your Stack URLs**
   - Dashboard URL: `https://YOUR-ORG.grafana.net`
   - Prometheus URL: `https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/...`
   - API Token: Create in Admin → API keys

3. **Configure Backend to Ship Metrics**
   - Option A: Use Prometheus remote write
   - Option B: Configure scrape config in Grafana Agent
   - Option C: Stay with local Prometheus (see Option 2)

4. **Import Dashboard**
   - In Grafana UI: Dashboards → Import
   - Paste content from `monitoring/grafana/dashboards/refyne-backend.json`
   - Select Prometheus datasource

## Option 2: Docker Compose on VPS/Self-Hosted

For on-premises or VPS deployment:

```bash
# 1. Copy monitoring files to server
scp -r monitoring/ user@server:/opt/refyne/

# 2. Pull and run stack
docker-compose -f docker-compose-monitoring.yml up -d

# 3. Access services
# Prometheus: http://YOUR-SERVER:9090
# Grafana: http://YOUR-SERVER:3000 (admin/admin)
```

### Configuration

Update `monitoring/prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'refyne-backend'
    static_configs:
      - targets: ['YOUR-BACKEND-URL:8080']  # or Railway backend URL
    scrape_interval: 10s
```

## Option 3: Railway + External Prometheus

Deploy additional Prometheus/Grafana services on Railway:

### Step 1: Add Prometheus Service to Railway

Create `railway-prometheus.yml`:
```yaml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - prometheus-data:/prometheus
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=15d'
    networks:
      - refyne-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
    networks:
      - refyne-network
    depends_on:
      - prometheus

volumes:
  prometheus-data:
  grafana-data:

networks:
  refyne-network:
    external: true
```

Process:
1. Create new Railway service from Docker image
2. Use postgres service PostgreSQL URL for data persistence
3. Configure environment variables in Railway dashboard

## Option 4: Prometheus Remote Write (Cloud Native)

Ship metrics directly from backend to managed backend:

```go
// In your metrics initialization
remoteWriteConfig := &remotewrite.Config{
    URL: "https://YOUR-PROMETHEUS.cloud/api/v1/write",
    Headers: map[string]string{
        "Authorization": "Bearer YOUR-TOKEN",
    },
}
```

Supported platforms:
- Grafana Cloud
- Prometheus Cloud
- Cortex
- Thanos

## Recommended Setup for Railway

**For simplicity, use Grafana Cloud:**

1. Create free Grafana Cloud account
2. Get Prometheus remote write URL and token
3. Update backend to ship metrics (configure remote write)
4. Import dashboard in Grafana Cloud

**Minimal configuration needed - no extra Railway services.**

## Backend Health Check

Verify metrics are being collected:

```bash
# Check metrics endpoint
curl -s http://BACKEND-URL:8080/metrics | head -20

# Should show Prometheus format output:
# TYPE refyne_http_requests_total counter
# TYPE refyne_request_duration_seconds histogram
# ...
```

## Grafana Cloud Configuration

### Add Prometheus Data Source:

1. Configuration → Data Sources → Add
2. Type: Prometheus
3. URL: `https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push`
4. HTTP headers:
   - Authorization: `Bearer YOUR-API-TOKEN`

### Import Custom Dashboard:

```bash
# Export from local Grafana
curl -s -H "Authorization: Bearer YOUR-TOKEN" \
  https://YOUR-ORG.grafana.net/api/dashboards/uid/refyne \
  | jq .dashboard > custom-dashboard.json

# Import into Grafana Cloud
# UI: Create → Import → Paste JSON
```

## Metrics Available

All metrics from `internal/monitoring/metrics.go`:

```
# HTTP/API Metrics
refyne_http_requests_total (counter)
refyne_request_duration_seconds (histogram)

# Database Metrics
refyne_db_connections_active (gauge)
refyne_db_query_duration_seconds (histogram)

# Redis Metrics
refyne_redis_operations_total (counter)
refyne_redis_operation_duration_seconds (histogram)

# Auth Metrics
refyne_auth_login_attempts_total (counter)
refyne_auth_failures_total (counter)

# Paddle/Payment Metrics
refyne_paddle_api_calls_total (counter)
refyne_paddle_api_errors_total (counter)

# Rate Limiting Metrics
refyne_rate_limit_violations_total (counter)
```

## Troubleshooting

### Metrics Not Appearing in Grafana

1. Verify `/metrics` endpoint is working:
   ```bash
   curl http://localhost:8080/metrics
   ```

2. Check Prometheus scrape config points to correct backend URL

3. Wait 1-2 minutes for first scrape (default 10s interval)

4. Check Prometheus UI: Status → Targets (all should be "UP")

### Grafana Can't Connect to Prometheus

1. Verify network connectivity between services
2. Check auth headers if using authentication
3. Ensure data source URL is correct
4. Check Prometheus is running and healthy

### High Memory Usage

- Reduce retention: `--storage.tsdb.retention.time=7d` (vs default 15d)
- Reduce scrape frequency: increase `scrape_interval`
- Reduce cardinality: remove high-cardinality labels

## Next Steps

1. Choose deployment option (recommend Grafana Cloud for MVP)
2. Set up metrics collection
3. Configure dashboard
4. Set up alerting rules (optional)
5. Monitor in production

## References

- Prometheus Docs: https://prometheus.io/docs/
- Grafana Cloud: https://grafana.com/products/cloud/
- Docker Compose: `docker-compose-monitoring.yml`
- Dashboard Config: `monitoring/grafana/dashboards/refyne-backend.json`
