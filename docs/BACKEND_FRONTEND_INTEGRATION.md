# Backend-Frontend API Integration Guide

This document explains how to configure the frontend to communicate with the Refyne backend API, including authentication, request/response handling, error management, and deployment integration.

**Last Updated:** 2026-04-18  
**API Version:** v1  
**Backend Status:** Production Ready (Railway)

---

## Table of Contents
1. [API Base Configuration](#api-base-configuration)
2. [Environment Setup](#environment-setup)
3. [Authentication & JWT](#authentication--jwt)
4. [HTTP Client Setup](#http-client-setup)
5. [Request/Response Patterns](#requestresponse-patterns)
6. [Error Handling](#error-handling)
7. [CORS Configuration](#cors-configuration)
8. [API Endpoints Reference](#api-endpoints-reference)
9. [State Management](#state-management)
10. [Real-Time Communication](#real-time-communication)
11. [Testing & Mocking](#testing--mocking)
12. [Deployment Integration](#deployment-integration)
13. [Troubleshooting](#troubleshooting)

---

## Quick Reference

### Production Backend URL
```
https://refyne-backend-production.up.railway.app
```

### Test the Backend (CLI)
```bash
# Check if backend is running
curl https://refyne-backend-production.up.railway.app/health

# Should respond with:
# {"status":"ok"}
```

### Frontend Environment Variables (Copy & Paste)

**For Production:**
```env
VITE_API_BASE_URL=https://refyne-backend-production.up.railway.app
```

**For Local Development:**
```env
VITE_API_BASE_URL=http://localhost:8080
```

---

## API Base Configuration

> **🚀 PRODUCTION URL:** `https://refyne-backend-production.up.railway.app`
> 
> This is the live backend API running on Railway. Use this URL for production frontend deployments.

### Development Environment

**Backend Running Locally:**
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT=30000
```

**Backend Running on Railway (Dev):**
```env
VITE_API_BASE_URL=https://dev-api.refyne.railway.app
VITE_API_TIMEOUT=30000
```

**Production Environment:**
```env
VITE_API_BASE_URL=https://refyne-backend-production.up.railway.app
VITE_API_TIMEOUT=30000
```

### API Endpoints Structure

All endpoints follow the pattern:
```
{BASE_URL}/api/{domain}/{resource}
```

Examples:
```
# Development (Local)
GET    http://localhost:8080/api/auth/login
POST   http://localhost:8080/api/auth/register

# Production (Railway)
GET    https://refyne-backend-production.up.railway.app/api/auth/login
POST   https://refyne-backend-production.up.railway.app/api/auth/register
GET    https://refyne-backend-production.up.railway.app/api/user/profile
POST   https://refyne-backend-production.up.railway.app/api/workspaces
```

---

## Environment Setup

### Frontend `.env.development` (Local Development)
```env
# API Configuration
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT=30000

# Environment
VITE_ENV=development
VITE_APP_NAME=Refyne
VITE_APP_VERSION=1.0.0

# Instagram OAuth (Sandbox)
VITE_INSTAGRAM_CLIENT_ID=your-sandbox-instagram-app-id
VITE_INSTAGRAM_REDIRECT_URI=http://localhost:3000/instagram/callback

# Paddle (Sandbox)
VITE_PADDLE_CLIENT_TOKEN=your-sandbox-paddle-token
VITE_PAYMENT_MODE=sandbox

# Analytics & Monitoring (optional)
VITE_SENTRY_DSN=https://key@sentry.io/project
VITE_POSTHOG_KEY=your-posthog-key
```

### Frontend `.env.production` (Production Build)
```env
# API Configuration
VITE_API_BASE_URL=https://refyne-backend-production.up.railway.app
VITE_API_TIMEOUT=30000

# Environment
VITE_ENV=production
VITE_APP_NAME=Refyne
VITE_APP_VERSION=1.0.0

# Instagram OAuth (Production)
VITE_INSTAGRAM_CLIENT_ID=your-production-instagram-app-id
VITE_INSTAGRAM_REDIRECT_URI=https://app.refyne.io/instagram/callback

# Paddle (Production)
VITE_PADDLE_CLIENT_TOKEN=your-production-paddle-token
VITE_PAYMENT_MODE=live

# Analytics & Monitoring
VITE_SENTRY_DSN=https://key@sentry.io/project
VITE_POSTHOG_KEY=your-posthog-key
```

### Backend `.env` (For Reference)
```bash
# The backend needs these for CORS to work with frontend:
CORS_ORIGINS=http://localhost:3000,http://localhost:3001,https://app.refyne.io,https://web.refyne.io

# On Railway, this is auto-configured via environment variable
```

---

## Authentication & JWT

### JWT Token Flow

```
1. User logs in
   Frontend: POST /api/auth/login
   ↓
2. Backend returns tokens
   {
     "access_token": "eyJhbGc...",  # 15-minute expiry
     "refresh_token": "eyJhbGc...", # 7-day expiry
     "user": { ... }
   }
   ↓
3. Frontend stores tokens
   - Access token: Secure HTTP-only cookie OR secure session storage
   - Refresh token: Secure HTTP-only cookie (preferred)
   ↓
4. Frontend includes access token in all requests
   Authorization: Bearer {access_token}
   ↓
5. Backend validates token
   ↓
6. When access token expires (15min)
   Frontend: POST /api/auth/refresh
   ↓
7. Backend returns new access token
   ↓
8. Frontend updates token and retries original request
```

### Token Storage Strategy

**Option 1: Secure Cookies (RECOMMENDED)**
```javascript
// Backend sets HttpOnly cookie
Set-Cookie: access_token={token}; HttpOnly; Secure; SameSite=Strict; Max-Age=900

// Frontend automatically includes cookie in requests (no manual handling needed)
// Fetch/Axios with credentials: 'include'
```

**Option 2: Session Storage (Less Secure)**
```javascript
// Store in sessionStorage (cleared on browser close)
sessionStorage.setItem('access_token', token);
sessionStorage.setItem('refresh_token', token);

// Add to every request manually
headers: {
  'Authorization': `Bearer ${sessionStorage.getItem('access_token')}`
}
```

**Option 3: Local Storage (Not Recommended)**
```javascript
// Vulnerable to XSS attacks - avoid if possible
localStorage.setItem('access_token', token);
```

**Recommended Implementation:**
```javascript
// Use secure HTTP-only cookies set by backend
// Configure fetch/axios to include credentials:

// Fetch
fetch(url, {
  credentials: 'include'  // Include cookies in request
})

// Axios
axios.defaults.withCredentials = true  // Include cookies
```

### Token Refresh Mechanism

```javascript
// Axios Interceptor - Auto-refresh on 401
api.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        // Refresh token
        await api.post('/api/auth/refresh');
        
        // Retry original request
        return api(originalRequest);
      } catch (refreshError) {
        // Refresh failed - redirect to login
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);
```

### Logout & Token Blacklist

```javascript
// On logout
async function logout() {
  try {
    // Notify backend to blacklist token
    await api.post('/api/auth/logout');
    
    // Clear local state
    sessionStorage.removeItem('user');
    
    // Clear cookies (browser handles HttpOnly cookies)
    
    // Redirect to login
    window.location.href = '/login';
  } catch (error) {
    console.error('Logout failed:', error);
    // Force redirect anyway
    window.location.href = '/login';
  }
}
```

---

## HTTP Client Setup

### Axios Configuration

**`src/services/api.ts` (or `apiClient.js`)**

```typescript
import axios, { AxiosInstance, AxiosError } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
const API_TIMEOUT = import.meta.env.VITE_API_TIMEOUT || 30000;

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT,
  withCredentials: true,  // Include cookies
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
});

// Request interceptor - Add auth header if needed
api.interceptors.request.use(
  (config) => {
    // Add any request-wide headers here
    config.headers['X-Request-ID'] = generateRequestId();
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - Handle tokens, errors, etc.
api.interceptors.response.use(
  (response) => {
    // Success - return data
    return response.data;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as any;
    
    // Handle 401 Unauthorized - try to refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        // Attempt token refresh
        await api.post('/api/auth/refresh');
        
        // Retry original request
        return api(originalRequest);
      } catch (refreshError) {
        // Refresh failed - logout user
        handleAuthError();
        return Promise.reject(refreshError);
      }
    }
    
    // Handle other errors
    handleAPIError(error);
    
    return Promise.reject(error);
  }
);

export default api;
```

### Fetch API Alternative

**`src/services/api.ts` (using Fetch)**

```typescript
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

interface FetchOptions extends RequestInit {
  timeout?: number;
}

async function fetchAPI(
  endpoint: string,
  options: FetchOptions = {}
): Promise<any> {
  const { timeout = 30000, ...fetchOptions } = options;
  
  // Build URL
  const url = `${API_BASE_URL}${endpoint}`;
  
  // Set default headers
  const headers = {
    'Content-Type': 'application/json',
    ...fetchOptions.headers,
  };
  
  // Create abort controller for timeout
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);
  
  try {
    const response = await fetch(url, {
      ...fetchOptions,
      headers,
      credentials: 'include', // Include cookies
      signal: controller.signal,
    });
    
    clearTimeout(timeoutId);
    
    // Parse response
    const data = await response.json();
    
    // Check if successful
    if (!response.ok) {
      // Handle specific status codes
      if (response.status === 401) {
        // Try to refresh token
        const refreshed = await refreshToken();
        if (refreshed) {
          // Retry original request
          return fetchAPI(endpoint, options);
        }
        // Refresh failed - redirect to login
        handleAuthError();
      }
      
      // Throw error
      throw new APIError(data.message || 'API request failed', response.status, data);
    }
    
    return data;
  } catch (error) {
    clearTimeout(timeoutId);
    
    if (error instanceof Error && error.name === 'AbortError') {
      throw new Error('Request timeout');
    }
    
    throw error;
  }
}

export { fetchAPI };
```

---

## Request/Response Patterns

### Standard Response Format

All API responses follow this envelope format:

**Success Response (200, 201):**
```json
{
  "success": true,
  "code": 200,
  "message": "Success",
  "data": {
    "user_id": "abc123",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
  },
  "error": null,
  "meta": {
    "timestamp": "2026-04-18T10:30:00Z",
    "request_id": "req-xyz789"
  }
}
```

**Error Response (400, 401, 500, etc.):**
```json
{
  "success": false,
  "code": 400,
  "message": "Bad Request",
  "data": null,
  "error": {
    "field": "email",
    "message": "Invalid email format",
    "code": "INVALID_EMAIL"
  },
  "meta": {
    "timestamp": "2026-04-18T10:30:00Z",
    "request_id": "req-xyz789"
  }
}
```

### Request Examples

**Login Request:**
```javascript
// Request
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}

// Response (Success)
{
  "success": true,
  "code": 200,
  "message": "Login successful",
  "data": {
    "user_id": "abc123",
    "email": "user@example.com",
    "first_name": "John",
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  },
  "error": null,
  "meta": { ... }
}

// Response (Error)
{
  "success": false,
  "code": 401,
  "message": "Unauthorized",
  "data": null,
  "error": {
    "message": "Invalid credentials"
  },
  "meta": { ... }
}
```

**Create Workspace Request:**
```javascript
// Request
POST /api/workspaces
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "My Workspace",
  "description": "Team workspace",
  "icon": "default"
}

// Response (Success)
{
  "success": true,
  "code": 201,
  "message": "Workspace created",
  "data": {
    "workspace_id": "ws-123",
    "name": "My Workspace",
    "owner_id": "user-456",
    "created_at": "2026-04-18T10:30:00Z"
  },
  "error": null,
  "meta": { ... }
}
```

**List Workspaces Request:**
```javascript
// Request with pagination
GET /api/workspaces?page=1&limit=10&sort=created_at

// Response
{
  "success": true,
  "code": 200,
  "message": "Success",
  "data": [
    {
      "workspace_id": "ws-123",
      "name": "Workspace 1",
      "owner_id": "user-456",
      "members_count": 3
    },
    {
      "workspace_id": "ws-456",
      "name": "Workspace 2",
      "owner_id": "user-456",
      "members_count": 1
    }
  ],
  "error": null,
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 2,
      "total_pages": 1
    }
  }
}
```

### Frontend Type Definitions

**TypeScript Interfaces:**

```typescript
// API Response Envelope
interface APIResponse<T> {
  success: boolean;
  code: number;
  message: string;
  data: T | null;
  error: APIError | null;
  meta: {
    timestamp: string;
    request_id: string;
    pagination?: {
      page: number;
      limit: number;
      total: number;
      total_pages: number;
    };
  };
}

// Common Error Type
interface APIError {
  field?: string;
  message: string;
  code?: string;
}

// User Type
interface User {
  user_id: string;
  email: string;
  first_name: string;
  last_name: string;
  username: string;
  avatar_url?: string;
  onboarding_completed: boolean;
  created_at: string;
}

// Workspace Type
interface Workspace {
  workspace_id: string;
  name: string;
  description: string;
  owner_id: string;
  members_count: number;
  created_at: string;
}

// Auth Response
interface AuthResponse {
  user_id: string;
  email: string;
  first_name: string;
  access_token: string;
  refresh_token: string;
}
```

---

## Error Handling

### Error Types & Status Codes

| Status | Type | Handling |
|--------|------|----------|
| **400** | Bad Request | Show validation errors to user |
| **401** | Unauthorized | Attempt token refresh, redirect to login if failed |
| **403** | Forbidden | Show "Access Denied" message |
| **404** | Not Found | Show "Resource not found" |
| **429** | Rate Limited | Show "Too many requests" with retry hint |
| **500** | Server Error | Show generic error, log to monitoring |
| **503** | Service Unavailable | Show "Service temporarily unavailable" |

### Error Handling Implementation

**Global Error Handler:**

```typescript
// src/services/errorHandler.ts

import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';

interface ErrorResponse {
  success: false;
  code: number;
  message: string;
  error: {
    field?: string;
    message: string;
    code?: string;
  };
}

export function handleAPIError(error: any): void {
  const authStore = useAuthStore();
  const notificationStore = useNotificationStore();
  
  // Network error (no response from server)
  if (!error.response) {
    notificationStore.addError('Network error. Please check your connection.');
    return;
  }
  
  const status = error.response.status;
  const data: ErrorResponse = error.response.data;
  
  switch (status) {
    case 400:
      // Validation error - show field-specific error
      if (data.error?.field) {
        notificationStore.addError(`${data.error.field}: ${data.error.message}`);
      } else {
        notificationStore.addError(data.error?.message || 'Invalid request');
      }
      break;
      
    case 401:
      // Unauthorized - redirect to login
      authStore.logout();
      window.location.href = '/login?redirect=' + encodeURIComponent(window.location.pathname);
      break;
      
    case 403:
      notificationStore.addError('You do not have permission to perform this action.');
      break;
      
    case 404:
      notificationStore.addError('Resource not found.');
      break;
      
    case 429:
      notificationStore.addWarning('Too many requests. Please wait a moment before trying again.');
      break;
      
    case 500:
    case 503:
      notificationStore.addError('Server error. Please try again later or contact support.');
      // Log to error tracking (Sentry)
      logErrorToMonitoring(error);
      break;
      
    default:
      notificationStore.addError(data.message || 'An unexpected error occurred.');
  }
}

function logErrorToMonitoring(error: any): void {
  // Integrate with Sentry or similar
  if (window.Sentry) {
    window.Sentry.captureException(error);
  }
}
```

**Component-Level Error Handling:**

```typescript
// In a React component
async function handleFormSubmit(formData) {
  try {
    setLoading(true);
    
    const response = await api.post('/api/auth/login', formData);
    
    // Success
    authStore.setUser(response.data);
    navigate('/dashboard');
    
  } catch (error) {
    // Error is already handled by global handler
    // Show validation-specific errors if needed
    if (error.response?.status === 400) {
      setFormErrors({
        [error.response.data.error.field]: error.response.data.error.message
      });
    }
  } finally {
    setLoading(false);
  }
}
```

---

## CORS Configuration

### Backend CORS Setup (Go/Gin)

The backend is configured in `cmd/main.go` with CORS middleware:

```go
// CORS configuration
corsConfig := cors.Config{
  AllowOrigins:     []string{os.Getenv("CORS_ORIGINS")}, // Comma-separated
  AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
  AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
  ExposeHeaders:    []string{"Content-Length"},
  AllowCredentials: true,
  MaxAge:           12 * time.Hour,
}
router.Use(cors.New(corsConfig))
```

**Backend Environment Variable:**
```env
# .env or Railway
CORS_ORIGINS=http://localhost:3000,http://localhost:3001,https://app.refyne.io,https://web.refyne.io,https://refyne-backend-production.up.railway.app
```

### Frontend CORS Setup (When Issues Occur)

**Development - Local Testing:**
```javascript
// If you get CORS errors locally, ensure:
// 1. Backend is running: make run (on port 8080)
// 2. Frontend calls correct API URL: http://localhost:8080
// 3. Include credentials in requests:

fetch('http://localhost:8080/api/auth/login', {
  method: 'POST',
  credentials: 'include',  // Important for cookies
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(data)
});
```

**Production - Railway:**
```javascript
// Production CORS is automatically configured
// Frontend at https://app.refyne.io
// Backend at https://api.refyne.io
// Both configured in backend CORS_ORIGINS
```

### Common CORS Issues & Solutions

**Issue: "Access to XMLHttpRequest blocked by CORS policy"**

**Solution:**
1. Check backend has frontend URL in `CORS_ORIGINS`
2. Check request includes `credentials: 'include'`
3. Check `Content-Type` header is allowed
4. Verify backend sends `Access-Control-Allow-Origin` header

```bash
# Debug CORS in browser console
curl -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -X OPTIONS http://localhost:8080/api/auth/login -v
```

---

## API Endpoints Reference

### Authentication Endpoints

```
POST /api/auth/register
  Body: { first_name, last_name, username, email, password }
  Response: { user_id, email }

POST /api/auth/login
  Body: { email, password }
  Response: { user_id, email, access_token, refresh_token }

POST /api/auth/otp/send
  Body: { email }
  Response: { message: "OTP sent" }

POST /api/auth/otp/verify
  Body: { email, otp }
  Response: { user_id, email, access_token, refresh_token }

POST /api/auth/refresh
  Headers: { Authorization: Bearer {refresh_token} }
  Response: { access_token }

POST /api/auth/logout
  Headers: { Authorization: Bearer {access_token} }
  Response: { message: "Logged out" }

POST /api/auth/password/reset/request
  Body: { email }
  Response: { message: "Reset email sent" }

POST /api/auth/password/reset/confirm
  Body: { token, new_password }
  Response: { message: "Password reset successful" }
```

### User Endpoints

```
GET /api/user/profile
  Headers: { Authorization: Bearer {access_token} }
  Response: { user_id, email, first_name, last_name, ... }

PUT /api/user/profile
  Headers: { Authorization: Bearer {access_token} }
  Body: { first_name, last_name, username, bio, ... }
  Response: { user_id, email, ... }

GET /api/user/settings
  Headers: { Authorization: Bearer {access_token} }
  Response: { language, timezone, email_notifications, ... }

PUT /api/user/settings
  Headers: { Authorization: Bearer {access_token} }
  Body: { language, timezone, email_notifications, ... }
  Response: { language, timezone, ... }

POST /api/user/onboarding/complete
  Headers: { Authorization: Bearer {access_token} }
  Response: { message: "Onboarding completed" }

DELETE /api/user/account
  Headers: { Authorization: Bearer {access_token} }
  Response: { message: "Account deleted" }
```

### Workspace Endpoints

```
POST /api/workspaces
  Headers: { Authorization: Bearer {access_token} }
  Body: { name, description }
  Response: { workspace_id, name, owner_id, ... }

GET /api/workspaces
  Headers: { Authorization: Bearer {access_token} }
  Query: ?page=1&limit=10&sort=created_at
  Response: [ { workspace_id, name, ... } ]

GET /api/workspaces/{id}
  Headers: { Authorization: Bearer {access_token} }
  Response: { workspace_id, name, owner_id, ... }

PUT /api/workspaces/{id}
  Headers: { Authorization: Bearer {access_token} }
  Body: { name, description }
  Response: { workspace_id, name, ... }

DELETE /api/workspaces/{id}
  Headers: { Authorization: Bearer {access_token} }
  Response: { message: "Workspace deleted" }

GET /api/workspaces/{id}/members
  Headers: { Authorization: Bearer {access_token} }
  Response: [ { user_id, email, role, ... } ]

POST /api/workspaces/{id}/members
  Headers: { Authorization: Bearer {access_token} }
  Body: { email, role }
  Response: { message: "Invitation sent" }

DELETE /api/workspaces/{id}/members/{user_id}
  Headers: { Authorization: Bearer {access_token} }
  Response: { message: "Member removed" }
```

### Instagram Endpoints

```
GET /api/instagram/auth/url
  Headers: { Authorization: Bearer {access_token} }
  Response: { auth_url: "https://instagram.com/oauth/..." }

POST /api/instagram/auth/callback
  Headers: { Authorization: Bearer {access_token} }
  Body: { code }
  Response: { message: "Account connected", account_id, ... }

GET /api/instagram/media
  Headers: { Authorization: Bearer {access_token} }
  Query: ?page=1&limit=20&sort=posted_at
  Response: [ { media_id, caption, likes, comments, ... } ]

GET /api/instagram/media/{id}
  Headers: { Authorization: Bearer {access_token} }
  Response: { media_id, caption, engagement_stats, comments, ... }

GET /api/instagram/analytics
  Headers: { Authorization: Bearer {access_token} }
  Query: ?start_date=2026-04-01&end_date=2026-04-18
  Response: { follower_growth, engagement_rate, top_posts, ... }

POST /api/instagram/media/sync
  Headers: { Authorization: Bearer {access_token} }
  Response: { message: "Sync started", synced_count: 150 }
```

### AI Assistant (Otto) Endpoints

```
POST /api/otto/conversations
  Headers: { Authorization: Bearer {access_token} }
  Body: { title }
  Response: { conversation_id, title, created_at }

GET /api/otto/conversations
  Headers: { Authorization: Bearer {access_token} }
  Query: ?page=1&limit=20
  Response: [ { conversation_id, title, last_message, ... } ]

GET /api/otto/conversations/{id}
  Headers: { Authorization: Bearer {access_token} }
  Response: { conversation_id, title, created_at, ... }

POST /api/otto/conversations/{id}/messages
  Headers: { Authorization: Bearer {access_token} }
  Body: { content, context_ids }
  Response: { message_id, content, timestamp, ... }

GET /api/otto/conversations/{id}/messages
  Headers: { Authorization: Bearer {access_token} }
  Query: ?page=1&limit=50
  Response: [ { message_id, content, sender, timestamp, ... } ]

POST /api/otto/conversations/{id}/feedback
  Headers: { Authorization: Bearer {access_token} }
  Body: { message_id, rating, comment }
  Response: { message: "Feedback saved" }
```

### Subscription Endpoints

```
GET /api/subscription/status
  Headers: { Authorization: Bearer {access_token} }
  Response: { status, plan, next_billing_date, ... }

POST /api/subscription/checkout
  Headers: { Authorization: Bearer {access_token} }
  Body: { plan_id, billing_cycle }
  Response: { checkout_url: "https://checkout.paddle.com/..." }

POST /api/subscription/cancel
  Headers: { Authorization: Bearer {access_token} }
  Response: { message: "Subscription canceled" }

POST /api/subscription/webhooks/paddle
  Body: { event_type, ... }
  Headers: { X-Paddle-Signature: ... }
  Response: { message: "Webhook processed" }
```

### Health & Monitoring Endpoints

```
GET /health
  Response: { status: "ok" }

GET /health/detailed
  Response: { status, database: ok, redis: ok, ... }

GET /health/ready
  Response: { ready: true }

GET /health/live
  Response: { alive: true }

GET /metrics
  Response: Prometheus format metrics
  Content-Type: text/plain
```

---

## State Management

### Recommended Architecture (Redux/Zustand Example)

**Stores to Create:**

1. **Auth Store** - User authentication
   ```typescript
   interface AuthState {
     user: User | null;
     isAuthenticated: boolean;
     loading: boolean;
     error: string | null;
     
     // Actions
     login: (email, password) => Promise<void>;
     register: (data) => Promise<void>;
     logout: () => Promise<void>;
     refreshToken: () => Promise<void>;
     setUser: (user) => void;
   }
   ```

2. **Workspace Store** - Workspace management
   ```typescript
   interface WorkspaceState {
     workspaces: Workspace[];
     activeWorkspace: Workspace | null;
     members: WorkspaceMember[];
     loading: boolean;
     
     // Actions
     fetchWorkspaces: () => Promise<void>;
     createWorkspace: (data) => Promise<Workspace>;
     updateWorkspace: (id, data) => Promise<void>;
     deleteWorkspace: (id) => Promise<void>;
     fetchMembers: (workspaceId) => Promise<void>;
     inviteMember: (workspaceId, email) => Promise<void>;
   }
   ```

3. **Instagram Store** - Instagram data
   ```typescript
   interface InstagramState {
     connected: boolean;
     account: InstagramAccount | null;
     media: InstagramMedia[];
     analytics: InstagramAnalytics | null;
     loading: boolean;
     
     // Actions
     connectAccount: (code) => Promise<void>;
     disconnectAccount: () => Promise<void>;
     fetchMedia: () => Promise<void>;
     fetchAnalytics: (dateRange) => Promise<void>;
     syncMedia: () => Promise<void>;
   }
   ```

4. **Notification Store** - UI notifications & toasts
   ```typescript
   interface NotificationState {
     notifications: Notification[];
     
     // Actions
     addNotification: (message, type) => void;
     addError: (message) => void;
     addSuccess: (message) => void;
     removeNotification: (id) => void;
     clearAll: () => void;
   }
   ```

### Zustand Example Implementation

```typescript
// src/stores/authStore.ts
import { create } from 'zustand';
import api from '@/services/api';

interface AuthStore {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setUser: (user: User) => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  isAuthenticated: false,
  loading: false,
  error: null,
  
  login: async (email, password) => {
    set({ loading: true, error: null });
    try {
      const response = await api.post('/api/auth/login', { email, password });
      set({
        user: response.data,
        isAuthenticated: true,
        loading: false,
      });
    } catch (error) {
      set({
        error: error.message || 'Login failed',
        loading: false,
      });
      throw error;
    }
  },
  
  logout: async () => {
    try {
      await api.post('/api/auth/logout');
    } finally {
      set({
        user: null,
        isAuthenticated: false,
      });
    }
  },
  
  setUser: (user) => {
    set({ user, isAuthenticated: true });
  },
  
  clearError: () => {
    set({ error: null });
  },
}));
```

---

## Real-Time Communication

### WebSocket Setup (for Real-Time Features)

**If implementing real-time chat/notifications:**

```typescript
// src/services/websocket.ts
import { useAuthStore } from '@/stores/authStore';

interface WebSocketMessage {
  type: 'message' | 'notification' | 'typing' | 'presence';
  data: any;
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  
  constructor() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    this.url = `${protocol}//${import.meta.env.VITE_API_BASE_URL.replace(/^https?:\/\//, '')}/ws`;
  }
  
  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(`${this.url}?token=${token}`);
        
        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          resolve();
        };
        
        this.ws.onmessage = (event) => {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.handleMessage(message);
        };
        
        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };
        
        this.ws.onclose = () => {
          console.log('WebSocket disconnected');
          this.attemptReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }
  
  send(message: WebSocketMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }
  
  private handleMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case 'message':
        // Handle incoming message
        break;
      case 'notification':
        // Handle notification
        break;
      case 'typing':
        // Handle typing indicator
        break;
      case 'presence':
        // Handle user presence
        break;
    }
  }
  
  private attemptReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.pow(2, this.reconnectAttempts) * 1000;
      
      setTimeout(() => {
        const authStore = useAuthStore();
        if (authStore.user) {
          this.connect(sessionStorage.getItem('access_token') || '');
        }
      }, delay);
    }
  }
  
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
    }
  }
}

export const wsClient = new WebSocketClient();
```

---

## Testing & Mocking

### API Mocking for Development

**Using MSW (Mock Service Worker):**

```typescript
// src/mocks/handlers.ts
import { http, HttpResponse } from 'msw';

export const handlers = [
  http.post('/api/auth/login', async ({ request }) => {
    return HttpResponse.json({
      success: true,
      code: 200,
      data: {
        user_id: 'user-123',
        email: 'user@example.com',
        access_token: 'mock-token',
      },
    });
  }),
  
  http.get('/api/workspaces', () => {
    return HttpResponse.json({
      success: true,
      code: 200,
      data: [
        {
          workspace_id: 'ws-123',
          name: 'Test Workspace',
          members_count: 3,
        },
      ],
    });
  }),
];
```

### Integration Testing

```typescript
// tests/auth.integration.test.ts
import { render, screen, userEvent } from '@testing-library/react';
import { LoginPage } from '@/pages/LoginPage';

describe('Authentication Integration', () => {
  test('user can login with valid credentials', async () => {
    render(<LoginPage />);
    
    // Fill form
    await userEvent.type(screen.getByPlaceholderText('Email'), 'user@example.com');
    await userEvent.type(screen.getByPlaceholderText('Password'), 'password123');
    
    // Submit
    await userEvent.click(screen.getByRole('button', { name: /login/i }));
    
    // Verify redirect
    await expect(screen.findByText(/dashboard/i)).toBeInTheDocument();
  });
});
```

---

## Deployment Integration

### Development (Local)

**Start Backend:**
```bash
cd /path/to/refyne-backend
make run
# Backend running on http://localhost:8080
```

**Start Frontend:**
```bash
cd /path/to/refyne-frontend
npm install
npm run dev
# Frontend running on http://localhost:3000
```

**Verify Connection:**
```bash
# Test API endpoint
curl http://localhost:8080/health
# Should respond with { "status": "ok" }
```

### Staging (Railway Dev)

**Frontend `.env.staging`:**
```env
VITE_API_BASE_URL=https://dev-api.refyne.railway.app
VITE_ENV=staging
```

**Deploy Frontend:**
```bash
npm run build:staging
# Deploy build to staging hosting (Vercel, Netlify, etc.)
```

### Production (Railway)

**Frontend `.env.production`:**
```env
VITE_API_BASE_URL=https://api.refyne.io
VITE_ENV=production
```

**Backend Railway Configuration:**
```env
# Automatically set on Railway
CORS_ORIGINS=https://app.refyne.io,https://web.refyne.io
DATABASE_URL=postgresql://...
REDIS_HOST=...
```

**Deployment:**
```bash
# Frontend to Vercel/Netlify
npm run build
npm run deploy

# Backend auto-deploys from main branch on GitHub
# (GitHub Actions → Railway)
```

### Environment Variable Mapping

| Environment | Backend | Frontend |
|-------------|---------|----------|
| Local | `http://localhost:8080` | `http://localhost:3000` |
| Dev (Railway) | `https://dev-api.refyne.railway.app` | `https://dev.refyne.io` |
| Production | `https://refyne-backend-production.up.railway.app` | `https://app.refyne.io` |

---

## Troubleshooting

### 1. "Cannot connect to API"

**Solution:**
```bash
# Check backend is running
curl http://localhost:8080/health

# Check API URL in frontend
console.log(import.meta.env.VITE_API_BASE_URL)

# Check network tab in browser dev tools
# Verify request URL and headers
```

### 2. "CORS Error"

**Solution:**
```bash
# Verify backend has frontend URL in CORS_ORIGINS
echo $CORS_ORIGINS

# Add frontend URL if missing
export CORS_ORIGINS="http://localhost:3000,$CORS_ORIGINS"

# Restart backend
make run
```

### 3. "401 Unauthorized"

**Solution:**
```javascript
// Check if token is being sent
console.log('Token:', sessionStorage.getItem('access_token'));

// Check if credentials: 'include' is set
// Check if HttpOnly cookie is being stored

// Try logging in again
// Check token format starts with "eyJ" (base64)
```

### 4. "Token Refresh Loop"

**Solution:**
```javascript
// Check refresh endpoint returns new token
// Check token expiry times are correct
// Check _retry flag prevents infinite loops

// Clear tokens and login again
sessionStorage.clear();
window.location.href = '/login';
```

### 5. "Slow API Responses"

**Solution:**
```bash
# Check backend performance
curl -w "@curl-format.txt" http://localhost:8080/api/user/profile

# Check database performance
# Check Redis connection
# Check network latency

# Increase timeout in frontend
VITE_API_TIMEOUT=60000
```

### 6. "Rate Limit Errors (429)"

**Solution:**
```javascript
// Expected behavior - backend limits to 100 req/min
// Wait before retrying
// Batch requests when possible
// Check frontend for duplicate requests

// Implement exponential backoff retry
```

---

## Quick Integration Checklist

- [ ] Backend running and accessible
- [ ] Frontend `.env` configured with correct API URL
- [ ] CORS configured in backend
- [ ] HTTP client (Axios/Fetch) setup with interceptors
- [ ] JWT token storage configured
- [ ] Token refresh mechanism implemented
- [ ] Auth middleware redirects to login on 401
- [ ] Error handling catches and displays errors
- [ ] State management stores created
- [ ] API endpoints tested with Postman/Insomnia
- [ ] TypeScript types defined for API responses
- [ ] Loading states show spinners/skeletons
- [ ] Success/error notifications display
- [ ] Login page connects to backend
- [ ] Dashboard loads user workspaces
- [ ] Logout clears tokens and redirects
- [ ] Token refresh works on 401
- [ ] WebSocket configured (if real-time needed)
- [ ] E2E tests pass locally
- [ ] Staging deployment works
- [ ] Production deployment works

---

## Additional Resources

### Documentation
- Backend API: See `/docs/PROJECT_OVERVIEW.md`
- Frontend Routes: See `/docs/FRONTEND_ROUTES_SPEC.md`
- Deployment: See `/docs/DEPLOYMENT.md`

### Tools
- **Postman:** Import API collection for testing endpoints
- **Insomnia:** REST client for API testing
- **VS Code REST Client:** Use `.http` files for requests
- **Browser DevTools:** Network tab for debugging CORS/requests

### Monitoring
- **Frontend Errors:** Sentry integration
- **Backend Logs:** Railway dashboard
- **Metrics:** Prometheus + Grafana Cloud
- **Performance:** Lighthouse, Web Vitals

---

**Document Version:** 1.0  
**Last Updated:** 2026-04-18  
**Status:** Complete & Production Ready
