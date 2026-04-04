# Grafana Cloud Production Setup (Railway)

**Status:** Production-ready for Railway deployment
**Cost:** Free tier (sufficient for MVP)
**Effort:** ~10 minutes one-time setup

## What is Grafana Cloud?

Grafana Cloud is a managed monitoring platform that:
- Stores your metrics (Prometheus remote storage)
- Hosts dashboards and alerts
- Handles scalability, backups, security
- **Free tier:** 3 stacks, 10GB metrics storage, 30-day retention

Perfect for Refyne on Railway because:
- No extra services to manage on Railway
- Backend just ships metrics to cloud
- Automatically backed up and secure

---

## Step 1: Create Grafana Cloud Account (One-time)

1. Go to https://grafana.com/signup/
2. Sign up with email
3. Verify email
4. Create new organization/stack:
   - Stack name: `refyne-production`
   - Region: Select closest to you (or `us-central1`)
5. Go to **Admin → API Keys**
6. Create new API token:
   - Name: `refyne-metrics-write`
   - Role: `MetricsPublisher`
   - Copy the token (save securely)

---

## Step 2: Configure Backend for Metrics Shipping

Add these environment variables to **Railway** (or `.env` for local testing):

```env
# Grafana Cloud Configuration
GRAFANA_CLOUD_ENABLED=true
GRAFANA_CLOUD_PROMETHEUS_URL=https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push
GRAFANA_CLOUD_API_KEY=<your-api-token-from-step-1>
GRAFANA_CLOUD_USERNAME=<your-grafana-username>
```

**Get these values from Grafana Cloud:**
1. Login to https://grafana.com/auth/grafana_com/login
2. Click your stack
3. **Connections → Data Sources → Prometheus**
4. Copy the "Remote Write URL" (should start with `https://prometheus-blocks-prod...`)
5. Go to **Admin → Users** to find your username

---

## Step 3: Backend Code (Already Done ✅)

The backend is already configured with Prometheus metrics:
- Metrics endpoint: `GET /metrics`
- Exposed metrics: HTTP requests, latency, errors, DB connections, Redis, etc.
- Metrics are collected in-memory and exposed in Prometheus format

**No code changes needed** - metrics are already being collected.

---

## Step 4: Test Locally Before Railway Push

```bash
# Set environment variables
export GRAFANA_CLOUD_ENABLED=true
export GRAFANA_CLOUD_PROMETHEUS_URL=https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push
export GRAFANA_CLOUD_API_KEY=your_token_here
export GRAFANA_CLOUD_USERNAME=your_username

# Start backend (includes metrics endpoint)
make run

# In another terminal, verify metrics are exposed
curl http://localhost:8080/metrics | head -20

# Check Grafana Cloud dashboard after ~2 minutes
# Login to https://YOUR-ORG.grafana.net
```

---

## Step 5: Deploy to Railway

1. **Add environment variables to Railway:**
   ```
   GRAFANA_CLOUD_ENABLED=true
   GRAFANA_CLOUD_PROMETHEUS_URL=https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push
   GRAFANA_CLOUD_API_KEY=<your-token>
   GRAFANA_CLOUD_USERNAME=<your-username>
   ```

2. **Push to main branch:**
   ```bash
   git add .
   git commit -m "Add Grafana Cloud configuration for production monitoring"
   git push origin main
   ```

3. **Railway auto-deploys** → metrics flow to Grafana Cloud

---

## Step 6: View Production Metrics

1. Login to Grafana Cloud: https://YOUR-ORG.grafana.net
2. **Dashboards → Browse → refyne-production** (or create new)
3. Click **+ New Dashboard**
4. Add panels using PromQL queries:

### Example Queries:

**HTTP Requests per Second:**
```promql
rate(http_requests_total[1m])
```

**API Latency (p95):**
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**Error Rate:**
```promql
rate(http_requests_total{status=~"5.."}[1m])
```

**Database Connections:**
```promql
db_connections_active
```

---

## Step 7: Set Up Alerts (Optional)

1. Go to **Alerts & IRM → Alerting → Contact Points**
2. Add notification channel (Slack, email, PagerDuty, etc.)
3. Create alert rule:
   - **Condition:** `rate(http_requests_total{status="500"}[1m]) > 0.1`
   - **Notify:** When error rate exceeds 0.1 req/s
   - **Channel:** Your notification method

---

## Production Checklist

- [ ] Grafana Cloud account created
- [ ] API token generated
- [ ] Environment variables set on Railway
- [ ] `/metrics` endpoint verified working
- [ ] Metrics appearing in Grafana Cloud dashboard (after 2 min)
- [ ] Dashboard created with key metrics
- [ ] Alerts configured (optional but recommended)

---

## Troubleshooting

**Metrics not appearing in Grafana Cloud?**
```bash
# 1. Verify metrics endpoint returns data
curl http://your-railway-backend:8080/metrics

# 2. Check backend logs for metrics shipping errors
# (in Railway dashboard → Logs)

# 3. Verify API token is correct
# Test token in Grafana UI: Admin → API Keys

# 4. Wait 2-3 minutes for data to appear (first-time delay)
```

**Can't find Remote Write URL?**
- Go to Grafana Cloud
- Click your stack
- **Connections → Data Sources → Prometheus**
- Scroll down to "Remote write" section
- Copy the URL

---

## Scaling Up (Later)

If free tier isn't enough:
- **10GB+ storage:** Upgrade to paid plan (~$50-200/month depending on usage)
- **Custom retention:** Adjust in stack settings (free tier: 30 days)
- **More dashboards:** Unlimited in any tier

---

## Local Development (Without Grafana Cloud)

Keep using local Prometheus + Grafana:
```bash
# Terminal 1: Start backend
make run

# Terminal 2: Start monitoring stack
docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d

# Access: http://localhost:3000 (Grafana)
```

The local stack is still useful for development and testing dashboards before moving to production.
