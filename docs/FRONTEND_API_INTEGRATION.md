# Backend API Endpoints - Frontend Integration Guide

**Document Version:** 1.0  
**Last Updated:** November 30, 2025  
**Backend Base URL (Development):** `http://localhost:8080`  
**Backend Base URL (Ngrok):** `https://uncircled-lucca-jowly.ngrok-free.dev`

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Authentication Flow](#authentication-flow)
3. [Authentication Endpoints](#1--authentication-endpoints)
4. [Subscription Endpoints](#2--subscription-endpoints)
5. [Health Check Endpoints](#3--health-check-endpoints)
6. [Protected Endpoints](#4--protected-endpoints)
7. [Error Handling](#error-handling)
8. [Frontend Integration Examples](#frontend-integration-examples)
9. [Known Issues & Limitations](#-known-issues--limitations)

---

## Overview

### ✅ **Fully Functional & Ready for Frontend Integration**

| Module | Status | Endpoints Ready | Notes |
|--------|--------|----------------|-------|
| **Authentication** | ✅ Production Ready | 9/9 | OTP-based auth, JWT tokens, full flow tested |
| **Subscription (Paddle)** | ✅ Production Ready | 4/4 | Checkout, webhooks, status, portal working |
| **Health Checks** | ✅ Production Ready | 4/4 | Basic, detailed, readiness, liveness |

### 🚧 **Known Limitations**

| Issue | Status | Impact |
|-------|--------|--------|
| Multiple plan purchases allowed | ⚠️ Needs Fix | User can create multiple checkouts |
| No upgrade/downgrade endpoint | ⚠️ Planned | Must use Paddle portal |
| Webhook processing async | ℹ️ By Design | 1-5s delay after payment |

### ❌ **Not Implemented Yet**

- User Profile Management
- Instagram Integration
- AI/Moderation Service
- Analytics Dashboard
- Otto Chat System
- Workspace Management

---

## Authentication Flow

### 🔐 Complete Authentication Journey

```
┌─────────────┐
│   Register  │
│  New User   │
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│ Verification Email  │
│      Sent           │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  User Clicks Link   │
│  Email Verified     │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│   Request OTP       │
│  (Email Provided)   │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  OTP Email Sent     │
│  (Valid 5 minutes)  │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│   Login with OTP    │
│  Receive JWT Tokens │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Access Protected   │
│     Endpoints       │
└─────────────────────┘
       │
       ▼
┌─────────────────────┐
│ Token Expires (15m) │
│  Refresh Token      │
└─────────────────────┘
```

---

## 1. 🔐 Authentication Endpoints

### **Base Path:** `/api/auth`

---

### 1.1 Register User

**Endpoint:** `POST /api/auth/register`

**Purpose:** Create a new user account

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe"
}
```

**Password Requirements:**
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

**Success Response (201):**
```json
{
  "status": "success",
  "message": "Registration successful. Please check your email to verify your account.",
  "data": {
    "user_id": "eae12c7b-bc59-43b8-be15-616e46529723",
    "email": "user@example.com"
  }
}
```

**Error Responses:**
- `400` - Invalid request body / validation errors
- `409` - Email already exists
- `500` - Server error

**Frontend Integration:**
```typescript
const register = async (email: string, password: string, fullName: string) => {
  const response = await fetch('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ 
      email, 
      password, 
      full_name: fullName 
    })
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Registration failed');
  }
  
  return response.json();
};
```

**Next Step:** User receives verification email → Frontend shows "Check your email" message

---

### 1.2 Verify Email

**Endpoint:** `POST /api/auth/verify`

**Purpose:** Verify user's email address with token from email

**Request:**
```json
{
  "token": "verification-token-from-email-link"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Email verified successfully. You can now log in."
}
```

**Error Responses:**
- `400` - Invalid or expired token
- `404` - Token not found
- `500` - Server error

**Token Validity:** 24 hours from registration

**Frontend Integration:**
```typescript
// Extract token from URL query parameter
const verifyEmail = async (token: string) => {
  const response = await fetch('/api/auth/verify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token })
  });
  
  if (!response.ok) {
    throw new Error('Verification failed');
  }
  
  return response.json();
};

// In Next.js page or React component
useEffect(() => {
  const token = router.query.token as string;
  if (token) {
    verifyEmail(token)
      .then(() => router.push('/login?verified=true'))
      .catch(() => router.push('/login?verification_failed=true'));
  }
}, [router.query]);
```

---

### 1.3 Resend Verification Email

**Endpoint:** `POST /api/auth/resend-verification`

**Purpose:** Resend verification email to user

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Verification email sent if account exists and is unverified"
}
```

**Error Responses:**
- `400` - Invalid email format
- `429` - Too many requests (rate limited)
- `500` - Server error

**Rate Limit:** 3 requests per 15 minutes per email

---

### 1.4 Request OTP

**Endpoint:** `POST /api/auth/request-otp`

**Purpose:** Request OTP code for passwordless login

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "OTP sent to your email",
  "data": {
    "expires_in": 300
  }
}
```

**Error Responses:**
- `400` - Invalid email format
- `404` - User not found or not verified
- `429` - Too many requests (rate limited)
- `500` - Server error

**Rate Limit:** 3 requests per 15 minutes per email

**OTP Details:**
- 6-digit numeric code
- Valid for 5 minutes
- Sent via email

**Frontend Integration:**
```typescript
const requestOTP = async (email: string) => {
  const response = await fetch('/api/auth/request-otp', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });
  
  if (response.status === 429) {
    throw new Error('Too many requests. Please try again later.');
  }
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to send OTP');
  }
  
  return response.json();
};
```

**Next Step:** User receives OTP via email (valid for 5 minutes)

---

### 1.5 Login with OTP

**Endpoint:** `POST /api/auth/login`

**Purpose:** Authenticate user with email and OTP

**Request:**
```json
{
  "email": "user@example.com",
  "otp": "123456"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "user": {
      "id": "eae12c7b-bc59-43b8-be15-616e46529723",
      "email": "user@example.com",
      "full_name": "John Doe",
      "subscription_status": "free",
      "subscription_tier": null
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh-token-uuid-here",
    "expires_in": 900
  }
}
```

**Error Responses:**
- `400` - Invalid OTP format
- `401` - Invalid or expired OTP
- `404` - User not found
- `500` - Server error

**Token Information:**
- **Access Token:** JWT, valid for 15 minutes (900 seconds)
- **Refresh Token:** UUID, valid for 7 days
- **Token Version:** Stored in database, invalidated on logout

**Frontend Integration:**
```typescript
const login = async (email: string, otp: string) => {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, otp })
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Login failed');
  }
  
  const data = await response.json();
  
  // Store tokens securely
  localStorage.setItem('access_token', data.data.access_token);
  localStorage.setItem('refresh_token', data.data.refresh_token);
  
  return data;
};
```

**Security Note:** Store tokens in httpOnly cookies for production (more secure than localStorage)

---

### 1.6 Refresh Access Token

**Endpoint:** `POST /api/auth/refresh`

**Purpose:** Get new access token using refresh token

**Request:**
```json
{
  "refresh_token": "refresh-token-uuid"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900
  }
}
```

**Error Responses:**
- `401` - Invalid or expired refresh token
- `401` - Token version mismatch (user logged out)
- `500` - Server error

**Frontend Integration:**
```typescript
const refreshAccessToken = async (refreshToken: string) => {
  const response = await fetch('/api/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken })
  });
  
  if (!response.ok) {
    // Refresh token expired or invalid, redirect to login
    localStorage.clear();
    window.location.href = '/login?session_expired=true';
    throw new Error('Session expired');
  }
  
  const data = await response.json();
  localStorage.setItem('access_token', data.data.access_token);
  
  return data.data.access_token;
};

// Axios interceptor for automatic token refresh
axios.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        try {
          const newToken = await refreshAccessToken(refreshToken);
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return axios(originalRequest);
        } catch (refreshError) {
          return Promise.reject(refreshError);
        }
      }
    }
    
    return Promise.reject(error);
  }
);
```

---

### 1.7 Logout

**Endpoint:** `POST /api/auth/logout`

**Purpose:** Invalidate current session and tokens

**Headers Required:**
```
Authorization: Bearer <access_token>
```

**Request Body:** None

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Logout successful"
}
```

**Error Responses:**
- `401` - Invalid or expired access token
- `500` - Server error

**What Happens:**
- Current refresh token is deleted from database
- Token version is incremented (invalidates all access tokens)
- User must login again to get new tokens

**Frontend Integration:**
```typescript
const logout = async () => {
  const accessToken = localStorage.getItem('access_token');
  
  try {
    await fetch('/api/auth/logout', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
  } finally {
    // Clear tokens regardless of response
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    window.location.href = '/login';
  }
};
```

---

### 1.8 Logout All Devices

**Endpoint:** `POST /api/auth/logout-all`

**Purpose:** Invalidate all sessions across all devices

**Headers Required:**
```
Authorization: Bearer <access_token>
```

**Request Body:** None

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Logged out from all devices"
}
```

**Error Responses:**
- `401` - Invalid or expired access token
- `500` - Server error

**What Happens:**
- All refresh tokens are deleted from database
- Token version is incremented twice (ensures complete invalidation)
- User must login again on all devices

**Frontend Integration:**
```typescript
const logoutAllDevices = async () => {
  const accessToken = localStorage.getItem('access_token');
  
  try {
    await fetch('/api/auth/logout-all', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
  } finally {
    localStorage.clear();
    window.location.href = '/login?logged_out_all=true';
  }
};
```

---

### 1.9 Forgot Password

**Endpoint:** `POST /api/auth/forgot-password`

**Purpose:** Request password reset email

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Password reset email sent if account exists"
}
```

**Note:** Always returns success (security best practice - don't reveal if email exists)

**Error Responses:**
- `429` - Too many requests
- `500` - Server error

**Rate Limit:** 3 requests per 15 minutes per IP

**Reset Token Validity:** 1 hour

---

### 1.10 Validate Reset Token

**Endpoint:** `POST /api/auth/validate-reset-token`

**Purpose:** Check if password reset token is valid (before showing password form)

**Request:**
```json
{
  "token": "reset-token-from-email"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Token is valid"
}
```

**Error Responses:**
- `400` - Invalid or expired token
- `500` - Server error

---

### 1.11 Reset Password

**Endpoint:** `POST /api/auth/reset-password`

**Purpose:** Reset password using token from email

**Request:**
```json
{
  "token": "reset-token-from-email",
  "new_password": "NewSecurePass123!"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Password reset successfully"
}
```

**Error Responses:**
- `400` - Invalid or expired token / weak password
- `500` - Server error

**Frontend Integration:**
```typescript
// Step 1: User enters email
const forgotPassword = async (email: string) => {
  await fetch('/api/auth/forgot-password', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });
  // Always show success message
  toast.success('If account exists, reset email has been sent');
};

// Step 2: User clicks link in email, lands on reset page
// Extract token from URL and validate it
const validateToken = async (token: string) => {
  const response = await fetch('/api/auth/validate-reset-token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token })
  });
  
  if (!response.ok) {
    router.push('/forgot-password?token_invalid=true');
    throw new Error('Invalid token');
  }
};

// Step 3: User enters new password
const resetPassword = async (token: string, newPassword: string) => {
  const response = await fetch('/api/auth/reset-password', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, new_password: newPassword })
  });
  
  if (!response.ok) {
    throw new Error('Password reset failed');
  }
  
  router.push('/login?password_reset=true');
};
```

---

## 2. 💳 Subscription Endpoints

### **Base Path:** `/api/subscription`

---

### 2.1 Create Checkout Session

**Endpoint:** `POST /api/subscription/checkout`

**Purpose:** Generate Paddle checkout URL for subscription purchase

**Headers Required:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "tier": "starter"
}
```

**Valid Tiers & Pricing:**
- `starter` - $29/month
- `professional` - $79/month
- `business` - $199/month
- `enterprise` - $499/month

**Success Response (200):**
```json
{
  "status": "success",
  "data": {
    "checkout_url": "https://uncircled-lucca-jowly.ngrok-free.dev/checkout.html?_ptxn=txn_01kb7x6vwjep7aqpqtyfkmgg9v",
    "transaction_id": "txn_01kb7x6vwjep7aqpqtyfkmgg9v",
    "expires_in": 3600
  }
}
```

**Error Responses:**
- `400` - Invalid tier (must be one of: starter, professional, business, enterprise)
- `401` - Unauthorized (invalid or expired token)
- `409` - Already has active subscription (⚠️ Not implemented yet)
- `500` - Failed to create checkout with Paddle

**⚠️ Current Limitation:** User can create multiple checkouts for different plans. This will be fixed in Phase 2 (see TODO list).

**Checkout URL Details:**
- Points to embedded checkout page on backend
- Transaction valid for 1 hour
- Paddle Sandbox environment (test mode)
- Supports test cards: `4242 4242 4242 4242`

**Frontend Integration:**
```typescript
const createCheckout = async (tier: string) => {
  const accessToken = localStorage.getItem('access_token');
  
  const response = await fetch('/api/subscription/checkout', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${accessToken}`
    },
    body: JSON.stringify({ tier })
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to create checkout');
  }
  
  const data = await response.json();
  
  // Redirect user to Paddle checkout
  window.location.href = data.data.checkout_url;
};
```

**Next Step:** User completes payment on Paddle → Backend receives webhook → Database updated → User redirected to success page

---

### 2.2 Get Subscription Status

**Endpoint:** `GET /api/subscription/status`

**Purpose:** Get current user's subscription details

**Headers Required:**
```
Authorization: Bearer <access_token>
```

**Success Response (200):**
```json
{
  "status": "success",
  "data": {
    "subscription_status": "active",
    "subscription_tier": "starter",
    "paddle_customer_id": "ctm_01kb65a1234567890",
    "paddle_subscription_id": "sub_01kb65b9876543210",
    "subscription_started_at": "2025-11-29T10:30:00Z",
    "subscription_ends_at": "2025-12-29T10:30:00Z",
    "cancel_at_period_end": false
  }
}
```

**Subscription Statuses:**
- `free` - No active subscription (default)
- `active` - Active paid subscription
- `trialing` - In trial period (if trial enabled)
- `past_due` - Payment failed, subscription at risk
- `canceled` - Subscription cancelled
- `paused` - Subscription paused

**Error Responses:**
- `401` - Unauthorized (invalid or expired token)
- `500` - Server error

**Frontend Integration:**
```typescript
const getSubscriptionStatus = async () => {
  const accessToken = localStorage.getItem('access_token');
  
  const response = await fetch('/api/subscription/status', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch subscription status');
  }
  
  return response.json();
};

// Usage in React component
const SubscriptionBadge = () => {
  const [subscription, setSubscription] = useState(null);
  
  useEffect(() => {
    getSubscriptionStatus()
      .then(data => setSubscription(data.data))
      .catch(err => console.error(err));
  }, []);
  
  if (!subscription) return <Loading />;
  
  return (
    <div className={`badge ${subscription.subscription_status}`}>
      {subscription.subscription_status === 'free' ? (
        <span>Free Plan</span>
      ) : (
        <span>{subscription.subscription_tier} - {subscription.subscription_status}</span>
      )}
    </div>
  );
};
```

---

### 2.3 Get Customer Portal URL

**Endpoint:** `POST /api/subscription/portal`

**Purpose:** Generate Paddle customer portal URL for managing subscription

**Headers Required:**
```
Authorization: Bearer <access_token>
```

**Request Body:** None

**Success Response (200):**
```json
{
  "status": "success",
  "data": {
    "portal_url": "https://sandbox-checkout.paddle.com/portal/...",
    "expires_in": 3600
  }
}
```

**Error Responses:**
- `400` - No active subscription found (user must be subscribed to access portal)
- `401` - Unauthorized (invalid or expired token)
- `500` - Failed to generate portal URL

**Customer Portal Features:**
- Update payment method
- View billing history & invoices
- Change subscription plan (upgrade/downgrade)
- Cancel subscription
- Update billing address & information
- Download receipts

**Frontend Integration:**
```typescript
const openCustomerPortal = async () => {
  const accessToken = localStorage.getItem('access_token');
  
  const response = await fetch('/api/subscription/portal', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  
  if (!response.ok) {
    const error = await response.json();
    if (error.error?.includes('No active subscription')) {
      toast.error('You need an active subscription to access the portal');
    } else {
      toast.error('Failed to open customer portal');
    }
    return;
  }
  
  const data = await response.json();
  
  // Open in new tab
  window.open(data.data.portal_url, '_blank');
};

// Usage in component
<button onClick={openCustomerPortal} disabled={!hasActiveSubscription}>
  Manage Subscription
</button>
```

---

### 2.4 Paddle Webhook Handler

**Endpoint:** `POST /api/webhooks/paddle`

**Purpose:** Receive and process Paddle webhook events (internal use)

**⚠️ Not for Frontend Use** - This endpoint is called by Paddle's servers, not your frontend.

**Webhook Events Handled:**
- `transaction.completed` - Payment successful, subscription pending
- `transaction.paid` - Payment confirmed and processed
- `subscription.created` - New subscription activated
- `subscription.updated` - Subscription modified (plan change, payment method update)
- `subscription.canceled` - Subscription cancelled
- `subscription.past_due` - Payment failed, subscription at risk
- `subscription.paused` - Subscription paused
- `subscription.resumed` - Subscription resumed from pause

**What Happens When Webhook Received:**
1. Webhook signature verified (security)
2. Event type identified
3. User found by Paddle customer ID
4. Database updated with new subscription details
5. Response sent to Paddle (200 OK)

**Database Fields Updated:**
- `subscription_status`
- `subscription_tier`
- `paddle_customer_id`
- `paddle_subscription_id`
- `subscription_started_at`
- `subscription_ends_at`

**Frontend Action:** After payment, poll `/api/subscription/status` to get updated subscription:

```typescript
// On subscription success page
useEffect(() => {
  const checkSubscription = async () => {
    const status = await getSubscriptionStatus();
    
    if (status.data.subscription_status === 'active') {
      setIsSubscribed(true);
      clearInterval(pollInterval);
      router.push('/dashboard');
    }
  };
  
  // Poll every 2 seconds for up to 30 seconds
  const pollInterval = setInterval(checkSubscription, 2000);
  const timeout = setTimeout(() => {
    clearInterval(pollInterval);
    toast.error('Subscription verification timed out. Please refresh.');
  }, 30000);
  
  return () => {
    clearInterval(pollInterval);
    clearTimeout(timeout);
  };
}, []);
```

---

## 3. ❤️ Health Check Endpoints

### **Base Path:** `/api/health`

---

### 3.1 Basic Health Check

**Endpoint:** `GET /api/health`

**Purpose:** Simple health status check (for load balancers)

**Success Response (200):**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-30T10:30:00Z"
}
```

**Use Case:** Load balancer health checks, uptime monitoring

---

### 3.2 Detailed Health Check

**Endpoint:** `GET /api/health/detailed`

**Purpose:** Comprehensive system health information

**Success Response (200):**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-30T10:30:00Z",
  "version": "1.0.0",
  "environment": "development",
  "uptime": "24h30m15s",
  "checks": {
    "database": "healthy",
    "redis": "healthy",
    "river_queue": "healthy"
  }
}
```

**Component Statuses:**
- `healthy` - Component operational
- `degraded` - Component working but slow
- `unhealthy` - Component failing

**Use Case:** Debugging, monitoring dashboards, incident response

---

### 3.3 Readiness Check

**Endpoint:** `GET /api/health/ready`

**Purpose:** Check if service is ready to accept traffic

**Success Response (200):**
```json
{
  "status": "ready",
  "timestamp": "2025-11-30T10:30:00Z"
}
```

**Failure Response (503):**
```json
{
  "status": "not_ready",
  "reason": "database_not_connected"
}
```

**Use Case:** Kubernetes readiness probes, deployment verification

---

### 3.4 Liveness Check

**Endpoint:** `GET /api/health/live`

**Purpose:** Check if service is alive (can respond to requests)

**Success Response (200):**
```json
{
  "status": "alive",
  "timestamp": "2025-11-30T10:30:00Z"
}
```

**Use Case:** Kubernetes liveness probes, automatic restart triggers

---

## 4. 👤 Protected Endpoints

### 4.1 Get Current User Profile

**Endpoint:** `GET /api/protected/me`

**Purpose:** Get authenticated user's complete profile information

**Headers Required:**
```
Authorization: Bearer <access_token>
```

**Success Response (200):**
```json
{
  "message": "Authentication successful",
  "user_id": "eae12c7b-bc59-43b8-be15-616e46529723",
  "email": "user@example.com",
  "username": "john_doe",
  "request_id": "req_abc123xyz"
}
```

**Error Responses:**
- `401` - Unauthorized (invalid/expired token)
- `500` - Server error

**Frontend Integration:**
```typescript
const getCurrentUser = async () => {
  const accessToken = localStorage.getItem('access_token');
  
  const response = await fetch('/api/protected/me', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch user profile');
  }
  
  return response.json();
};

// Usage in layout/navbar
const Navbar = () => {
  const [user, setUser] = useState(null);
  
  useEffect(() => {
    getCurrentUser()
      .then(data => setUser(data))
      .catch(err => {
        console.error(err);
        router.push('/login');
      });
  }, []);
  
  return (
    <nav>
      <span>Welcome, {user?.email}</span>
    </nav>
  );
};
```

---

## Complete Subscription Flow

### 🛒 End-to-End Purchase Journey

```
1. User Authentication
   ├─ User logs in with OTP
   ├─ Receives access_token & refresh_token
   └─ Can access protected endpoints

2. Browse Plans
   ├─ GET /api/subscription/status (check current status)
   ├─ Display tier options on frontend
   └─ User selects "Professional - $79/month"

3. Create Checkout
   ├─ POST /api/subscription/checkout { tier: "professional" }
   ├─ Backend creates Paddle transaction
   ├─ Returns checkout_url
   └─ Frontend redirects to checkout_url

4. Complete Payment (on Paddle page)
   ├─ User enters payment details
   ├─ Paddle processes payment
   ├─ Paddle sends webhook to backend
   └─ User redirected to success page

5. Webhook Processing (backend, async)
   ├─ Backend receives transaction.paid webhook
   ├─ Verifies webhook signature
   ├─ Updates user in database:
   │  ├─ subscription_status = "active"
   │  ├─ subscription_tier = "professional"
   │  ├─ paddle_customer_id = "ctm_xxx"
   │  └─ paddle_subscription_id = "sub_xxx"
   └─ Sends 200 OK to Paddle

6. Verify Subscription (frontend polls)
   ├─ GET /api/subscription/status (every 2 seconds)
   ├─ Check if subscription_status === "active"
   ├─ Stop polling when active
   └─ Redirect to dashboard

7. Access Subscribed Features
   ├─ All protected endpoints now authorized
   └─ Features unlocked based on tier
```

### Integration Code Example:

```tsx
// pages/pricing.tsx
const PricingPage = () => {
  const handleSelectPlan = async (tier: string) => {
    try {
      const data = await createCheckout(tier);
      // Redirect to Paddle checkout
      window.location.href = data.data.checkout_url;
    } catch (error) {
      toast.error('Failed to create checkout');
    }
  };
  
  return (
    <div className="pricing-grid">
      <PlanCard
        name="Starter"
        price="$29/month"
        onSelect={() => handleSelectPlan('starter')}
      />
      <PlanCard
        name="Professional"
        price="$79/month"
        onSelect={() => handleSelectPlan('professional')}
      />
      {/* ... more plans */}
    </div>
  );
};

// pages/subscription-success.tsx
const SubscriptionSuccessPage = () => {
  const [loading, setLoading] = useState(true);
  const [subscription, setSubscription] = useState(null);
  const router = useRouter();
  
  useEffect(() => {
    let pollCount = 0;
    const maxPolls = 15; // 30 seconds (2s interval)
    
    const checkSubscription = async () => {
      try {
        const status = await getSubscriptionStatus();
        
        if (status.data.subscription_status === 'active') {
          setSubscription(status.data);
          setLoading(false);
          
          setTimeout(() => {
            router.push('/dashboard');
          }, 3000);
        } else {
          pollCount++;
          if (pollCount >= maxPolls) {
            setLoading(false);
            toast.error('Subscription verification timed out');
          }
        }
      } catch (error) {
        console.error(error);
      }
    };
    
    // Initial check
    checkSubscription();
    
    // Poll every 2 seconds
    const interval = setInterval(checkSubscription, 2000);
    
    return () => clearInterval(interval);
  }, []);
  
  if (loading) {
    return (
      <div className="loading">
        <Spinner />
        <p>Verifying your subscription...</p>
      </div>
    );
  }
  
  return (
    <div className="success">
      <CheckIcon />
      <h1>Welcome to {subscription.subscription_tier}!</h1>
      <p>Your subscription is now active.</p>
      <p>Redirecting to dashboard...</p>
    </div>
  );
};
```

---

## Error Handling

### Standard Error Response Format

All error responses follow this structure:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Specific validation error"
  }
}
```

### Common HTTP Status Codes

| Status | Meaning | Frontend Action |
|--------|---------|----------------|
| `200` | Success | Process response data |
| `201` | Created | Resource created successfully |
| `400` | Bad Request | Show validation errors to user |
| `401` | Unauthorized | Redirect to login / refresh token |
| `403` | Forbidden | Show "Access Denied" message |
| `404` | Not Found | Show "Resource not found" |
| `409` | Conflict | Show conflict message (e.g., email exists) |
| `429` | Rate Limited | Show "Too many requests" with retry timer |
| `500` | Server Error | Show generic error + log to monitoring |
| `503` | Service Unavailable | Show "Service temporarily unavailable" |

### Frontend Error Handler

```typescript
class APIError extends Error {
  constructor(
    public statusCode: number,
    message: string,
    public code?: string,
    public details?: Record<string, any>
  ) {
    super(message);
    this.name = 'APIError';
  }
}

const handleAPIError = async (response: Response): Promise<never> => {
  const contentType = response.headers.get('content-type');
  
  if (contentType?.includes('application/json')) {
    const error = await response.json();
    throw new APIError(
      response.status,
      error.error || `HTTP ${response.status} Error`,
      error.code,
      error.details
    );
  }
  
  throw new APIError(response.status, `HTTP ${response.status} Error`);
};

// Usage in API calls
const makeAPIRequest = async (url: string, options?: RequestInit) => {
  try {
    const response = await fetch(url, options);
    
    if (!response.ok) {
      await handleAPIError(response);
    }
    
    return response.json();
  } catch (error) {
    if (error instanceof APIError) {
      // Handle specific error codes
      switch (error.statusCode) {
        case 401:
          // Try to refresh token
          const refreshToken = localStorage.getItem('refresh_token');
          if (refreshToken) {
            try {
              await refreshAccessToken(refreshToken);
              // Retry original request
              return makeAPIRequest(url, options);
            } catch {
              // Refresh failed, redirect to login
              localStorage.clear();
              window.location.href = '/login?session_expired=true';
            }
          }
          break;
          
        case 429:
          toast.error('Too many requests. Please try again later.');
          break;
          
        case 500:
          toast.error('Server error. Please try again.');
          // Log to error tracking service
          console.error('Server error:', error);
          break;
          
        default:
          toast.error(error.message);
      }
    }
    
    throw error;
  }
};
```

### Validation Error Handling

```typescript
// Example: Registration form validation errors
try {
  await register(email, password, fullName);
} catch (error) {
  if (error instanceof APIError && error.statusCode === 400) {
    // Display field-specific errors
    if (error.details) {
      Object.entries(error.details).forEach(([field, message]) => {
        setFieldError(field, message as string);
      });
    } else {
      toast.error(error.message);
    }
  }
}
```

---

## Frontend Integration Examples

### Complete React Hook for Authentication

```typescript
// hooks/useAuth.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  id: string;
  email: string;
  full_name: string;
  subscription_status: string;
  subscription_tier: string | null;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  
  // Actions
  register: (email: string, password: string, fullName: string) => Promise<void>;
  requestOTP: (email: string) => Promise<void>;
  login: (email: string, otp: string) => Promise<void>;
  logout: () => Promise<void>;
  logoutAllDevices: () => Promise<void>;
  refreshAccessToken: () => Promise<void>;
  getCurrentUser: () => Promise<void>;
  forgotPassword: (email: string) => Promise<void>;
  resetPassword: (token: string, newPassword: string) => Promise<void>;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const useAuth = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      
      register: async (email: string, password: string, fullName: string) => {
        const response = await fetch(`${API_BASE}/api/auth/register`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password, full_name: fullName })
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || 'Registration failed');
        }
        
        return response.json();
      },
      
      requestOTP: async (email: string) => {
        const response = await fetch(`${API_BASE}/api/auth/request-otp`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email })
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || 'Failed to send OTP');
        }
        
        return response.json();
      },
      
      login: async (email: string, otp: string) => {
        const response = await fetch(`${API_BASE}/api/auth/login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, otp })
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || 'Login failed');
        }
        
        const data = await response.json();
        
        set({
          user: data.data.user,
          accessToken: data.data.access_token,
          refreshToken: data.data.refresh_token,
          isAuthenticated: true
        });
      },
      
      logout: async () => {
        const { accessToken } = get();
        
        try {
          await fetch(`${API_BASE}/api/auth/logout`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${accessToken}`
            }
          });
        } finally {
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false
          });
        }
      },
      
      logoutAllDevices: async () => {
        const { accessToken } = get();
        
        try {
          await fetch(`${API_BASE}/api/auth/logout-all`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${accessToken}`
            }
          });
        } finally {
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false
          });
        }
      },
      
      refreshAccessToken: async () => {
        const { refreshToken } = get();
        
        if (!refreshToken) {
          throw new Error('No refresh token available');
        }
        
        const response = await fetch(`${API_BASE}/api/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken })
        });
        
        if (!response.ok) {
          // Refresh token invalid, clear auth state
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false
          });
          throw new Error('Token refresh failed');
        }
        
        const data = await response.json();
        set({ accessToken: data.data.access_token });
      },
      
      getCurrentUser: async () => {
        const { accessToken } = get();
        
        const response = await fetch(`${API_BASE}/api/protected/me`, {
          headers: {
            'Authorization': `Bearer ${accessToken}`
          }
        });
        
        if (!response.ok) {
          throw new Error('Failed to fetch user');
        }
        
        const data = await response.json();
        set({ user: data });
      },
      
      forgotPassword: async (email: string) => {
        const response = await fetch(`${API_BASE}/api/auth/forgot-password`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email })
        });
        
        if (!response.ok) {
          throw new Error('Failed to send reset email');
        }
        
        return response.json();
      },
      
      resetPassword: async (token: string, newPassword: string) => {
        const response = await fetch(`${API_BASE}/api/auth/reset-password`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ token, new_password: newPassword })
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || 'Password reset failed');
        }
        
        return response.json();
      }
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken
      })
    }
  )
);
```

### Axios Instance with Auto Token Refresh

```typescript
// lib/api.ts
import axios from 'axios';
import { useAuth } from '@/hooks/useAuth';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json'
  }
});

// Request interceptor - add auth token
api.interceptors.request.use(
  (config) => {
    const token = useAuth.getState().accessToken;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor - handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    // If 401 and haven't retried yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        // Attempt to refresh token
        await useAuth.getState().refreshAccessToken();
        const newToken = useAuth.getState().accessToken;
        
        // Retry original request with new token
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        return api(originalRequest);
      } catch (refreshError) {
        // Refresh failed, logout and redirect
        useAuth.getState().logout();
        window.location.href = '/login?session_expired=true';
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);

export default api;
```

### Protected Route Component

```typescript
// components/ProtectedRoute.tsx
import { useEffect } from 'react';
import { useRouter } from 'next/router';
import { useAuth } from '@/hooks/useAuth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireSubscription?: boolean;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requireSubscription = false
}) => {
  const router = useRouter();
  const { isAuthenticated, user } = useAuth();
  
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login?redirect=' + router.pathname);
      return;
    }
    
    if (requireSubscription && user?.subscription_status !== 'active') {
      router.push('/pricing?upgrade_required=true');
      return;
    }
  }, [isAuthenticated, user, requireSubscription]);
  
  if (!isAuthenticated) {
    return <div>Loading...</div>;
  }
  
  if (requireSubscription && user?.subscription_status !== 'active') {
    return <div>Subscription required...</div>;
  }
  
  return <>{children}</>;
};

// Usage
// pages/dashboard.tsx
export default function DashboardPage() {
  return (
    <ProtectedRoute requireSubscription>
      <Dashboard />
    </ProtectedRoute>
  );
}
```

---

## 🚨 Known Issues & Limitations

### 1. **Multiple Subscription Purchase** ⚠️ HIGH PRIORITY
**Issue:** User can create multiple checkout sessions for different plans without checking existing subscription.

**Impact:** 
- User could accidentally purchase multiple plans
- Billing confusion
- Support burden

**Current Workaround:** Frontend should check subscription status before showing checkout:

```tsx
const handleSubscribe = async (tier: string) => {
  // Check if user already has active subscription
  const status = await getSubscriptionStatus();
  
  if (status.data.subscription_status === 'active') {
    toast.error('You already have an active subscription');
    toast.info('Use the customer portal to change your plan');
    return;
  }
  
  // Proceed with checkout
  await createCheckout(tier);
};
```

**Permanent Fix:** Backend will add validation in Phase 2 (ETA: 1 hour)
- Block checkout creation if active subscription exists
- Return 409 Conflict with helpful message
- Direct user to customer portal for plan changes

---

### 2. **No Upgrade/Downgrade Endpoint** ⚠️ MEDIUM PRIORITY
**Issue:** No dedicated API endpoint for changing subscription plans directly.

**Impact:**
- Users must cancel and repurchase
- Loss of billing history continuity
- Poor UX for plan changes

**Current Workaround:** Direct users to Paddle customer portal:

```tsx
const changePlan = async () => {
  toast.info('Opening customer portal to change your plan...');
  await openCustomerPortal();
};
```

**Permanent Fix:** Planned for Phase 3 (ETA: 2-3 hours)
- New endpoint: `POST /api/subscription/change-plan`
- Supports upgrade/downgrade
- Handles proration automatically
- Maintains billing continuity

---

### 3. **Webhook Processing is Async** ℹ️ BY DESIGN
**Issue:** After payment, database update happens via webhook (1-5 second delay).

**Impact:**
- User redirected to success page before subscription activated
- Need to poll for updated status

**Solution:** Implement polling on success page:

```tsx
// pages/subscription-success.tsx
useEffect(() => {
  let attempts = 0;
  const maxAttempts = 15; // 30 seconds total
  
  const pollSubscription = async () => {
    const status = await getSubscriptionStatus();
    
    if (status.data.subscription_status === 'active') {
      setSubscriptionActive(true);
      return true;
    }
    
    attempts++;
    if (attempts >= maxAttempts) {
      toast.error('Verification timed out. Please refresh the page.');
      return false;
    }
    
    return false;
  };
  
  // Initial check
  pollSubscription();
  
  // Poll every 2 seconds
  const interval = setInterval(async () => {
    const done = await pollSubscription();
    if (done) clearInterval(interval);
  }, 2000);
  
  return () => clearInterval(interval);
}, []);
```

**Why Not Change It:**
- Webhook approach is industry standard (Stripe, Paddle, PayPal all use this)
- More reliable than inline API calls during checkout
- Handles network issues gracefully
- Separates payment processing from business logic

---

### 4. **Rate Limiting Not Visible in Responses** ℹ️ ENHANCEMENT
**Issue:** Rate limit headers not exposed to help frontend show countdown.

**Current State:**
- Rate limits exist (3 requests per 15 minutes)
- No `X-RateLimit-*` headers in response
- Frontend can't show "Try again in X minutes"

**Workaround:** Show generic message:

```tsx
catch (error) {
  if (error.statusCode === 429) {
    toast.error('Too many requests. Please try again in 15 minutes.');
  }
}
```

**Enhancement:** Add rate limit headers (Phase 7):
```
X-RateLimit-Limit: 3
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1701345600
```

---

### 5. **No Idempotency Keys for Checkout** ⚠️ MEDIUM PRIORITY
**Issue:** Duplicate checkout button clicks create multiple transactions.

**Impact:**
- Network issues can cause duplicate transactions
- User confusion about which checkout to use

**Workaround:** Disable button after click:

```tsx
const [loading, setLoading] = useState(false);

const handleCheckout = async () => {
  if (loading) return; // Prevent double click
  
  setLoading(true);
  try {
    await createCheckout(tier);
  } finally {
    setLoading(false);
  }
};

<button onClick={handleCheckout} disabled={loading}>
  {loading ? 'Creating checkout...' : 'Subscribe'}
</button>
```

**Permanent Fix:** Planned for Phase 5 (ETA: 2 hours)
- Add optional `idempotency_key` to checkout request
- Backend stores and checks keys
- Returns existing checkout if key matches

---

### 6. **Token Storage in localStorage** ⚠️ SECURITY CONCERN
**Current State:** Frontend examples use localStorage for tokens.

**Security Issue:**
- Vulnerable to XSS attacks
- Tokens accessible to JavaScript

**Recommendation for Production:**
```typescript
// Use httpOnly cookies instead (more secure)
// Backend needs to set cookies in auth responses

// In backend response headers:
Set-Cookie: access_token=xxx; HttpOnly; Secure; SameSite=Strict; Max-Age=900
Set-Cookie: refresh_token=xxx; HttpOnly; Secure; SameSite=Strict; Max-Age=604800

// Frontend makes requests with credentials
fetch('/api/protected/me', {
  credentials: 'include' // Sends cookies automatically
});
```

**Status:** Backend ready to support cookies, frontend needs to implement

---

## 📞 Next Steps & Support

### For Frontend Developers

**Immediate Actions:**
1. ✅ Implement authentication flow (register → verify → login)
2. ✅ Add token refresh logic (Axios interceptor)
3. ✅ Create protected route wrapper
4. ✅ Implement subscription checkout flow
5. ⚠️ Add subscription status polling after payment
6. ⚠️ Implement frontend validation for existing subscriptions

**Before Production:**
1. ⚠️ Switch from localStorage to httpOnly cookies
2. ⚠️ Add rate limit handling with countdown timers
3. ⚠️ Implement proper loading states
4. ⚠️ Add error boundary components
5. ⚠️ Set up error tracking (Sentry, etc.)

### Backend Improvements in Progress

**Phase 2 (Next 1-2 days):**
- ✅ Prevent multiple subscription purchases
- ✅ Add upgrade/downgrade endpoint
- ✅ Enhance webhook handling

**Phase 5 (Next week):**
- ✅ Add idempotency keys
- ✅ Add rate limit headers
- ✅ Implement request deduplication

### Testing Credentials

**Test Card (Paddle Sandbox):**
```
Card Number: 4242 4242 4242 4242
Expiry: Any future date
CVV: Any 3 digits
ZIP: Any valid ZIP
```

**Test OTP:**
- Check server logs if testing locally
- Check email if using real email service

### API Endpoints Summary

| Endpoint | Method | Auth Required | Ready | Notes |
|----------|--------|---------------|-------|-------|
| `/api/auth/register` | POST | ❌ | ✅ | Creates user, sends verification email |
| `/api/auth/verify` | POST | ❌ | ✅ | Verifies email with token |
| `/api/auth/request-otp` | POST | ❌ | ✅ | Sends OTP to email |
| `/api/auth/login` | POST | ❌ | ✅ | Authenticates with OTP |
| `/api/auth/refresh` | POST | ❌ | ✅ | Refreshes access token |
| `/api/auth/logout` | POST | ✅ | ✅ | Logs out current session |
| `/api/auth/logout-all` | POST | ✅ | ✅ | Logs out all sessions |
| `/api/auth/forgot-password` | POST | ❌ | ✅ | Sends password reset email |
| `/api/auth/reset-password` | POST | ❌ | ✅ | Resets password with token |
| `/api/subscription/checkout` | POST | ✅ | ✅ | Creates Paddle checkout |
| `/api/subscription/status` | GET | ✅ | ✅ | Gets subscription details |
| `/api/subscription/portal` | POST | ✅ | ✅ | Opens customer portal |
| `/api/webhooks/paddle` | POST | ❌ | ✅ | Handles Paddle webhooks (internal) |
| `/api/protected/me` | GET | ✅ | ✅ | Gets current user profile |
| `/api/health` | GET | ❌ | ✅ | Basic health check |
| `/api/health/detailed` | GET | ❌ | ✅ | Detailed health check |
| `/api/health/ready` | GET | ❌ | ✅ | Readiness probe |
| `/api/health/live` | GET | ❌ | ✅ | Liveness probe |

### Contact & Support

**Backend Developer:** Available for integration support

**Questions?** Check:
1. This documentation
2. Server logs at `http://localhost:8080` or ngrok URL
3. Postman collection (request from backend team)

**Found a Bug?** Include:
- Request method and endpoint
- Request body/headers
- Response status and body
- Expected vs actual behavior

---

**Document Last Updated:** November 30, 2025  
**Backend Version:** 1.0.0  
**Ready for Frontend Integration:** ✅ **YES** (with noted limitations)  
**Postman Collection:** Available on request  
**Support:** Backend team standing by for integration questions
