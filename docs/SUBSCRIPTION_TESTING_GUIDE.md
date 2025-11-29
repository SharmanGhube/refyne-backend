# Subscription Testing Guide - Phase 1.5

Complete step-by-step guide to test Paddle subscription integration with expected outputs.

---

## 📋 Prerequisites

### **1. Environment Setup**
- ✅ Server running on `http://localhost:8080`
- ✅ PostgreSQL database running
- ✅ Postman installed
- ✅ Email access for OTP (sharmanghube@gmail.com)

### **2. Import Postman Collection**
1. Open Postman
2. Click **Import** → **Files**
3. Import: `Refyne_API.postman_collection.json`
4. Import: `Refyne_Local.postman_environment.json`
5. Select environment: **Refyne Local Development**

### **3. Verify Server is Running**
```powershell
# Start server
cd d:\Refyne\refyne-backend
.\bin\app.exe
```

**Expected Output:**
```
INFO Paddle configuration initialized mode=sandbox sandbox_configured=true
[GIN-debug] POST   /api/subscription/checkout
[GIN-debug] GET    /api/subscription/status
[GIN-debug] POST   /api/subscription/portal
[GIN-debug] POST   /api/webhooks/paddle
INFO Server starting on port 8080
```

---

## 🧪 Testing Flow

---

## **Step 1: Request OTP**

### **Action:**
In Postman: **Authentication** → **2. Request OTP**

**Request:**
```http
POST http://localhost:8080/api/auth/request-otp
Content-Type: application/json

{
  "email": "sharmanghube@gmail.com",
  "password": "Goobs@123"
}
```

### **Expected Response (200 OK):**
```json
{
  "status": "success",
  "message": "OTP sent successfully",
  "data": {
    "request_id": "655ad8ce-65de-448c-8d4a-62914b16e20f",
    "expires_in": 900
  },
  "request_id": "abc123..."
}
```

### **Verify:**
- ✅ Status code: `200`
- ✅ Message: "OTP sent successfully"
- ✅ Check email inbox for OTP (6-digit code)
- ✅ Server logs show: `INFO OTP sent to email`

### **Troubleshooting:**
❌ **401 Unauthorized - Invalid Password**
```json
{
  "error": {
    "code": "INVALID_PASSWORD",
    "message": "Invalid password"
  }
}
```
→ Verify password in `.env` matches: `Goobs@123`

---

## **Step 2: Login with OTP**

### **Action:**
Copy OTP from email (e.g., `835067`)

In Postman: **Authentication** → **3. Login with OTP**

**Request:**
```http
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "sharmanghube@gmail.com",
  "otp": "835067"
}
```

### **Expected Response (200 OK):**
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "user": {
      "id": "65dc3072-5ab0-45b7-91a1-8f4b923bc67e",
      "email": "sharmanghube@gmail.com",
      "username": "testuser_email",
      "is_verified": true,
      "is_active": true
    }
  }
}
```

### **Verify:**
- ✅ Status code: `200`
- ✅ `access_token` is present (long JWT string)
- ✅ `expires_in: 900` (15 minutes)
- ✅ `user.is_verified: true`
- ✅ Token auto-saved to Postman environment variable
- ✅ Server logs show: `INFO User logged in successfully`

### **Troubleshooting:**
❌ **401 Unauthorized - Invalid OTP**
```json
{
  "error": {
    "code": "INVALID_OTP",
    "message": "Invalid or expired OTP"
  }
}
```
→ OTP expired (15 min) or incorrect. Request new OTP.

---

## **Step 3: Check Current Subscription Status**

### **Action:**
In Postman: **Subscription & Payments** → **1. Get Subscription Status**

**Request:**
```http
GET http://localhost:8080/api/subscription/status
Authorization: Bearer {{access_token}}
```

### **Expected Response (200 OK) - No Active Subscription:**
```json
{
  "status": "success",
  "message": "Subscription status retrieved",
  "data": {
    "user_id": "65dc3072-5ab0-45b7-91a1-8f4b923bc67e",
    "email": "sharmanghube@gmail.com",
    "subscription_tier": "starter",
    "subscription_status": "inactive",
    "subscription_expires_at": null,
    "paddle_customer_id": null,
    "paddle_subscription_id": null,
    "can_upgrade": true,
    "is_trial": false
  }
}
```

### **Verify:**
- ✅ Status code: `200`
- ✅ `subscription_tier: "starter"` (default)
- ✅ `subscription_status: "inactive"`
- ✅ `paddle_customer_id: null` (no payment yet)
- ✅ `paddle_subscription_id: null`
- ✅ `can_upgrade: true`

### **Postman Console Output:**
```
📊 Subscription Status:
  Tier: starter
  Status: inactive
  Expires: N/A
  Customer ID: N/A
  Subscription ID: N/A
```

---

## **Step 4: Create Checkout Session (Starter Tier)**

### **Action:**
In Postman: **Subscription & Payments** → **2. Create Checkout - Starter ($29/mo)**

**Request:**
```http
POST http://localhost:8080/api/subscription/checkout
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
  "tier": "starter"
}
```

### **Expected Response (200 OK):**
```json
{
  "status": "success",
  "message": "Checkout session created",
  "data": {
    "checkout_url": "https://sandbox-checkout.paddle.com/checkout/custom/...",
    "tier": "starter"
  },
  "request_id": "def456..."
}
```

### **Verify:**
- ✅ Status code: `200`
- ✅ `checkout_url` starts with `https://sandbox-checkout.paddle.com`
- ✅ URL is valid and clickable
- ✅ Server logs show: `INFO Checkout session created tier=starter`

### **Postman Console Output:**
```
✅ Checkout URL: https://sandbox-checkout.paddle.com/checkout/custom/...
🌐 Open this URL in browser to complete payment
```

### **Backend Logs:**
```
INFO Creating Paddle checkout session tier=starter user_id=65dc3072-5ab0-45b7-91a1-8f4b923bc67e
INFO Checkout session created successfully checkout_id=che_01xxx
```

### **Troubleshooting:**
❌ **400 Bad Request - Invalid Tier**
```json
{
  "error": "Invalid request data",
  "details": "tier must be one of: starter, professional, business, enterprise"
}
```
→ Check tier spelling in request body

❌ **401 Unauthorized**
```json
{
  "error": "Unauthorized",
  "message": "Token has been invalidated. Please login again."
}
```
→ Token expired (15 min). Go back to Step 1.

❌ **500 Internal Server Error - Paddle API Failed**
```json
{
  "error": "Failed to create checkout session",
  "message": "Internal server error"
}
```
→ Check server logs for Paddle API error
→ Verify `PADDLE_SANDBOX_API_KEY` in `.env` is valid

---

## **Step 5: Complete Payment on Paddle**

### **Action:**
1. **Copy** the `checkout_url` from Step 4 response
2. **Open** the URL in your browser
3. You'll see **Paddle Sandbox Checkout Page**

### **Paddle Checkout Page - What You'll See:**
```
Refyne - Starter Plan
$29.00 USD / month

Email: sharmanghube@gmail.com
Payment Method: [Card details form]
```

### **Enter Test Card Details:**
```
Card Number: 4242 4242 4242 4242
Expiry Date: 12/26 (any future date)
CVV: 123 (any 3 digits)
Cardholder Name: Test User
```

### **Billing Address (required):**
```
Country: United States
ZIP Code: 12345
```

### **Click:** "Complete Purchase" or "Subscribe"

### **Expected Outcome:**
1. ✅ Payment processes successfully (instant in sandbox)
2. ✅ Redirect to: `http://localhost:3000/subscription-success` (will 404 if frontend not running - this is OK)
3. ✅ Paddle sends webhooks to your backend

### **Alternative - Simulate Payment Failure:**
To test error handling, use:
```
Card Number: 4000 0000 0000 0002 (Declined card)
```

---

## **Step 6: Verify Webhook Processing**

### **Backend Logs (Immediate):**
After completing payment, check your server terminal:

### **Expected Logs:**
```
INFO Received Paddle webhook event_type=transaction.completed
INFO Validating webhook signature
INFO Webhook signature valid
INFO Processing transaction.completed event transaction_id=txn_01xxx

INFO Received Paddle webhook event_type=subscription.created
INFO Validating webhook signature
INFO Webhook signature valid
INFO Processing subscription.created event subscription_id=sub_01xxx
INFO User subscription created user_id=65dc3072... tier=starter status=active
INFO Successfully updated user subscription customer_id=ctm_01xxx
```

### **Verify:**
- ✅ 2 webhook events received
- ✅ `transaction.completed` processed first
- ✅ `subscription.created` processed second
- ✅ Both signatures validated
- ✅ Database updated with `paddle_customer_id` and `paddle_subscription_id`

### **Troubleshooting:**
❌ **No webhooks received**
→ **Cause:** Paddle can't reach `http://localhost:8080` (local network)

**Solution: Use ngrok**
```powershell
# Install ngrok
choco install ngrok

# Expose localhost
ngrok http 8080

# Copy the ngrok URL (e.g., https://abcd-1234.ngrok.io)
```

**Update Paddle Webhook:**
1. Go to: https://sandbox-vendors.paddle.com/notifications-v2
2. Edit webhook endpoint
3. Change URL to: `https://your-ngrok-url.ngrok.io/api/webhooks/paddle`
4. Save
5. Retry payment

---

## **Step 7: Verify Subscription Status (After Payment)**

### **Action:**
In Postman: **Subscription & Payments** → **1. Get Subscription Status**

**Request:** (Same as Step 3)
```http
GET http://localhost:8080/api/subscription/status
Authorization: Bearer {{access_token}}
```

### **Expected Response (200 OK) - Active Subscription:**
```json
{
  "status": "success",
  "message": "Subscription status retrieved",
  "data": {
    "user_id": "65dc3072-5ab0-45b7-91a1-8f4b923bc67e",
    "email": "sharmanghube@gmail.com",
    "subscription_tier": "starter",
    "subscription_status": "active",
    "subscription_expires_at": "2025-12-29T00:00:00Z",
    "paddle_customer_id": "ctm_01kb6abc123xyz",
    "paddle_subscription_id": "sub_01kb6def456uvw",
    "can_upgrade": true,
    "is_trial": false
  }
}
```

### **Verify:**
- ✅ Status code: `200`
- ✅ `subscription_status: "active"` (changed from "inactive")
- ✅ `subscription_expires_at` is ~30 days in future
- ✅ `paddle_customer_id` starts with `ctm_`
- ✅ `paddle_subscription_id` starts with `sub_`

### **Postman Console Output:**
```
📊 Subscription Status:
  Tier: starter
  Status: active
  Expires: 2025-12-29T00:00:00Z
  Customer ID: ctm_01kb6abc123xyz
  Subscription ID: sub_01kb6def456uvw
```

### **Compare with Step 3:**
| Field | Before Payment | After Payment |
|-------|----------------|---------------|
| `subscription_status` | `inactive` | `active` |
| `subscription_expires_at` | `null` | `2025-12-29T00:00:00Z` |
| `paddle_customer_id` | `null` | `ctm_01kb6abc123xyz` |
| `paddle_subscription_id` | `null` | `sub_01kb6def456uvw` |

---

## **Step 8: Verify Database Updates**

### **Action:**
Open **pgAdmin** or use **psql**:

```sql
SELECT 
    email,
    subscription_tier,
    subscription_status,
    subscription_expires_at,
    paddle_customer_id,
    paddle_subscription_id,
    onboarding_completed,
    updated_at
FROM users 
WHERE email = 'sharmanghube@gmail.com';
```

### **Expected Result:**
```
email                  : sharmanghube@gmail.com
subscription_tier      : starter
subscription_status    : active
subscription_expires_at: 2025-12-29 00:00:00
paddle_customer_id     : ctm_01kb6abc123xyz
paddle_subscription_id : sub_01kb6def456uvw
onboarding_completed   : false
updated_at            : 2025-11-29 03:45:12.456
```

### **Verify:**
- ✅ All 6 subscription fields updated
- ✅ `updated_at` is recent (timestamp of webhook processing)
- ✅ IDs match Postman response

---

## **Step 9: Get Customer Portal URL**

### **Action:**
In Postman: **Subscription & Payments** → **6. Get Customer Portal URL**

**Request:**
```http
POST http://localhost:8080/api/subscription/portal
Authorization: Bearer {{access_token}}
```

### **Expected Response (200 OK):**
```json
{
  "status": "success",
  "message": "Customer portal URL retrieved",
  "data": {
    "portal_url": "https://sandbox-vendors.paddle.com/customers/ctm_01kb6abc123xyz/portal/...",
    "paddle_customer_id": "ctm_01kb6abc123xyz"
  }
}
```

### **Verify:**
- ✅ Status code: `200`
- ✅ `portal_url` starts with `https://sandbox-vendors.paddle.com`
- ✅ URL includes customer ID
- ✅ Server logs show: `INFO Customer portal URL retrieved`

### **Postman Console Output:**
```
✅ Customer Portal URL: https://sandbox-vendors.paddle.com/customers/ctm_01kb6abc123xyz/portal/...
🌐 User can manage subscription here
```

### **Action (Optional):**
Open the `portal_url` in browser - you'll see:
- Current subscription details
- Payment method
- Billing history
- Cancel/Update subscription options

### **Troubleshooting:**
❌ **404 Not Found - No Active Subscription**
```json
{
  "error": {
    "code": "SUBSCRIPTION_NOT_FOUND",
    "message": "No active subscription found"
  }
}
```
→ User doesn't have `paddle_customer_id` in database
→ Webhook processing failed - check server logs

---

## **Step 10: Verify in Paddle Dashboard**

### **Action:**
1. Go to: https://sandbox-vendors.paddle.com/
2. Login to your Paddle Sandbox account

### **Check Subscriptions:**
Navigate: **Subscriptions** tab

**Expected:**
```
Subscription ID: sub_01kb6def456uvw
Customer: sharmanghube@gmail.com
Product: Refyne Starter
Status: Active
Next Billing: 2025-12-29
Amount: $29.00 USD
```

### **Check Customers:**
Navigate: **Customers** tab → Search: `sharmanghube@gmail.com`

**Expected:**
```
Customer ID: ctm_01kb6abc123xyz
Email: sharmanghube@gmail.com
Total Subscriptions: 1
Status: Active
Created: 2025-11-29
```

### **Check Transactions:**
Navigate: **Transactions** tab

**Expected:**
```
Transaction ID: txn_01kb6xxx
Customer: sharmanghube@gmail.com
Amount: $29.00 USD
Status: Completed
Product: Refyne Starter
Date: 2025-11-29 03:45
```

### **Check Webhook Events:**
Navigate: **Developer Tools** → **Notifications** → **Event Logs**

**Expected: 2 Events**

**Event 1:**
```
Event Type: transaction.completed
Status: Success (200)
Timestamp: 2025-11-29 03:45:10
Response: OK
```

**Event 2:**
```
Event Type: subscription.created
Status: Success (200)
Timestamp: 2025-11-29 03:45:12
Response: OK
```

Click **View Payload** to see full JSON sent to your backend.

### **Verify:**
- ✅ Subscription shows in dashboard
- ✅ Customer created with correct email
- ✅ Transaction marked as completed
- ✅ 2 webhook events both succeeded (200 status)

---

## 🧪 Additional Test Cases

---

## **Test Case A: Create Checkout for Professional Tier**

### **Request:**
In Postman: **Subscription & Payments** → **3. Create Checkout - Professional ($99/mo)**

```http
POST http://localhost:8080/api/subscription/checkout
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
  "tier": "professional"
}
```

### **Expected Response:**
```json
{
  "status": "success",
  "message": "Checkout session created",
  "data": {
    "checkout_url": "https://sandbox-checkout.paddle.com/...",
    "tier": "professional"
  }
}
```

**Verify:** Different product ID used for professional tier in Paddle

---

## **Test Case B: Invalid Tier**

### **Request:**
In Postman: **Subscription & Payments** → **ERROR: Invalid Tier**

```http
POST http://localhost:8080/api/subscription/checkout
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
  "tier": "premium"
}
```

### **Expected Response (400 Bad Request):**
```json
{
  "error": "Invalid request data",
  "details": "tier must be one of: starter, professional, business, enterprise",
  "request_id": "xyz789..."
}
```

### **Verify:**
- ✅ Status code: `400`
- ✅ Error message explains valid tiers
- ✅ Server logs show validation error

---

## **Test Case C: No Authentication**

### **Request:**
In Postman: **Subscription & Payments** → **ERROR: No Authentication**

```http
GET http://localhost:8080/api/subscription/status
# No Authorization header
```

### **Expected Response (401 Unauthorized):**
```json
{
  "error": "Unauthorized",
  "message": "Authorization header required",
  "request_id": "abc123..."
}
```

### **Verify:**
- ✅ Status code: `401`
- ✅ Cannot access without token
- ✅ Server logs show: `WARN Unauthorized access attempt`

---

## **Test Case D: Rate Limiting**

### **Action:**
Rapidly send **11 checkout requests** in 1 hour

In Postman:
1. Go to **Subscription & Payments** → **2. Create Checkout - Starter**
2. Click **Send** button 11 times rapidly

### **Expected Response (First 10 requests):**
```json
{
  "status": "success",
  "message": "Checkout session created",
  ...
}
```

### **Expected Response (11th request) - 429 Too Many Requests:**
```json
{
  "error": "Too many requests",
  "message": "Rate limit exceeded. Try again later.",
  "retry_after": 3600,
  "request_id": "def456..."
}
```

### **Verify:**
- ✅ First 10 requests: `200 OK`
- ✅ 11th request: `429 Too Many Requests`
- ✅ `retry_after: 3600` (1 hour in seconds)
- ✅ Server logs show: `WARN Rate limit exceeded`

---

## **Test Case E: Expired Token**

### **Action:**
1. Wait 16 minutes after login (token expires in 15 min)
2. Try to access any protected endpoint

### **Request:**
```http
GET http://localhost:8080/api/subscription/status
Authorization: Bearer {{expired_token}}
```

### **Expected Response (401 Unauthorized):**
```json
{
  "error": "Unauthorized",
  "message": "Token has expired. Please login again.",
  "request_id": "ghi789..."
}
```

### **Verify:**
- ✅ Status code: `401`
- ✅ Message indicates token expiration
- ✅ Must go back to Step 1 (request new OTP)

---

## 🔍 Webhook Event Testing

---

## **Webhook Event 1: subscription.created**

### **Paddle Sends (After Payment):**
```json
{
  "event_id": "evt_01kb6xxx",
  "event_type": "subscription.created",
  "occurred_at": "2025-11-29T03:45:12.000000Z",
  "data": {
    "id": "sub_01kb6def456uvw",
    "status": "active",
    "customer_id": "ctm_01kb6abc123xyz",
    "items": [
      {
        "price": {
          "product_id": "pro_01kb658vg4yn2kfa05bgpea0mn"
        }
      }
    ],
    "next_billed_at": "2025-12-29T00:00:00.000000Z"
  }
}
```

### **Backend Processing:**
```
INFO Processing subscription.created event
INFO Mapping product_id to tier product_id=pro_01kb658vg4yn2kfa05bgpea0mn tier=starter
INFO Updating user subscription user_id=65dc3072... status=active
INFO Database update successful
```

### **Database Update:**
```sql
UPDATE users SET
    subscription_tier = 'starter',
    subscription_status = 'active',
    subscription_expires_at = '2025-12-29 00:00:00',
    paddle_customer_id = 'ctm_01kb6abc123xyz',
    paddle_subscription_id = 'sub_01kb6def456uvw',
    updated_at = NOW()
WHERE paddle_customer_id = 'ctm_01kb6abc123xyz';
```

---

## **Webhook Event 2: subscription.updated**

### **When:** User upgrades from Starter to Professional

### **Paddle Sends:**
```json
{
  "event_type": "subscription.updated",
  "data": {
    "id": "sub_01kb6def456uvw",
    "status": "active",
    "items": [
      {
        "price": {
          "product_id": "pro_01kb65ddszngn8rw1xxgt4d9dz"
        }
      }
    ],
    "next_billed_at": "2025-12-29T00:00:00.000000Z"
  }
}
```

### **Backend Processing:**
```
INFO Processing subscription.updated event
INFO Tier changed old_tier=starter new_tier=professional
INFO Database update successful
```

### **Database Update:**
```sql
UPDATE users SET
    subscription_tier = 'professional',
    subscription_status = 'active',
    updated_at = NOW()
WHERE paddle_subscription_id = 'sub_01kb6def456uvw';
```

---

## **Webhook Event 3: subscription.canceled**

### **When:** User cancels subscription

### **Paddle Sends:**
```json
{
  "event_type": "subscription.canceled",
  "data": {
    "id": "sub_01kb6def456uvw",
    "status": "canceled",
    "canceled_at": "2025-11-29T04:00:00.000000Z"
  }
}
```

### **Backend Processing:**
```
INFO Processing subscription.canceled event
INFO Subscription canceled subscription_id=sub_01kb6def456uvw
INFO Database update successful
```

### **Database Update:**
```sql
UPDATE users SET
    subscription_status = 'canceled',
    updated_at = NOW()
WHERE paddle_subscription_id = 'sub_01kb6def456uvw';
```

---

## 📊 Complete Test Summary

### **Successful Test Run Should Include:**

✅ **Authentication Flow**
- [ ] OTP requested and received
- [ ] Login successful with valid OTP
- [ ] Access token obtained (valid for 15 min)

✅ **Subscription Status (Before Payment)**
- [ ] Status: inactive
- [ ] No Paddle IDs present
- [ ] Default tier: starter

✅ **Checkout Creation**
- [ ] Checkout URL generated
- [ ] URL opens in browser
- [ ] Paddle checkout page loads

✅ **Payment Processing**
- [ ] Test card accepted
- [ ] Payment successful
- [ ] Redirect to success URL

✅ **Webhook Processing**
- [ ] 2 webhooks received
- [ ] Signatures validated
- [ ] Database updated

✅ **Subscription Status (After Payment)**
- [ ] Status: active
- [ ] Paddle customer ID present
- [ ] Paddle subscription ID present
- [ ] Expiry date set (~30 days)

✅ **Database Verification**
- [ ] All 6 subscription fields populated
- [ ] Values match API responses

✅ **Paddle Dashboard Verification**
- [ ] Subscription visible
- [ ] Customer created
- [ ] Transaction completed
- [ ] Webhooks succeeded (200 status)

✅ **Customer Portal**
- [ ] Portal URL retrieved
- [ ] URL opens in browser
- [ ] Subscription details visible

✅ **Error Cases**
- [ ] Invalid tier rejected (400)
- [ ] No auth rejected (401)
- [ ] Rate limit enforced (429 after 10 req)
- [ ] Expired token rejected (401)

---

## 🐛 Common Issues & Solutions

### **Issue 1: Webhooks Not Received**

**Symptom:** Payment completes but database not updated

**Cause:** Paddle can't reach `http://localhost:8080`

**Solution:**
```powershell
# Use ngrok to expose localhost
ngrok http 8080

# Update webhook URL in Paddle dashboard to ngrok URL
```

---

### **Issue 2: Webhook Signature Validation Failed**

**Symptom:** Logs show "Invalid webhook signature"

**Cause:** Wrong `PADDLE_SANDBOX_WEBHOOK_SECRET` in `.env`

**Solution:**
1. Go to Paddle → Developer Tools → Notifications
2. Copy webhook secret key
3. Update `.env`:
```env
PADDLE_SANDBOX_WEBHOOK_SECRET=ntfset_01kb652fnnwxmbpgagfdrdrj60
```
4. Restart server

---

### **Issue 3: Token Expired**

**Symptom:** "Token has been invalidated" error

**Cause:** Access token expires after 15 minutes

**Solution:**
1. Go back to Step 1
2. Request new OTP
3. Login again
4. New token auto-saves

---

### **Issue 4: Rate Limit Hit**

**Symptom:** 429 Too Many Requests

**Cause:** More than 10 checkout requests in 1 hour

**Solution:**
```sql
-- Clear rate limit in database (dev only!)
DELETE FROM rate_limit_entries WHERE identifier = 'sharmanghube@gmail.com';
```

Or wait 1 hour for automatic reset.

---

### **Issue 5: Paddle API Error**

**Symptom:** 500 Internal Server Error when creating checkout

**Cause:** Invalid Paddle API key or product IDs

**Solution:**
1. Verify in `.env`:
```env
PADDLE_SANDBOX_API_KEY=pdl_sdbx_apikey_01kb64spayf1444r4pf3q1h4x4_bYcgZ5r2tCRWczKaSx4hzV_AEX
PADDLE_SANDBOX_PRODUCT_ID_STARTER=pro_01kb658vg4yn2kfa05bgpea0mn
```
2. Check Paddle dashboard → Products → Verify product IDs
3. Restart server

---

## 📈 Performance Benchmarks

### **Expected Response Times (Sandbox):**

| Endpoint | Expected Time | Acceptable |
|----------|--------------|------------|
| `POST /auth/request-otp` | 200-500ms | < 1s |
| `POST /auth/login` | 100-300ms | < 500ms |
| `GET /subscription/status` | 50-150ms | < 300ms |
| `POST /subscription/checkout` | 500-1500ms | < 3s |
| `POST /subscription/portal` | 300-800ms | < 2s |
| `POST /webhooks/paddle` | 100-300ms | < 500ms |

**Note:** Paddle API calls (checkout, portal) are slower due to external API latency.

---

## 🎯 Test Completion Checklist

Print this and check off as you test:

```
SUBSCRIPTION TESTING - PHASE 1.5
================================

PRE-FLIGHT
[ ] Server running on port 8080
[ ] Database running
[ ] Postman collection imported
[ ] Environment selected

AUTHENTICATION
[ ] OTP requested
[ ] OTP received via email
[ ] Login successful
[ ] Token saved in Postman

SUBSCRIPTION STATUS (BEFORE)
[ ] Status: inactive
[ ] No Paddle IDs
[ ] Default tier: starter

CHECKOUT CREATION
[ ] Checkout URL generated for Starter
[ ] Checkout URL generated for Professional
[ ] Checkout URL generated for Business
[ ] Checkout URL generated for Enterprise
[ ] Invalid tier rejected (400)

PAYMENT FLOW
[ ] Paddle checkout page opens
[ ] Test card accepted
[ ] Payment successful
[ ] Redirect to success URL

WEBHOOK PROCESSING
[ ] transaction.completed received
[ ] subscription.created received
[ ] Both signatures validated
[ ] Database updated with Paddle IDs

SUBSCRIPTION STATUS (AFTER)
[ ] Status: active
[ ] paddle_customer_id present
[ ] paddle_subscription_id present
[ ] Expiry date set

CUSTOMER PORTAL
[ ] Portal URL retrieved
[ ] Portal opens in browser
[ ] Subscription details visible

DATABASE VERIFICATION
[ ] All 6 fields updated
[ ] Values match API responses
[ ] updated_at is recent

PADDLE DASHBOARD
[ ] Subscription visible
[ ] Customer exists
[ ] Transaction completed
[ ] Webhooks succeeded (200)

ERROR HANDLING
[ ] No auth returns 401
[ ] Invalid tier returns 400
[ ] Expired token returns 401
[ ] Rate limit enforced (429)

COMPLETION
[ ] All tests passed
[ ] No errors in logs
[ ] Ready for Phase 2
```

---

## 📚 Next Steps

After completing all tests:

1. ✅ **Phase 1.5 Complete** - Subscription infrastructure tested
2. 🚀 **Proceed to Phase 2** - Workspace Management
3. 📝 **Document any issues** found during testing
4. 🔄 **Repeat tests** after code changes

---

## 💡 Tips for Testing

- **Use Postman Console** (`View` → `Show Postman Console`) to see all requests/responses
- **Keep server logs visible** to catch webhook processing in real-time
- **Use ngrok** for webhook testing (Paddle can't reach localhost)
- **Test all 4 tiers** to ensure product ID mapping works
- **Clear tokens** between test runs to simulate fresh user
- **Check Paddle dashboard** after each payment to verify data sync

---

**Testing Guide Version:** 1.0  
**Last Updated:** November 29, 2025  
**Status:** ✅ Ready for Testing
