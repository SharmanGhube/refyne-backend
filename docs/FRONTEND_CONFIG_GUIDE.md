# Frontend Configuration & API Integration Guide

**Last Updated:** April 17, 2026  
**Backend Status:** ✅ Production-Ready  
**API Version:** 1.0.0

---

## 🚀 Quick Start for Frontend Developers

### Environment Variables

Create `.env.local` in your frontend project:

```env
# Development (Local Backend)
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Refyne
NEXT_PUBLIC_APP_VERSION=1.0.0

# Production (Railway Backend)
# NEXT_PUBLIC_API_URL=https://your-refyne-service.railway.app
```

### Connect Backend to Frontend

```typescript
// lib/api.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export const api = {
  // Authentication
  register: (data) => fetch(`${API_BASE}/api/auth/register`, { method: 'POST', body: JSON.stringify(data) }),
  login: (email, otp) => fetch(`${API_BASE}/api/auth/login`, { method: 'POST', body: JSON.stringify({ email, otp }) }),
  logout: (token) => fetch(`${API_BASE}/api/auth/logout`, { method: 'POST', headers: { 'Authorization': `Bearer ${token}` } }),
  
  // Subscription
  checkout: (token, tier) => fetch(`${API_BASE}/api/subscription/checkout`, { 
    method: 'POST', 
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ tier })
  }),
  getSubscription: (token) => fetch(`${API_BASE}/api/subscription/status`, { 
    headers: { 'Authorization': `Bearer ${token}` } 
  }),
  
  // User Profile
  getMe: (token) => fetch(`${API_BASE}/api/protected/me`, { 
    headers: { 'Authorization': `Bearer ${token}` } 
  }),
  
  // Health
  health: () => fetch(`${API_BASE}/api/health`)
}
```

---

## 📍 Backend Locations

### Development Environment (Local)

| Service | URL | Purpose |
|---------|-----|---------|
| **Backend API** | `http://localhost:8080` | Main API server |
| **Database (PgAdmin)** | `http://localhost:5050` | Visual database admin |
| **Redis UI** | `http://localhost:5540` | Redis data explorer |
| **DB Connection** | `localhost:5432` | Raw PostgreSQL connection |
| **Redis Connection** | `localhost:6379` | Raw Redis connection |

### Production Environment (Railway)

| Service | URL |
|---------|-----|
| **Backend API** | `https://your-refyne-service.railway.app` |
| **Database** | Private Railway PostgreSQL (not exposed) |
| **Redis** | Private Railway Redis (not exposed) |

---

## 🔐 Authentication Flow (Complete)

### 1. User Registration

```typescript
// Frontend sends
POST http://localhost:8080/api/auth/register
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe"
}

// Backend responds
{
  "status": "success",
  "data": {
    "user_id": "uuid-here",
    "email": "user@example.com"
  }
}

// Frontend action: Show "Check your email"
```

### 2. Email Verification (User clicks link from email)

```typescript
// Frontend extracts token from URL query: ?token=xxx
POST http://localhost:8080/api/auth/verify
{
  "token": "verification-token-from-email"
}

// Backend responds
{
  "status": "success",
  "message": "Email verified successfully"
}

// Frontend action: Redirect to login with message "You can now log in"
```

### 3. Request OTP (on login page)

```typescript
// Frontend sends
POST http://localhost:8080/api/auth/request-otp
{
  "email": "user@example.com"
}

// Backend responds
{
  "status": "success",
  "data": {
    "expires_in": 300  // 5 minutes
  }
}

// Frontend action: Show "OTP sent to email, valid for 5 minutes"
```

### 4. Login with OTP

```typescript
// Frontend sends
POST http://localhost:8080/api/auth/login
{
  "email": "user@example.com",
  "otp": "123456"
}

// Backend responds
{
  "status": "success",
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "full_name": "John Doe",
      "subscription_status": "free",
      "subscription_tier": null
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh-uuid-here",
    "expires_in": 900  // 15 minutes
  }
}

// Frontend action:
localStorage.setItem('access_token', response.access_token)
localStorage.setItem('refresh_token', response.refresh_token)
// Redirect to dashboard
```

### 5. Refresh Expired Token

```typescript
// When access_token expires (after 15 minutes), request new one
POST http://localhost:8080/api/auth/refresh
{
  "refresh_token": "refresh-uuid-here"
}

// Backend responds
{
  "status": "success",
  "data": {
    "access_token": "new-jwt-token",
    "expires_in": 900
  }
}

// Frontend action:
localStorage.setItem('access_token', response.access_token)
// Retry original request with new token
```

### 6. Access Protected Endpoints

```typescript
// All protected endpoints require Authorization header
GET http://localhost:8080/api/protected/me
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

// Backend responds
{
  "user_id": "uuid",
  "email": "user@example.com",
  "username": "john_doe"
}
```

### 7. Logout

```typescript
// Frontend sends (must be authenticated)
POST http://localhost:8080/api/auth/logout
Headers:
  Authorization: Bearer access_token

// Backend responds
{
  "status": "success",
  "message": "Logout successful"
}

// Frontend action:
localStorage.removeItem('access_token')
localStorage.removeItem('refresh_token')
// Redirect to login
```

---

## 💳 Subscription / Payment Flow

### Paddle Integration Overview

```
User clicks "Subscribe"
          ↓
Frontend calls POST /api/subscription/checkout
          ↓
Backend creates Paddle transaction → returns checkout_url
          ↓
Frontend redirects to Paddle checkout page
          ↓
User enters payment details on Paddle (not on your site)
          ↓
User completes payment
          ↓
Paddle sends webhook to backend → updates database
          ↓
User redirected to success page
          ↓
Frontend polls GET /api/subscription/status
          ↓
Once subscription_status = "active" → show dashboard
```

### Step 1: Check Subscription Status

```typescript
// Frontend sends
GET http://localhost:8080/api/subscription/status
Headers:
  Authorization: Bearer access_token

// Backend responds
{
  "status": "success",
  "data": {
    "subscription_status": "free",
    "subscription_tier": null,
    "paddle_customer_id": null,
    "paddle_subscription_id": null,
    "subscription_started_at": null,
    "subscription_ends_at": null,
    "cancel_at_period_end": false
  }
}

// OR (after user has subscription)
{
  "status": "success",
  "data": {
    "subscription_status": "active",
    "subscription_tier": "pro",
    "paddle_customer_id": "ctm_01kb65a1234567890",
    "paddle_subscription_id": "sub_01kb65b9876543210",
    "subscription_started_at": "2026-04-17T10:30:00Z",
    "subscription_ends_at": "2026-05-17T10:30:00Z",
    "cancel_at_period_end": false
  }
}
```

### Step 2: Create Checkout

```typescript
// Frontend sends
POST http://localhost:8080/api/subscription/checkout
Headers:
  Authorization: Bearer access_token
  Content-Type: application/json

Body:
{
  "tier": "pro"
}

// Backend responds
{
  "status": "success",
  "data": {
    "checkout_url": "https://sandbox-checkout.paddle.com/p/test_01h8h8h8h8h8h8h8h8h8h8h8h8h8",
    "transaction_id": "txn_01kb7x6vwjep7aqpqtyfkmgg9v",
    "expires_in": 3600
  }
}

// Frontend action:
window.location.href = data.data.checkout_url
// User taken to Paddle checkout page (NOT your site)
```

### Step 3: Handle Checkout Success

```typescript
// User completes payment on Paddle
// Paddle redirects to: FRONTEND_CHECKOUT_SUCCESS_URL
// From .env: http://localhost:3000/subscription-success

// On this page, poll for subscription update
const pollSubscription = async () => {
  const response = await fetch('http://localhost:8080/api/subscription/status', {
    headers: { 'Authorization': `Bearer ${localStorage.getItem('access_token')}` }
  })
  const data = await response.json()
  
  if (data.data.subscription_status === 'active') {
    // Subscription confirmed!
    router.push('/dashboard')
  }
}

// Poll every 2 seconds for up to 30 seconds
let attempts = 0
const interval = setInterval(() => {
  pollSubscription()
  attempts++
  if (attempts > 15) clearInterval(interval)
}, 2000)
```

### Step 4: Manage Subscription (Customer Portal)

```typescript
// If user wants to change plan, cancel subscription, etc.
POST http://localhost:8080/api/subscription/portal
Headers:
  Authorization: Bearer access_token

// Backend responds
{
  "status": "success",
  "data": {
    "portal_url": "https://sandbox-checkout.paddle.com/portal/ctm_01kb65a1234567890/subscriptions",
    "expires_in": 3600
  }
}

// Frontend action:
window.open(data.data.portal_url, '_blank')
// Opens Paddle customer portal in new tab
```

---

## 🏠 User Profile Endpoints

### Get Current User

```typescript
GET http://localhost:8080/api/user/profile
Headers:
  Authorization: Bearer access_token

// Response (200)
{
  "id": "uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "username": "john_doe",
  "profile_picture": "https://...",
  "bio": "...",
  "created_at": "2026-04-01T10:00:00Z"
}
```

### Update User Profile

```typescript
PUT http://localhost:8080/api/user/profile
Headers:
  Authorization: Bearer access_token
  Content-Type: application/json

Body:
{
  "first_name": "John",
  "last_name": "Doe",
  "username": "john_doe_new",
  "bio": "Updated bio"
}

// Response (200)
{
  "status": "success",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "username": "john_doe_new"
  }
}
```

### Get User Settings

```typescript
GET http://localhost:8080/api/user/settings
Headers:
  Authorization: Bearer access_token

// Response (200)
{
  "language": "en",
  "timezone": "UTC",
  "email_notifications": true,
  "theme": "light"
}
```

### Update User Settings

```typescript
PUT http://localhost:8080/api/user/settings
Headers:
  Authorization: Bearer access_token
  Content-Type: application/json

Body:
{
  "language": "en",
  "timezone": "America/New_York",
  "email_notifications": false,
  "theme": "dark"
}

// Response (200)
{
  "status": "success",
  "data": {
    "language": "en",
    "timezone": "America/New_York",
    "email_notifications": false,
    "theme": "dark"
  }
}
```

---

## 🏥 Health Check Endpoints

Use these to check if backend is online:

```typescript
// Basic health check
GET http://localhost:8080/api/health
// { "status": "healthy" }

// Detailed health with component status
GET http://localhost:8080/api/health/detailed
// { "status": "healthy", "checks": { "database": "healthy", "redis": "healthy" } }

// Readiness (for deployment)
GET http://localhost:8080/api/health/ready
// { "status": "ready" }

// Liveness (for health checks)
GET http://localhost:8080/api/health/live
// { "status": "alive" }
```

---

## 📋 HTTP Status Codes & Error Handling

### Standard Responses

```typescript
// Success (200, 201)
{
  "status": "success",
  "data": { ... }
}

// Error (400, 401, 409, etc)
{
  "error": "Email already exists",
  "code": "EMAIL_ALREADY_EXISTS",
  "details": {
    "email": "user@example.com"
  }
}
```

### Common Status Codes

| Code | Meaning | Action |
|------|---------|--------|
| 200 | Success | Process normally |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Show validation errors |
| 401 | Unauthorized | Attempt token refresh, then redirect to login |
| 409 | Conflict | Show conflict error (e.g., "Email already exists") |
| 429 | Rate Limited | Show "Too many requests, try again in X minutes" |
| 500 | Server Error | Show generic error, log to monitoring |

### Error Handling Example

```typescript
try {
  const response = await fetch(`${API_BASE}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, otp })
  })
  
  if (!response.ok) {
    const error = await response.json()
    
    if (response.status === 401) {
      throw new Error('Invalid OTP. Please try again.')
    } else if (response.status === 429) {
      throw new Error('Too many requests. Please try again later.')
    } else if (response.status === 400) {
      throw new Error(error.error || 'Invalid request')
    } else {
      throw new Error('Login failed. Please try again.')
    }
  }
  
  const data = await response.json()
  return data.data
} catch (error) {
  // Handle error - show to user
  console.error(error.message)
}
```

---

## 🔄 Token Lifecycle

### Token Expiry Times

```
Access Token:   15 minutes (900 seconds)
Refresh Token:  7 days
```

### Token Refresh Intercept

```typescript
// Axios interceptor pattern
api.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      // Try to refresh token
      const refreshToken = localStorage.getItem('refresh_token')
      
      try {
        const response = await fetch(`${API_BASE}/api/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken })
        })
        
        if (response.ok) {
          const data = await response.json()
          localStorage.setItem('access_token', data.data.access_token)
          
          // Retry original request
          error.config.headers.Authorization = `Bearer ${data.data.access_token}`
          return api(error.config)
        }
      } catch (e) {
        // Refresh failed - logout user
        localStorage.clear()
        window.location.href = '/login?session_expired=true'
      }
    }
    
    return Promise.reject(error)
  }
)
```

---

## ⚙️ Environment Setup

### Development (.env.local)

```env
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080

# App Metadata
NEXT_PUBLIC_APP_NAME=Refyne
NEXT_PUBLIC_APP_VERSION=1.0.0

# Features (for frontend flags)
NEXT_PUBLIC_ENABLE_ANALYTICS=true
NEXT_PUBLIC_ENABLE_SENTRY=false

# Paddle Sandbox (for payment testing)
NEXT_PUBLIC_PADDLE_SANDBOX_MODE=true
```

### Production (.env.production)

```env
# API Configuration
NEXT_PUBLIC_API_URL=https://your-refyne-service.railway.app

# App Metadata
NEXT_PUBLIC_APP_NAME=Refyne
NEXT_PUBLIC_APP_VERSION=1.0.0

# Features
NEXT_PUBLIC_ENABLE_ANALYTICS=true
NEXT_PUBLIC_ENABLE_SENTRY=true

# Paddle Production (for live payments)
NEXT_PUBLIC_PADDLE_SANDBOX_MODE=false
```

---

## 🧪 Testing Credentials

### Test User Registration

```
Email: test@example.com
Password: TestPass123!
Full Name: Test User

✅ Requirements:
- Minimum 8 characters
- At least 1 uppercase
- At least 1 lowercase
- At least 1 number
- At least 1 special character
```

### Test Payment (Paddle Sandbox)

```
Card Number: 4111 1111 1111 1111
Expiry: 12/25
CVV: 123
ZIP: 12345

✅ This card will be accepted in sandbox
```

### Test OTP

When you request OTP, check:
1. **Local dev:** Backend console logs show the OTP
2. **Email integration:** Check inbox or spam folder

---

## 📞 Common Issues & Fixes

### "CORS Error: Access-Control-Allow-Origin"

**Problem:**
```
Access to XMLHttpRequest has been blocked by CORS policy
```

**Solution:**
1. Check `FRONTEND_URL` is in backend's `CORS_ALLOWED_ORIGINS`
2. Verify URL exactly matches (including port)

```env
# Backend .env must include frontend URL
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://your-domain.com
```

---

### "Token Invalid" on Protected Endpoints

**Problem:**
```
401 Unauthorized
```

**Solution:**
```typescript
// 1. Check token exists
console.log(localStorage.getItem('access_token'))

// 2. Check token format (should be JWT with 3 parts: xxx.yyy.zzz)
const token = localStorage.getItem('access_token')
console.log(token.split('.').length === 3 ? 'Valid format' : 'Invalid format')

// 3. Check Authorization header
fetch(`${API_BASE}/api/protected/me`, {
  headers: {
    'Authorization': `Bearer ${token}`  // Must be "Bearer <token>"
  }
})

// 4. If still failing, try refreshing token
const refreshToken = localStorage.getItem('refresh_token')
// Call refresh endpoint
```

---

### "Email Already Exists" on Register

**Problem:**
```json
{ "error": "email already exists", "code": "EMAIL_ALREADY_EXISTS" }
```

**Solution:**
```typescript
// User with this email already registered
// Options:
// 1. Show message: "This email is already registered"
// 2. Offer password reset: "Did you forget your password?"
// 3. Suggest login instead
```

---

### "Too Many Requests" (429)

**Problem:**
```
You are making too many requests. Please try again later.
```

**Solution:**
```typescript
// Rate limit: 100 requests/minute per endpoint
// Wait and retry after a few moments

// Show user: "Too many attempts. Please try again in X minutes"

// Store rate limit info
if (response.status === 429) {
  const retryAfter = response.headers.get('Retry-After') || '60'
  localStorage.setItem('nextRetryAt', Date.now() + retryAfter * 1000)
}
```

---

## 🚀 Deployment Checklist

Before deploying frontend to production:

- [ ] Backend is deployed on Railway
- [ ] `NEXT_PUBLIC_API_URL` points to production backend
- [ ] JWT secret is same on backend and verified
- [ ] CORS `FRONTEND_URL` is added to backend
- [ ] Email provider is configured (SMTP)
- [ ] Paddle is switched to production (if live payments)
- [ ] All API endpoints tested end-to-end
- [ ] Token refresh flow is working
- [ ] Error handling covers all status codes
- [ ] Monitoring/analytics are configured

---

## 📚 Additional Resources

- **Full API Documentation:** `docs/FRONTEND_API_INTEGRATION.md`
- **Backend Infrastructure Guide:** `docs/INFRASTRUCTURE_COMPLETE_GUIDE.md`
- **Deployment Guide:** `docs/DEPLOYMENT.md`
- **Backend Health:** GET `/api/health/detailed`

---

**Last Updated:** April 17, 2026  
**Status:** ✅ Ready for Frontend Integration  
**Support:** Contact backend team for integration questions
