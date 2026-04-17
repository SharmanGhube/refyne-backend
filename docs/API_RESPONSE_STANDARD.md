# API Response Standard - Refyne Backend

**Version:** 1.0  
**Date:** 2026-04-18  
**Status:** All endpoints standardized  

## Overview

All Refyne API endpoints follow a **consistent, industry-standard response envelope format**. This ensures frontend developers can reliably parse all responses with a unified structure.

## Response Envelope Format

### Success Response (2xx Status Codes)

```json
{
  "success": true,
  "message": "Human-readable success message",
  "data": {
    /* endpoint-specific response data */
  },
  "request_id": "req_1234567890abcdef",
  "timestamp": "2026-04-18T10:30:45Z"
}
```

**Fields:**
- `success` (boolean): Always `true` for success responses
- `message` (string): User-friendly description of the operation result
- `data` (object/array): The actual response payload (structure varies by endpoint)
- `request_id` (string): Unique identifier for this request (for debugging/tracing)
- `timestamp` (string): ISO 8601 timestamp of the response

**HTTP Status Codes:**
- `200 OK` — Successful GET, PUT, DELETE, or other retrieval/modification
- `201 Created` — Successful resource creation (POST)
- `202 Accepted` — Async operation started (if applicable)

---

### Error Response (4xx, 5xx Status Codes)

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  },
  "details": {
    "field_name": "Field-specific error message",
    "another_field": "Another error for this field"
  },
  "request_id": "req_1234567890abcdef",
  "timestamp": "2026-04-18T10:30:45Z"
}
```

**Fields:**
- `success` (boolean): Always `false` for error responses
- `error` (object):
  - `code` (string): Machine-readable error code (e.g., `VALIDATION_ERROR`, `UNAUTHORIZED`, `NOT_FOUND`)
  - `message` (string): Human-readable error description
- `details` (object, optional): Field-level error information for validation errors
  - Key: field name (snake_case)
  - Value: specific error message for that field
- `request_id` (string): Unique identifier for this request
- `timestamp` (string): ISO 8601 timestamp of the response

**HTTP Status Codes:**
- `400 Bad Request` — Validation error, missing required fields, malformed JSON
- `401 Unauthorized` — Missing or invalid authentication
- `403 Forbidden` — Authenticated but not authorized for this resource
- `404 Not Found` — Resource doesn't exist
- `409 Conflict` — Resource already exists or state conflict
- `429 Too Many Requests` — Rate limit exceeded
- `500 Internal Server Error` — Unexpected server error

---

## Standard Error Codes

### Authentication & Authorization
- `UNAUTHORIZED` — Missing/invalid JWT token
- `INVALID_CREDENTIALS` — Wrong email/password
- `TOKEN_EXPIRED` — JWT token has expired
- `INVALID_SIGNATURE` — Webhook signature invalid

### Validation
- `VALIDATION_ERROR` — Request body failed validation
- `INVALID_REQUEST` — Malformed request format
- `INVALID_EMAIL` — Email format invalid
- `INVALID_PASSWORD` — Password doesn't meet requirements
- `DUPLICATE_EMAIL` — Email already registered
- `DUPLICATE_USERNAME` — Username already taken

###  Resource Management
- `NOT_FOUND` — Resource doesn't exist
- `ALREADY_EXISTS` — Resource already created
- `FORBIDDEN` — Access denied to resource

### Rate Limiting & Throttling
- `RATE_LIMIT_EXCEEDED` — Too many requests in time window
- `ACCOUNT_LOCKED` — Too many failed login attempts

### Payment & Subscription
- `INVALID_PRODUCT` — Product ID not found
- `NO_SUBSCRIPTION` — No active subscription for user
- `PAYMENT_FAILED` — Paddle payment processing failed

### System
- `INTERNAL_ERROR` — Unexpected server error
- `SERVICE_UNAVAILABLE` — Temporary service unavailable
- `INVALID_FORMAT` — Response/webhook format invalid

---

## Field Naming Convention

**All JSON fields use `snake_case`** (industry REST API standard):

✅ Correct:
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "is_verified": true,
  "email_address": "john@example.com",
  "token_pair": { "access_token": "...", "refresh_token": "..." },
  "created_at": "2026-04-18T10:30:00Z"
}
```

❌ Incorrect (camelCase, PascalCase):
```json
{
  "firstName": "John",
  "LastName": "Doe",
  "UserId": "123e4567-e89b-12d3-a456-426614174000",
  "isVerified": true,
  "TokenPair": { ... }
}
```

---

## Endpoint Reference by Domain

### Authentication (`/api/auth/*`)

| Endpoint | Method | Request | Response Data | Status |
|----------|--------|---------|---------------|--------|
| /register | POST | email, password, first_name, last_name, username | user object (full profile) | 201 Created |
| /request-otp | POST | email, password | { expires_in, message } | 200 OK |
| /verify-otp | POST | email, otp | { user, token_pair } | 200 OK |
| /refresh | POST | refresh_token | { token_pair } | 200 OK |
| /verify-account | POST | token | { status: "verified" } | 200 OK |
| /resend-verification | POST | email | { message } | 200 OK |
| /forgot-password | POST | email | { message } | 200 OK |
| /reset-password | POST | token, new_password | { status: "password_reset" } | 200 OK |
| /validate-reset-token | POST | token | { user_id } | 200 OK |
| /logout | POST | (auth header) | { status: "logged_out" } | 200 OK |
| /logout-all-devices | POST | (auth header) | { status: "logged_out_all_devices" } | 200 OK |

### User Management (`/api/user/*`)

| Endpoint | Method | Response Data | Status |
|----------|--------|---------------|--------|
| /profile | GET | user object | 200 OK |
| /profile | PUT | updated user object | 200 OK |
| /settings | GET | settings object | 200 OK |
| /settings | PUT | updated settings object | 200 OK |
| /onboarding | POST | { status: "completed" } | 200 OK |
| /account | DELETE | { status: "deleted" } | 200 OK |

### Workspaces (`/api/workspaces/*`)

| Endpoint | Method | Response Data | Status |
|----------|--------|---------------|--------|
| / | POST | workspace object | 201 Created |
| / | GET | { workspaces: [] } | 200 OK |
| /:id | GET | workspace object | 200 OK |
| /:id | PUT | updated workspace object | 200 OK |
| /:id | DELETE | { status: "deleted" } | 200 OK |
| /:id/members | GET | { members: [] } | 200 OK |
| /:id/members/invite | POST | { status: "invited" } | 200 OK |
| /:id/members/:memberId | DELETE | { status: "removed" } | 200 OK |

### Subscriptions (`/api/subscriptions/*`)

| Endpoint | Method | Response Data | Status |
|----------|--------|---------------|--------|
| /checkout | POST | { checkout_url, mode } | 200 OK |
| /status | GET | subscription status object | 200 OK |
| /portal | GET | { portal_url } | 200 OK |
| /webhook | POST | { event_id } | 200 OK |

---

## Common Patterns

### Validation Error Response

When a request fails validation:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed"
  },
  "details": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters",
    "first_name": "This field is required"
  },
  "request_id": "req_abc123",
  "timestamp": "2026-04-18T10:30:45Z"
}
```

**HTTP Status:** `400 Bad Request`

### Authentication Error Response

```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "User authentication required"
  },
  "request_id": "req_abc123",
  "timestamp": "2026-04-18T10:30:45Z"
}
```

**HTTP Status:** `401 Unauthorized`

### Resource Not Found Response

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Workspace not found"
  },
  "request_id": "req_abc123",
  "timestamp": "2026-04-18T10:30:45Z"
}
```

**HTTP Status:** `404 Not Found`

### Rate Limit Exceeded Response

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later."
  },
  "request_id": "req_abc123",
  "timestamp": "2026-04-18T10:30:45Z"
}
```

**HTTP Status:** `429 Too Many Requests`  
**Headers:** `Retry-After: 60` (seconds to wait before retry)

---

## Response Examples by Domain

### Authentication: Register

**Request:**
```bash
POST /api/auth/register
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Success Response (201 Created):**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "username": "john_doe",
    "is_verified": false,
    "is_active": false,
    "status": "inactive",
    "onboarding_completed": false,
    "subscription_status": "free",
    "subscription_tier": null,
    "created_at": "2026-04-18T10:30:00Z",
    "message": "User registered successfully. Please check your email to verify your account."
  },
  "request_id": "req_xyz789",
  "timestamp": "2026-04-18T10:30:45Z"
}
```

### User: Get Profile

**Request:**
```bash
GET /api/user/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "first_name": "John",
    "last_name": "Doe",
    "username": "john_doe",
    "email": "john@example.com",
    "is_verified": true,
    "is_active": true,
    "bio": "Software engineer",
    "avatar_url": "https://...",
    "created_at": "2026-04-18T10:30:00Z",
    "updated_at": "2026-04-18T11:00:00Z"
  },
  "request_id": "req_abc123",
  "timestamp": "2026-04-18T10:31:45Z"
}
```

### Workspaces: List

**Request:**
```bash
GET /api/workspaces
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Workspaces retrieved successfully",
  "data": {
    "workspaces": [
      {
        "id": "ws_001",
        "name": "Main Workspace",
        "description": "Primary workspace",
        "created_at": "2026-04-18T10:30:00Z"
      },
      {
        "id": "ws_002",
        "name": "Analytics",
        "description": "Analytics team",
        "created_at": "2026-04-18T11:00:00Z"
      }
    ]
  },
  "request_id": "req_def456",
  "timestamp": "2026-04-18T10:32:45Z"
}
```

### Subscription: Create Checkout

**Request:**
```bash
POST /api/subscriptions/checkout
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Checkout URL generated successfully",
  "data": {
    "checkout_url": "https://checkout.paddle.com/...",
    "mode": "sandbox"
  },
  "request_id": "req_ghi789",
  "timestamp": "2026-04-18T10:33:45Z"
}
```

---

## Implementation Notes for Frontend

### Parsing Success Responses

```javascript
const handleApiResponse = (response) => {
  if (response.success) {
    // Access the actual data
    const data = response.data;
    console.log(`Success: ${response.message}`, data);
  } else {
    // Handle error
    const errorCode = response.error.code;
    const errorMessage = response.error.message;
    const fieldErrors = response.details; // e.g., { "email": "Invalid format" }
    console.error(`Error [${errorCode}]: ${errorMessage}`, fieldErrors);
  }
};
```

### Handling Validation Errors

```javascript
const displayValidationErrors = (response) => {
  if (response.details) {
    // Display field-level errors
    Object.entries(response.details).forEach(([field, message]) => {
      showFieldError(field, message);
    });
  } else {
    // Display generic error
    showGenericError(response.error.message);
  }
};
```

### Retry Logic for Rate Limiting

```javascript
const fetchWithRetry = async (url, options, retries = 3) => {
  try {
    const response = await fetch(url, options);
    
    if (response.status === 429) {
      const retryAfter = response.headers.get('Retry-After') || 60;
      if (retries > 0) {
        await sleep(retryAfter * 1000);
        return fetchWithRetry(url, options, retries - 1);
      }
    }
    
    return response;
  } catch (error) {
    if (retries > 0) {
      await sleep(1000);
      return fetchWithRetry(url, options, retries - 1);
    }
    throw error;
  }
};
```

---

## Endpoint Status

### ✅ Fully Standardized Domains

- **Authentication** (11 endpoints) — All auth flows
- **User Management** (6 endpoints) — Profile, settings, onboarding, deletion
- **Workspaces** (8 endpoints) — CRUD + member management
- **Subscriptions** (4 endpoints) — Checkout, status, portal, webhooks

### Total Standardized Endpoints: 29+

All endpoints now return consistent, predictable response envelopes with:
- ✅ Uniform success/error structure
- ✅ Field-level validation error details
- ✅ Request ID for debugging
- ✅ ISO 8601 timestamps
- ✅ snake_case field naming
- ✅ Standard HTTP status codes

---

## Testing the Standard

### Health Check Endpoint

```bash
GET /health
```

Response:
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "ok",
    "version": "1.0.0"
  },
  "request_id": "req_health_123",
  "timestamp": "2026-04-18T10:34:45Z"
}
```

---

## Migration Guide

If you have existing frontend code expecting the old format:

### Old Format
```json
{
  "message": "Success",
  "data": { ... },
  "error": "Error message",
  "details": "error details"
}
```

### New Format
```json
{
  "success": true/false,
  "message": "...",
  "data": { ... },
  "error": { "code": "...", "message": "..." },
  "details": { ... },
  "request_id": "...",
  "timestamp": "..."
}
```

**Breaking Changes:**
- Added `success` boolean field (always check this first)
- Error now has `code` and `message` sub-fields
- All responses now include `request_id` and `timestamp`
- Field names are now consistently `snake_case`

---

**Last Updated:** 2026-04-18  
**Standardization Complete:** ✅ All 29+ endpoints follow industry-standard response envelope format
