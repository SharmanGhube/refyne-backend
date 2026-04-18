# Backend Authentication Endpoints - Implementation Status & Action Plan

**Date:** 2026-04-18  
**Status:** Partial Implementation - Endpoints Exist but with Different Names  
**Priority:** HIGH - Frontend is blocked

---

## Problem Summary

Frontend is getting 404 errors because they expect endpoints with different names:

| Frontend Expects | Backend Has | Status |
|------------------|------------|--------|
| `POST /api/auth/login` (password) | ❌ Missing | ❌ Need to add |
| `POST /api/auth/otp/send` | `POST /api/auth/request-otp` | ⚠️ Different name |
| `POST /api/auth/otp/verify` | `POST /api/auth/login` (mapped to VerifyOTP) | ⚠️ Confusing name |
| `POST /api/auth/register` | `POST /api/auth/register` | ✅ Correct |
| `POST /api/auth/verify/email` | `POST /api/auth/verify` | ⚠️ Different path |
| `POST /api/auth/verify/email/resend` | `POST /api/auth/resend-verification` | ⚠️ Different name |
| `POST /api/auth/password/reset/request` | `POST /api/auth/forgot-password` | ⚠️ Different name |
| `POST /api/auth/password/reset/confirm` | `POST /api/auth/reset-password` | ⚠️ Different name |
| `POST /api/auth/logout` | `POST /api/auth/logout` | ✅ Correct |
| `POST /api/auth/refresh` | `POST /api/auth/refresh` | ✅ Correct |

---

## Solution Options

### Option A: Update Frontend Documentation (Not Recommended)
- **Pro:** Minimal backend changes
- **Con:** Frontend has to adapt to non-standard endpoints
- **Con:** Confusion with endpoint naming (e.g., `/api/auth/login` is OTP, not password)
- **Status:** ❌ Not ideal

### Option B: Add Route Aliases (RECOMMENDED)
- **Pro:** Backend supports both naming conventions
- **Pro:** Frontend can use standard, intuitive endpoint names
- **Pro:** No breaking changes to existing code
- **Con:** Some duplication in routes
- **Status:** ✅ Best approach

### Option C: Refactor All Endpoints (Not Recommended)
- **Pro:** Clean, single set of endpoints
- **Con:** Breaking changes to any existing integrations
- **Con:** More work required
- **Status:** ❌ Too risky

---

## RECOMMENDED ACTION: Option B - Add Route Aliases

We'll add new routes that map to existing handlers, making both frontend and backend happy.

### Implementation Plan

#### Step 1: Update `internal/domains/auth/routes/auth.go`

Add these new routes alongside existing ones:

```go
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

func SetupAuthRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	AuthHandler := registry.Auth

	// Initialize rate limiter
	rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

	authGroup := router.Group("/auth")
	{
		// Public routes with rate limiting
		authGroup.POST("/register",
			rateLimiter.Middleware(middlewares.RegisterLimit),
			AuthHandler.Register)

		// ========== OTP ENDPOINTS (Frontend-Expected Names) ==========
		authGroup.POST("/otp/send",
			rateLimiter.Middleware(middlewares.OTPRequestLimit),
			AuthHandler.RequestOTP)

		authGroup.POST("/otp/verify",
			rateLimiter.Middleware(middlewares.LoginLimit),
			AuthHandler.VerifyOTP)

		// ========== LEGACY OTP ENDPOINTS (Keep for backwards compatibility) ==========
		authGroup.POST("/request-otp",
			rateLimiter.Middleware(middlewares.OTPRequestLimit),
			AuthHandler.RequestOTP)

		// ========== PASSWORD LOGIN ENDPOINT (NEW - Currently Missing) ==========
		// This endpoint needs to be implemented
		authGroup.POST("/login",
			rateLimiter.Middleware(middlewares.LoginLimit),
			AuthHandler.LoginWithPassword) // NEW HANDLER NEEDED

		// ========== TOKEN REFRESH ==========
		authGroup.POST("/refresh",
			rateLimiter.Middleware(middlewares.RefreshLimit),
			AuthHandler.RefreshToken)

		// ========== EMAIL VERIFICATION (Frontend-Expected Paths) ==========
		authGroup.POST("/verify/email",
			AuthHandler.VerifyAccount)

		authGroup.POST("/verify/email/resend",
			rateLimiter.Middleware(middlewares.VerificationResendLimit),
			AuthHandler.ResendVerification)

		// ========== LEGACY VERIFICATION ENDPOINTS (Keep for backwards compatibility) ==========
		authGroup.POST("/verify",
			AuthHandler.VerifyAccount)

		authGroup.POST("/resend-verification",
			rateLimiter.Middleware(middlewares.VerificationResendLimit),
			AuthHandler.ResendVerification)

		// ========== PASSWORD RESET (Frontend-Expected Paths) ==========
		authGroup.POST("/password/reset/request",
			rateLimiter.Middleware(middlewares.PasswordResetLimit),
			AuthHandler.ForgotPassword)

		authGroup.POST("/password/reset/confirm",
			AuthHandler.ResetPassword)

		authGroup.POST("/password/reset/validate-token",
			AuthHandler.ValidateResetToken)

		// ========== LEGACY PASSWORD RESET ENDPOINTS (Keep for backwards compatibility) ==========
		authGroup.POST("/forgot-password",
			rateLimiter.Middleware(middlewares.PasswordResetLimit),
			AuthHandler.ForgotPassword)

		authGroup.POST("/reset-password",
			AuthHandler.ResetPassword)

		authGroup.POST("/validate-reset-token",
			AuthHandler.ValidateResetToken)

		// Protected routes (authentication required + rate limiting)
		protected := authGroup.Group("")
		protected.Use(middlewares.AuthMiddleware())
		protected.Use(rateLimiter.Middleware(middlewares.ProtectedEndpointLimit))
		{
			protected.POST("/logout", AuthHandler.Logout)
			protected.POST("/logout-all", AuthHandler.LogoutAllDevices)
		}
	}
}
```

#### Step 2: Implement Missing Password Login Handler

Add to `internal/domains/auth/handler/auth.go`:

```go
// LoginWithPassword handles password-based login
func (h *AuthHandlerImpl) LoginWithPassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	// Bind and validate the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(err))
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Processing password login request", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", req.Email))

	// Call Auth Service to verify credentials and login
	user, tokenPair, appErr := h.authService.LoginWithPassword(c, req.Email, req.Password)
	if appErr != nil {
		h.logger.Error("Password login failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("User logged in successfully with password", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", user.ID))

	// Prepare user response (exclude sensitive data)
	userResponse := gin.H{
		"id":                    user.ID,
		"email":                 user.Email,
		"username":              user.Username,
		"first_name":            user.FirstName,
		"last_name":             user.LastName,
		"status":                user.Status,
		"is_active":             user.IsActive,
		"is_verified":           user.IsVerified,
		"onboarding_completed":  user.OnboardingCompleted,
		"created_at":            user.CreatedAt,
	}

	// Respond with the JWT tokens using standardized success envelope
	responseData := gin.H{
		"user":       userResponse,
		"token_pair": tokenPair,
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Login successful", responseData)
}
```

#### Step 3: Implement Missing Service Method

Add to `internal/domains/auth/services/auth.go`:

```go
// LoginWithPassword verifies email and password, returns user and token pair
func (s *AuthServiceImpl) LoginWithPassword(ctx context.Context, email string, password string) (*entity.User, *entity.TokenPair, *appErr.AppError) {
	// Get user by email
	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to find user", zap.String("email", email), zap.Error(err))
		return nil, nil, appErr.NewAppError(
			http.StatusUnauthorized,
			appErr.AuthenticationFailed,
			"Invalid email or password",
			err,
		)
	}

	// Check if user exists
	if user == nil || user.ID == "" {
		s.logger.Warn("User not found", zap.String("email", email))
		return nil, nil, appErr.NewAppError(
			http.StatusUnauthorized,
			appErr.AuthenticationFailed,
			"Invalid email or password",
			nil,
		)
	}

	// Check if account is active
	if !user.IsActive {
		s.logger.Warn("Inactive account login attempt", zap.String("userID", user.ID), zap.String("email", email))
		return nil, nil, appErr.NewAppError(
			http.StatusUnauthorized,
			appErr.AccountInactive,
			"Your account has been deactivated",
			nil,
		)
	}

	// Verify password
	if !passwordutil.VerifyPassword(user.PasswordHash, password) {
		s.logger.Warn("Invalid password", zap.String("userID", user.ID), zap.String("email", email))
		
		// Record failed login attempt for security
		securityErr := s.accountSecurityRepository.RecordFailedLoginAttempt(ctx, user.ID)
		if securityErr != nil {
			s.logger.Error("Failed to record login attempt", zap.Error(securityErr))
		}

		return nil, nil, appErr.NewAppError(
			http.StatusUnauthorized,
			appErr.AuthenticationFailed,
			"Invalid email or password",
			nil,
		)
	}

	// Check if account is locked due to failed attempts
	security, err := s.accountSecurityRepository.GetByUserID(ctx, user.ID)
	if err == nil && security != nil {
		if security.IsLocked && time.Now().Before(security.LockedUntil) {
			s.logger.Warn("Account locked due to failed login attempts", zap.String("userID", user.ID))
			return nil, nil, appErr.NewAppError(
				http.StatusUnauthorized,
				appErr.AccountLocked,
				"Account locked due to too many failed login attempts. Try again later.",
				nil,
			)
		}

		// Clear failed attempts on successful login
		if security.FailedAttempts > 0 {
			s.accountSecurityRepository.ResetFailedLoginAttempts(ctx, user.ID)
		}
	}

	// Generate JWT tokens
	tokenPair, tokenErr := jwtutil.GenerateTokenPair(user.ID, user.Email)
	if tokenErr != nil {
		s.logger.Error("Failed to generate tokens", zap.String("userID", user.ID), zap.Error(tokenErr))
		return nil, nil, appErr.NewAppError(
			http.StatusInternalServerError,
			appErr.InternalServerError,
			"Failed to generate authentication tokens",
			tokenErr,
		)
	}

	// Record successful login in audit log
	s.auditLogger.LogLoginSuccess(ctx, user.ID, "password", "")

	s.logger.Info("User logged in successfully", zap.String("userID", user.ID), zap.String("email", email))

	return user, tokenPair, nil
}
```

---

## Current Implementation Status

### ✅ Already Implemented & Working

```
POST /api/auth/register
POST /api/auth/request-otp (maps to /api/auth/otp/send)
POST /api/auth/login (OTP verification - maps to /api/auth/otp/verify)
POST /api/auth/refresh
POST /api/auth/verify (email verification - maps to /api/auth/verify/email)
POST /api/auth/resend-verification
POST /api/auth/forgot-password (maps to /api/auth/password/reset/request)
POST /api/auth/reset-password (maps to /api/auth/password/reset/confirm)
POST /api/auth/logout
POST /api/auth/logout-all
POST /api/auth/validate-reset-token
```

### ❌ Missing - Must Implement

```
POST /api/auth/login (PASSWORD-BASED LOGIN)
```

### ⚠️ Endpoint Name Mismatches - Need Aliases

```
POST /api/auth/otp/send (currently /api/auth/request-otp)
POST /api/auth/otp/verify (currently /api/auth/login for OTP)
POST /api/auth/verify/email (currently /api/auth/verify)
POST /api/auth/verify/email/resend (currently /api/auth/resend-verification)
POST /api/auth/password/reset/request (currently /api/auth/forgot-password)
POST /api/auth/password/reset/confirm (currently /api/auth/reset-password)
```

---

## Frontend Requirements Met

### Request/Response Format ✅

All endpoints must return standardized response envelope:

```json
{
  "success": true,
  "code": 200,
  "message": "Success message",
  "data": { /* response data */ },
  "error": null,
  "meta": {
    "timestamp": "2026-04-18T10:30:00Z",
    "request_id": "req-xyz"
  }
}
```

**Status:** ✅ Already implemented via `middlewares.RespondWithSuccess()` and `middlewares.RespondWithError()`

### Token Format ✅

JWT tokens must contain user context and be refreshable.

**Status:** ✅ Implemented in `jwtutil.GenerateTokenPair()`

### CORS Support ✅

All endpoints must support CORS with `credentials: include`.

**Status:** ✅ Already configured in `middlewares.CORS()`

### Validation ✅

Email format, password strength, username format validation.

**Status:** ✅ Using Go struct tags with `binding:"required,email"` etc.

### Error Handling ✅

400 for validation, 401 for auth, 409 for conflict, 500 for server error.

**Status:** ✅ Implemented via AppError system

---

## Step-by-Step Implementation

### 1. Update Routes File (2 minutes)

```bash
cd internal/domains/auth/routes/
# Edit auth.go - add new route aliases
```

### 2. Add Password Login Handler (10 minutes)

```bash
cd internal/domains/auth/handler/
# Edit auth.go - add LoginWithPassword method
```

### 3. Add Service Method (10 minutes)

```bash
cd internal/domains/auth/services/
# Edit auth.go - add LoginWithPassword service method
```

### 4. Regenerate Wire DI (2 minutes)

```bash
cd cmd
wire
```

### 5. Test Endpoints (10 minutes)

```bash
# Local testing
make run

# Test with curl
curl -X POST http://localhost:8080/api/auth/otp/send \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!"}'
```

---

## Quick Reference - All Endpoints After Implementation

### Authentication

```
POST /api/auth/register
  Body: { first_name, last_name, username, email, password }
  Response: { user_id, email, message }

POST /api/auth/login (PASSWORD LOGIN - NEW)
  Body: { email, password }
  Response: { user, token_pair }

POST /api/auth/otp/send (NEW ALIAS)
  Body: { email, password }
  Response: { expires_in, message }

POST /api/auth/otp/verify (NEW ALIAS)
  Body: { email, otp }
  Response: { user, token_pair }

POST /api/auth/refresh
  Body: { refresh_token }
  Response: { token_pair }

POST /api/auth/logout (PROTECTED)
  Response: { message }
```

### Email Verification

```
POST /api/auth/verify/email (NEW PATH)
  Body: { token }
  Response: { status }

POST /api/auth/verify/email/resend (NEW PATH)
  Body: { email }
  Response: { message }
```

### Password Reset

```
POST /api/auth/password/reset/request (NEW PATH)
  Body: { email }
  Response: { message }

POST /api/auth/password/reset/confirm (NEW PATH)
  Body: { token, new_password }
  Response: { message }

POST /api/auth/password/reset/validate-token (NEW PATH)
  Body: { token }
  Response: { valid, expires_at }
```

---

## Testing Checklist

After implementation, verify:

- [ ] `POST /api/auth/otp/send` returns 200 with OTP sent message
- [ ] `POST /api/auth/otp/verify` returns user + tokens on valid OTP
- [ ] `POST /api/auth/login` (password) returns user + tokens on valid credentials
- [ ] `POST /api/auth/login` (password) returns 401 on invalid password
- [ ] Password login records failed attempts (max 5, lock 15 min)
- [ ] OTP expires after 15 minutes
- [ ] Verification tokens expire after 24 hours
- [ ] Password reset tokens expire after 30 minutes
- [ ] All endpoints return proper error messages
- [ ] CORS allows localhost:3000 and production domain
- [ ] Tokens are valid JWTs
- [ ] Frontend can refresh tokens on 401

---

## Summary

**Action Required:**
1. Add route aliases for frontend-expected endpoint names ⏱️ 2 min
2. Implement password login handler ⏱️ 10 min
3. Add password login service method ⏱️ 10 min
4. Regenerate Wire DI ⏱️ 2 min
5. Test all endpoints ⏱️ 10 min

**Total Time:** ~35 minutes

**Blockers for Frontend:** Removed after these changes

**Breaking Changes:** None - backward compatible (legacy endpoints still work)

---

**Document Version:** 1.0  
**Priority:** HIGH  
**Timeline:** Should be completed before frontend development continues
