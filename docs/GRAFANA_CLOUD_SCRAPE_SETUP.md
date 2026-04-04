# Grafana Cloud Scrape Job Setup

This guide explains how to configure Grafana Cloud to scrape metrics from the Refyne backend `/metrics` endpoint.

## Overview

Instead of pushing metrics to Grafana Cloud (which requires protobuf format), we use **Grafana Cloud's scrape model**:
- The backend exposes metrics at `/metrics` endpoint
- Grafana Cloud pulls these metrics via HTTP GET at regular intervals
- Eliminates format conversion issues and simplifies setup

## Prerequisites

- ✅ Backend deployed on Railway with `/metrics` endpoint accessible
- ✅ Grafana Cloud account (free tier available)
- ✅ Service account created (from earlier setup)

## Setup Steps

### Step 1: Get Your Backend URL

From Railway dashboard:
1. Go to **Backend** service
2. Copy the **public URL** (e.g., `https://your-railway-app.up.railway.app`)
3. The metrics endpoint is: `https://your-railway-app.up.railway.app/metrics`

### Step 2: Configure Scrape Job in Grafana Cloud

1. Go to **Grafana Cloud** → **Connections** → **Add new connection**
2. Search for and select **Prometheus** (if not already added)
3. Click **Create a Prometheus data source**
4. Configure with these settings:

```
Data source name: Your choice (e.g., "Refyne Backend Metrics")
Prometheus server URL: https://prometheus-prod-43-prod-ap-south-1.grafana.net/api/prom
Default: Yes
```

### Step 3: Create Scrape Configuration

In Grafana Cloud Prometheus settings, add a scrape job using the Grafana Agent or via configuration:

**Option A: Using Grafana Agent (Recommended)**

If using Grafana Cloud with managed Prometheus, use their Grafana Agent:

1. Go to **Grafana Cloud** → **Manage your stack** → **Grafana Agent**
2. Click **Edit configuration**
3. Add this scrape job to `scrape_configs`:

```yaml
scrape_configs:
  - job_name: 'refyne-backend'
    scrape_interval: 15s
    scrape_timeout: 10s
    static_configs:
      - targets: ['your-railway-app.up.railway.app']
    scheme: 'https'
    relabel_configs:
      - source_labels: [__address__]
        regex: '([^:]+)(?::\d+)?'
        target_label: __address__
        replacement: '${1}/metrics'
```

**Option B: Using curl to test**

Test that your endpoint is accessible from anywhere:

```bash
curl -v https://your-railway-app.up.railway.app/metrics
```

You should see Prometheus format output starting with:
```
# HELP refyne_http_requests_total HTTP request counter
# TYPE refyne_http_requests_total counter
```

### Step 4: Verify Metrics Flow

1. Go to **Grafana Cloud** → **Explore**
2. Select your Prometheus data source
3. In the query box, try:
   ```
   up{job="refyne-backend"}
   ```
4. Or any custom metric:
   ```
   refyne_http_requests_total
   ```

Metrics should appear within 30-60 seconds of the first scrape.

## Available Metrics

The backend exposes these metrics:

- **HTTP Metrics**
  - `refyne_http_requests_total` - Total HTTP requests by endpoint, method, status
  - `refyne_http_request_duration_seconds` - Request latency histogram
  - `refyne_http_request_size_bytes` - Request body size
  - `refyne_http_response_size_bytes` - Response body size

- **Authentication**
  - `refyne_auth_login_success_total` - Successful login attempts
  - `refyne_auth_login_failed_total` - Failed login attempts
  - `refyne_auth_token_issued_total` - JWT tokens issued

- **Database**
  - `refyne_db_connections_active` - Active database connections
  - `refyne_db_query_duration_seconds` - Query execution time

- **Cache/Redis**
  - `refyne_redis_operations_total` - Redis command counts
  - `refyne_redis_operation_duration_seconds` - Redis operation latency

- **Rate Limiting**
  - `refyne_rate_limit_exceeded_total` - Rate limit violations

- **Paddle Integration** (if configured)
  - `refyne_paddle_webhooks_total` - Webhook events received
  - `refyne_paddle_webhooks_failed_total` - Failed webhook processing

## Troubleshooting

### Metrics Not Appearing

1. **Check backend is running**
   ```bash
   curl https://your-railway-app.up.railway.app/api/health
   ```
   Should return status 200

2. **Check /metrics endpoint**
   ```bash
   curl https://your-railway-app.up.railway.app/metrics
   ```
   Should return Prometheus format data

3. **Check Grafana Agent logs**
   - Go to Grafana Cloud dashboard → Logs
   - Search for scrape errors with label `job="refyne-backend"`

4. **Verify scrape configuration**
   - In Grafana, go to **Connections** → **Data sources** → **Prometheus**
   - Click **Test** button to verify connectivity

### Authentication Issues

If using Basic Auth in the scrape job:

```yaml
basic_auth:
  username: sharmanghube
  password: [your-grafana-api-token]
```

### Firewall/Network Issues

- Ensure Railway backend is publicly accessible (not in private network)
- Check Railway security settings allow external HTTPS requests to `/metrics`
- Verify Grafana Cloud IPs are not blocked

## Creating Dashboards

Once metrics are flowing:

1. Go to **Grafana Cloud** → **Build a dashboard**
2. Click **Add a new panel**
3. Select your Prometheus data source
4. Example panels to create:
   - HTTP requests per second: `rate(refyne_http_requests_total[1m])`
   - Error rate: `rate(refyne_http_requests_total{status=~"5.."}[1m])`
   - P95 latency: `histogram_quantile(0.95, refyne_http_request_duration_seconds_bucket)`

## Disable Remote Push (Optional)

The remote push functionality is disabled by default. If you want to keep it disabled in future deployments, set in Railway variables:

```
GRAFANA_CLOUD_ENABLED=false
```

## Cost Considerations

- **Grafana Cloud Free Tier:**
  - 3 stacks
  - 10GB storage
  - 30-day retention
  - Unlimited scrape jobs

- **Estimated cost for 15-second scrape interval:**
  - ~5,760 data points/day (96 metrics × 60 values/hour)
  - Well within free tier limits

## Next Steps

- Set up alert rules for critical metrics
- Create custom dashboards for your team
- Configure Slack/email notifications for alerts
- Monitor performance trends over time
