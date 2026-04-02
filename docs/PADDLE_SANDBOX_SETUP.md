# Paddle Sandbox Setup Guide

## Overview

This guide walks you through setting up Paddle Sandbox mode for testing payments in Refyne without charging real cards. Once tested and working, you can switch to Production mode for live payments.

## Payment Modes

Refyne supports three Paddle modes controlled by the `PAYMENT_MODE` environment variable:

| Mode | Purpose | API Keys Used | Use Case |
|---|---|---|---|
| `mock` | Testing without Paddle API | None needed | Local development, automated tests |
| `sandbox` | Real API calls to sandbox environment | Sandbox keys | Integration testing, staging |
| `production` | Real payments charged | Live keys | Production |

## Step 1: Create Paddle Account

1. Go to https://www.paddle.com (NOT sandbox-vendors.paddle.com yet)
2. Create account or sign in
3. Complete business verification

## Step 2: Create Sandbox Account

1. Once verified, go to https://sandbox-vendors.paddle.com
2. Create a new sandbox account (separate from production)
3. Verify email

## Step 3: Obtain Sandbox API Key

In Paddle Sandbox Vendor Dashboard:

1. Click "Developers" → "API credentials" (or Settings → Developers)
2. Generate new API key:
   - Click "Generate key"
   - Copy the key (starts with `pdl_`)
   - This is your `PADDLE_SANDBOX_API_KEY`

Example format:
```
pdl_sandbox_apikey_01kmg59k2jfaj5aqfev0m590dr_CsjSN419gAPSNqNjCVjD7J_AGa
```

Save this in a secure location.

## Step 4: Obtain Webhook Secret

In Paddle Sandbox Vendor Dashboard:

1. Click "Developers" → "Webhooks"
2. Create new webhook:
   - **Destination URL:** `https://your-refyne-domain.railway.app/api/webhook/paddle`
   - **Events:** Select all relevant events:
     - `subscription.created`
     - `subscription.updated`
     - `subscription.canceled`
     - `transaction.completed`
     - `transaction.updated`

3. Copy the webhook signing secret
4. This is your `PADDLE_SANDBOX_WEBHOOK_SECRET`

Example format:
```
ntfset_01kmg5p43qm4py8d9rqganehx9
```

**Note:** Replace `your-refyne-domain` with your actual Railway domain.

## Step 5: Create Product Prices in Sandbox

For each subscription tier, create a price in Paddle Sandbox:

1. Go to Products page in Sandbox Vendor Dashboard
2. Create product: "Refyne Subscriptions" (if not existing)
3. Create prices for each tier:

### Price 1: Starter ($29/month)
- Name: "Refyne Starter"
- Billing cycle: Monthly
- Price: $29
- **Copy the Pricer ID** → Set as `PADDLE_SANDBOX_PRODUCT_ID_STARTER`

### Price 2: Professional ($99/month)
- Name: "Refyne Professional"
- Billing cycle: Monthly
- Price: $99
- **Copy the Pricer ID** → Set as `PADDLE_SANDBOX_PRODUCT_ID_PROFESSIONAL`

### Price 3: Business ($299/month)
- Name: "Refyne Business"
- Billing cycle: Monthly
- Price: $299
- **Copy the Pricer ID** → Set as `PADDLE_SANDBOX_PRODUCT_ID_BUSINESS`

### Price 4: Enterprise (Custom)
- Name: "Refyne Enterprise"
- Billing cycle: Monthly
- Price: $0 (custom amount at checkout)
- **Copy the Pricer ID** → Set as `PADDLE_SANDBOX_PRODUCT_ID_ENTERPRISE`

Example Pricer IDs:
```
pri_01kb65b3gzy2xn21nh0zw922yn
pri_01kb65e87ysyhrzh9gd372jgyd
pri_01kb65fj789924z9y2qd70x6f3
pri_01kb65gvdrbky2hs3y1hegmbqv
```

## Step 6: Configure Environment Variables

### Local Development (`.env`)

```env
PAYMENT_MODE=sandbox
PADDLE_SANDBOX_API_KEY=pdl_sandbox_apikey_01kmg59k2jfaj5aqfev0m590dr_CsjSN419gAPSNqNjCVjD7J_AGa
PADDLE_SANDBOX_WEBHOOK_SECRET=ntfset_01kmg5p43qm4py8d9rqganehx9
PADDLE_SANDBOX_PRODUCT_ID_STARTER=pri_01kb65b3gzy2xn21nh0zw922yn
PADDLE_SANDBOX_PRODUCT_ID_PROFESSIONAL=pri_01kb65e87ysyhrzh9gd372jgyd
PADDLE_SANDBOX_PRODUCT_ID_BUSINESS=pri_01kb65fj789924z9y2qd70x6f3
PADDLE_SANDBOX_PRODUCT_ID_ENTERPRISE=pri_01kb65gvdrbky2hs3y1hegmbqv
```

### Railway Deployment

1. Click Refyne service → Variables tab
2. Add each variable from above
3. Confirm all are visible in the dashboard
4. Redeploy the service (automatic on next push or manual restart)

## Step 7: Test Sandbox Integration

### Local Testing

1. Ensure app is running: `make run`
2. Check logs for:
   ```
   Running in sandbox payment mode
   Paddle configuration validated successfully
   ```

3. Test subscription creation via API:
   ```bash
   curl -X POST http://localhost:8080/api/subscription/checkout \
     -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "tier": "starter",
       "currency": "USD",
       "email": "test@example.com"
     }'
   ```

   Response should contain a Paddle checkout URL:
   ```json
   {
     "checkout_url": "https://sandbox-checkout.paddle.com/...",
     "tier": "starter"
   }
   ```

### Test Card Numbers

Use these test card numbers in Paddle Sandbox checkout:

| Card Type | Number | Exp | CVC |
|---|---|---|---|
| Visa (Success) | 4111 1111 1111 1111 | 12/25 | 123 |
| Visa (Decline) | 4000 0000 0000 0002 | 12/25 | 123 |
| Mastercard | 5555 5555 5555 4444 | 12/25 | 123 |

### Verify Webhook Events

1. After payment in sandbox, check:
   - Database: Should have new subscription record
   - Logs: Should show webhook received and processed
   - Paddle Sandbox Dashboard → Events: Should see transaction events

## Step 8: Monitor Webhook Verification

The app validates webhook signatures using HMAC-SHA256:

```go
// From internal/domains/subscription/services/paddle_sandbox_service.go
MAC := hmac.New(sha256.New, []byte(cfg.WebhookSecret))
MAC.Write(body)
expectedSignature := base64.URLEncoding.EncodeToString(MAC.Sum(nil))
```

If webhook signature fails:
- Check `PADDLE_SANDBOX_WEBHOOK_SECRET` is correct
- Verify webhook destination URL matches Railway domain exactly
- Check logs for: `Webhook signature verification failed`

## Step 9: Switch to Production (When Ready)

When launching to production:

1. Create production Paddle account (verified)
2. Go to https://vendors.paddle.com (NOT sandbox)
3. Obtain production:
   - `PADDLE_LIVE_API_KEY`
   - `PADDLE_LIVE_WEBHOOK_SECRET`
   - Product IDs for each tier

4. Update environment variables:
   ```env
   PAYMENT_MODE=production
   PADDLE_LIVE_API_KEY=pdl_live_apikey_...
   PADDLE_LIVE_WEBHOOK_SECRET=ntfset_...
   PADDLE_LIVE_PRODUCT_ID_STARTER=pri_...
   # ... etc for all tiers
   ```

5. Update webhook destination to production URL

6. Redeploy

7. Verify logs show: `Running in production payment mode`

## Troubleshooting

### "Webhook signature verification failed"
- ✅ Solution: Ensure webhook destination URL in Paddle matches exactly your Railway domain
- ✅ Solution: Verify `PADDLE_SANDBOX_WEBHOOK_SECRET` is correctly copied (no spaces/typos)

### "Product ID not found"
- ✅ Solution: Ensure Pricer IDs are correctly copied from Paddle Sandbox
- ✅ Solution: Confirm you're using Sandbox tokens in sandbox, Production tokens in production

### "Invalid API key"
- ✅ Solution: Confirm key starts with `pdl_` not `pdl_live_`
- ✅ Solution: Check for leading/trailing spaces in environment variable

### API Call Fails - "401 Unauthorized"
- ✅ Solution: Verify `PADDLE_SANDBOX_API_KEY` has not expired
- ✅ Solution: Generate new key from Paddle Sandbox Dashboard

### No webhook events received
- ✅ Solution: Ensure webhook URL is publicly accessible (Railway URL, not localhost)
- ✅ Solution: Check Paddle Sandbox Dashboard → Events for delivery failures
- ✅ Solution: Verify `/api/webhook/paddle` endpoint is not behind authentication

## Reference: Code Integration Points

### Payment Service Factory
**File:** `internal/domains/subscription/services/paddle_factory.go`
- Automatically selects correct service (Mock/Sandbox/Production) based on `PAYMENT_MODE`

### Configuration Validation
**File:** `internal/domains/subscription/config/paddle_config.go`
- Validates required credentials on startup
- Fails fast if incomplete configuration

### Webhook Handler
**File:** `internal/api/handlers/webhook_handler.go`
- Receives webhook events from Paddle
- Verifies signature
- Updates subscription status in database

### Subscription Service
**File:** `internal/domains/subscription/services/subscription_service.go`
- Initiates checkout with Paddle
- Returns checkout URL to frontend
