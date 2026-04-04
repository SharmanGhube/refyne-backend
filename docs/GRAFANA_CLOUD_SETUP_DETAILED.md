# Grafana Cloud API Key Generation & Railway Setup Guide

Complete step-by-step guide to get metrics monitoring working on Railway for free.

---

## Part 1: Create Grafana Cloud Account (5 minutes)

### Step 1: Sign Up
1. Go to https://grafana.com/signup/
2. Choose **"Sign up with email"** (or GitHub if you prefer)
3. Enter your email and create password
4. Click "Create account"
5. **Check your email and verify** (click confirmation link)

### Step 2: Create Your First Stack
After email verification, you'll see the Grafana Cloud dashboard:

1. Click **"Create stack"** (or **"New stack"**)
2. Fill in:
   - **Stack name:** `refyne-production`
   - **Region:** Choose closest to you (e.g., `us-central1` if USA, `eu-west-1` if EU)
   - **Grafana version:** Leave as default (latest)
3. Click **"Create stack"**
4. Wait 2-3 minutes for stack to be created
5. You'll see: **"Your stack is ready!"** ✅
6. Click **"Open Grafana"** → You're now in your Grafana Cloud dashboard

---

## Part 2: Generate API Key (2 minutes)

Now you're inside Grafana Cloud. You need to create an API token for metrics upload.

### Step 3: Navigate to API Keys

1. In Grafana Cloud, click your **avatar (circle icon)** in bottom left corner
2. Click **"Administration"** (or **"Admin"** in sidebar)
3. In left sidebar, click **"API keys"**
4. You'll see: **"API keys"** page with a button **"+ New API key"** or **"Create API key"**

### Step 4: Create Metrics Publisher Key

1. Click **"+ New API key"** button
2. Fill in the form:
   ```
   Name:        refyne-metrics-writer
   Role:        MetricsPublisher (IMPORTANT: Must be "MetricsPublisher", not "Admin")
   Expiration:  Leave blank (or set to 1 year)
   Permissions: Default (auto-selected)
   ```
3. Click **"Create API key"**
4. **IMPORTANT:** A modal will pop up with your token:
   ```
   Your API key is:
   glc_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9eyJhbGc...
   ```
5. **COPY THIS TOKEN IMMEDIATELY** and save it somewhere safe (you can't see it again!)
   - You can paste it in a .txt file temporarily
   - Or keep this browser tab open while setting up Railway

### ⚠️ Warning
- **Don't close this modal** until you copy the token
- If you miss it, delete the key and create a new one
- The token starts with `glc_` prefix

---

## Part 3: Get Your Stack Info (2 minutes)

You need these details to configure the backend:

### Step 5: Find Stack Name (Username)
1. Still in Grafana Cloud, click your **avatar** (bottom left)
2. Click **"Profile"**
3. Look for: **"Login name"** or **"Username"** (e.g., `yourname@example.com` or `username`)
   - **Write this down** - you'll need it as `GRAFANA_CLOUD_USERNAME`

### Step 6: Find Remote Write URL
1. In Grafana Cloud, go to **"Connections"** (left sidebar)
2. Click **"Data sources"**
3. Look for **"Prometheus"** in the list (should be there by default)
4. Click **"Prometheus"**
5. Scroll down to **"Remote write"** section
6. You'll see:
   ```
   Remote write configuration (for Prometheus remote_write)
   URL: https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push
   ```
7. **Copy this URL** - you'll need it as `GRAFANA_CLOUD_PROMETHEUS_URL`

---

## Part 4: Add Environment Variables to Railway (3 minutes)

Now you have all the information needed:

| Variable | Example Value |
|----------|---|
| `GRAFANA_CLOUD_ENABLED` | `true` |
| `GRAFANA_CLOUD_API_KEY` | `glc_eyJhbGciOiJIUzI1NiI...` |
| `GRAFANA_CLOUD_USERNAME` | `yourname@example.com` |
| `GRAFANA_CLOUD_PROMETHEUS_URL` | `https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push` |

### Step 7: Go to Railway Dashboard
1. Go to https://railway.app
2. Login with your GitHub account
3. Click on your **Refyne Backend project**
4. You should see: Dashboard with services (db, redis, backend)
5. Click the **Backend service** (the main API)

### Step 8: Add Environment Variables
1. On the backend service page, look for **"Variables"** tab (or **"Settings"**)
2. Click **"+ Add variable"** or **"New variable"**
3. Add each variable:

**Variable 1:**
```
Name:  GRAFANA_CLOUD_ENABLED
Value: true
```
Click **"Add"**

**Variable 2:**
```
Name:  GRAFANA_CLOUD_API_KEY
Value: glc_eyJhbGciOiJIUzI1NiI...  (the token you copied from Grafana Cloud)
```
Click **"Add"**

**Variable 3:**
```
Name:  GRAFANA_CLOUD_USERNAME
Value: yourname@example.com  (your Grafana Cloud username/email)
```
Click **"Add"**

**Variable 4:**
```
Name:  GRAFANA_CLOUD_PROMETHEUS_URL
Value: https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push
```
Click **"Add"**

### Step 9: Deploy
After adding all 4 variables:
1. Go to **"Deployments"** tab
2. You should see Railway is **auto-redeploying** (look for "In progress" status)
3. Wait for it to complete (~2-3 minutes)
4. Once it says **"Active"** with a green checkmark ✅, you're deployed!

---

## Part 5: Verify Metrics Are Flowing (3 minutes)

### Step 10: Check Backend Metrics Endpoint
1. Open your Railway backend URL (something like `https://refyne-backend-xyz.up.railway.app`)
2. Add `/metrics` to the URL:
   ```
   https://refyne-backend-xyz.up.railway.app/metrics
   ```
3. You should see raw Prometheus metrics (lots of text starting with `#`)
   - If you see metrics: ✅ Backend is exposing metrics correctly

### Step 11: Check Grafana Cloud Dashboard
1. Go back to https://grafana.com
2. Login if needed
3. Click your **Refyne stack** → **"Open Grafana"**
4. Wait 2-3 minutes (first metrics take time to show)
5. Go to **"Explore"** (left sidebar)
6. In the query box, type:
   ```
   up
   ```
7. Click **"Run query"** button
8. **If you see data:** ✅ Metrics are flowing to Grafana Cloud!

---

## Troubleshooting

### Metrics Not Showing in Grafana Cloud?

**Check 1: Verify Railway deployment succeeded**
```
Go to Railway → Backend service → Deployments
Status should be "Active" (green checkmark)
```

**Check 2: Verify metrics endpoint works**
```
curl https://your-railway-backend.up.railway.app/metrics
Should return Prometheus metrics text
```

**Check 3: Verify environment variables are set**
```
Go to Railway → Backend service → Variables
All 4 variables should be listed
```

**Check 4: Check Railway logs for errors**
```
Go to Railway → Backend service → Logs
Search for "GRAFANA" to see if there are any error messages
```

**Check 5: Wait longer**
- First metrics can take 2-5 minutes to appear
- Refresh Grafana Cloud page
- Try the "up" query again (see Step 11)

---

## Quick Reference Checklist

- [ ] Created Grafana Cloud account (free)
- [ ] Created stack named `refyne-production`
- [ ] Generated API key with "MetricsPublisher" role (copied token)
- [ ] Found stack username (email/GitHub)
- [ ] Found Remote Write URL
- [ ] Added 4 environment variables to Railway backend:
  - [ ] `GRAFANA_CLOUD_ENABLED=true`
  - [ ] `GRAFANA_CLOUD_API_KEY=glc_...`
  - [ ] `GRAFANA_CLOUD_USERNAME=...`
  - [ ] `GRAFANA_CLOUD_PROMETHEUS_URL=https://...`
- [ ] Railway auto-deployed with new variables
- [ ] Verified metrics endpoint works (`/metrics`)
- [ ] Verified metrics appear in Grafana Cloud (after 2-5 min)

---

## Next Steps (Optional)

Once metrics are flowing:

1. **Create a Dashboard** in Grafana Cloud
   - Go to "Dashboards" → "New Dashboard"
   - Add panels with queries like `rate(http_requests_total[1m])`

2. **Set Up Alerts** (optional)
   - Go to "Alerting" → "Alert rules"
   - Create alert if error rate > 10%
   - Connect to Slack/email notification

3. **Local Development** (still useful)
   - Keep using `docker-compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d`
   - View local Grafana at http://localhost:3000
   - Test dashboards before pushing to production

---

## Support

**Still stuck?**
- Check Railway logs: `Railway → Backend → Logs`
- Check Grafana Cloud stack health: `https://grafana.com → Your stack → Status`
- Verify connectivity: Can Railway backend reach `prometheus-blocks-prod-us-central1.grafana.net`?
