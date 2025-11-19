# Implementation Strategy - Complete Auth System & Beyond

## 📋 Overview

This document outlines the step-by-step strategy for completing the Refyne backend implementation, following professional development practices with clean, production-ready code.

---

## ✅ Phase 0: Critical Fixes (COMPLETED)

1. ✅ Fixed JWT package import
2. ✅ Registered RequestID middleware
3. ✅ Created .env.example file
4. ✅ Completed user_settings migration
5. ✅ Fixed Go version compatibility
6. ✅ Build verification

---

## ✅ Phase 1.1: JWT Authentication & Logout (COMPLETED)

1. ✅ JWT authentication middleware
2. ✅ Token blacklist manager
3. ✅ Logout functionality (single device)
4. ✅ Logout all devices
5. ✅ Protected route examples
6. ✅ Helper functions for user context

---

## 🎯 Phase 1.2: Password Reset Flow (NEXT - 2-3 hours)

### Step 1: Database Migration for Reset Tokens
**Files to create:**
- `internal/database/migrations/sql/000003_create_password_reset_tokens.up.sql`
- `internal/database/migrations/sql/000003_create_password_reset_tokens.down.sql`

**Schema:**
```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Step 2: Password Reset Repository
**File:** `internal/domains/user/core/repository/password_reset.go`

**Methods:**
- `CreateResetToken(userID, token, expiresAt)`
- `GetResetToken(token)` 
- `MarkTokenAsUsed(token)`
- `DeleteExpiredTokens()`

### Step 3: Password Reset Service
**File:** `internal/domains/auth/services/password_reset.go`

**Methods:**
- `RequestPasswordReset(email)` - Generate and store token
- `ValidateResetToken(token)` - Check validity
- `ResetPassword(token, newPassword)` - Update password
- `InvalidateUserResetTokens(userID)` - Security feature

### Step 4: Password Reset Handlers
**File:** `internal/domains/auth/handler/password_reset.go`

**Endpoints:**
- `ForgotPassword(c)` - `POST /api/auth/forgot-password`
- `ResetPassword(c)` - `POST /api/auth/reset-password`
- `ValidateResetToken(c)` - `GET /api/auth/reset-password/validate/:token`

### Step 5: Update Interfaces & Routes
- Update `AuthService` interface
- Update `AuthHandler` interface  
- Register new routes in `auth/routes/auth.go`

**Testing checklist:**
- [ ] Request password reset
- [ ] Validate reset token
- [ ] Reset password successfully
- [ ] Verify old password no longer works
- [ ] Verify token is marked as used
- [ ] Test expired token rejection

---

## 🎯 Phase 1.3: Email Service Integration (3-4 hours)

### Step 1: Email Service Foundation
**File:** `internal/domains/email/service/email.go`

**Interface:**
```go
type EmailService interface {
    SendEmail(to, subject, body string)
    SendTemplatedEmail(to, templateName string, data interface{})
    SendOTP(to, otp string)
    SendPasswordReset(to, token string)
    SendWelcome(to, username string)
}
```

### Step 2: SMTP Configuration
**File:** `internal/domains/email/config/smtp.go`

**Features:**
- SMTP connection pooling
- TLS/SSL configuration
- Retry mechanism
- Error handling

### Step 3: Email Templates
**Files:** `internal/domains/email/templates/`
- `welcome.html` - Welcome email
- `otp.html` - OTP email
- `password_reset.html` - Password reset
- `verification.html` - Account verification

**Template engine:** Go html/template

### Step 4: River Job Integration
**File:** `internal/domains/email/jobs/send_email.go`

**Implementation:**
- Async email sending via River queue
- Retry on failure
- Job tracking
- Error logging

### Step 5: Integration Points
**Update files:**
- `internal/domains/auth/services/auth.go` - Send OTP via email
- `internal/domains/auth/services/password_reset.go` - Send reset email
- Registration handler - Send welcome email

**Testing checklist:**
- [ ] SMTP connection works
- [ ] Email templates render correctly
- [ ] OTP sent via email
- [ ] Password reset email sent
- [ ] Welcome email on registration
- [ ] Async job processing works
- [ ] Email retry on failure

---

## 🎯 Phase 2: User Management (4-5 hours)

### Step 1: User Profile Endpoints
**File:** `internal/domains/user/handler/profile.go`

**Endpoints:**
- `GET /api/user/profile` - Get current user profile
- `PUT /api/user/profile` - Update profile
- `DELETE /api/user/account` - Soft delete account
- `POST /api/user/profile/avatar` - Upload avatar (future)

### Step 2: User Settings Management
**File:** `internal/domains/user/handler/settings.go`

**Endpoints:**
- `GET /api/user/settings` - Get user settings
- `PUT /api/user/settings` - Update settings
- `PATCH /api/user/settings` - Partial update

### Step 3: User Service Layer
**File:** `internal/domains/user/services/user.go`

**Methods:**
- `GetUserProfile(userID)`
- `UpdateUserProfile(userID, updates)`
- `UpdateUserSettings(userID, settings)`
- `DeactivateAccount(userID)`
- `ReactivateAccount(userID)`

### Step 4: Extended Repository Methods
**File:** `internal/domains/user/core/repository/crud.go`

**Add methods:**
- `UpdateUser(userID, updates)`
- `SoftDeleteUser(userID)`
- `GetUserSettings(userID)`
- `UpdateUserSettings(userID, settings)`
- `CreateDefaultSettings(userID)`

### Step 5: User Routes & Registry
**File:** `internal/domains/user/routes/user.go`

**Implementation:**
- Create route group `/api/user`
- Apply auth middleware
- Register all user endpoints

**Testing checklist:**
- [ ] Get user profile
- [ ] Update profile fields
- [ ] Get user settings
- [ ] Update settings
- [ ] Account deactivation
- [ ] Prevent access after deactivation

---

## 🎯 Phase 3: Workspace Management (5-6 hours)

### Step 1: Workspace Database Schema
**Migration:** `000004_create_workspaces.sql`

**Tables:**
- `workspaces` - Workspace details
- `workspace_members` - Member associations
- `workspace_roles` - Role definitions

### Step 2: Workspace Models
**File:** `internal/domains/workspace/models/workspace.go`

**Models:**
- `Workspace` - Main workspace entity
- `WorkspaceMember` - Member association
- `WorkspaceRole` - Role definition

### Step 3: Workspace Repository
**File:** `internal/domains/workspace/repository/workspace.go`

**Methods:**
- CRUD operations for workspaces
- Member management
- Role assignment
- Access control queries

### Step 4: Workspace Service
**File:** `internal/domains/workspace/services/workspace.go`

**Business logic:**
- Workspace creation
- Member invitation
- Permission checking
- Workspace settings

### Step 5: Workspace Handlers & Routes
**Files:**
- `internal/domains/workspace/handler/workspace.go`
- `internal/domains/workspace/routes/workspace.go`

**Endpoints:**
- Workspace CRUD
- Member management
- Settings

---

## 🎯 Phase 4: Additional Features (Priority-based)

### Option A: AI Integration
- Google Gemini API integration
- Context analysis
- Response generation
- Caching layer

### Option B: Social Media Integration
- Instagram API integration
- Comment fetching
- Engagement tracking

### Option C: Analytics & Reporting
- Analytics service
- Data aggregation
- Report generation

---

## 🛠️ Development Best Practices

### Code Quality Standards
1. **Error Handling:**
   - Use custom `AppError` types
   - Proper HTTP status codes
   - Client-safe error messages
   - Internal error logging

2. **Logging:**
   - Request ID in all logs
   - Structured logging with Zap
   - Appropriate log levels
   - Context information

3. **Validation:**
   - Input validation at handler level
   - Business rule validation in service
   - Database constraints
   - Clear validation messages

4. **Testing:**
   - Unit tests for services
   - Integration tests for repositories
   - API tests for handlers
   - Edge case coverage

5. **Documentation:**
   - Godoc comments
   - API documentation
   - Implementation notes
   - TODO comments for future work

### Security Checklist
- ✅ JWT authentication
- ✅ Password hashing
- ✅ Input validation
- ✅ SQL injection prevention
- ⏳ Rate limiting
- ⏳ CORS configuration
- ⏳ Request size limits
- ⏳ XSS prevention
- ⏳ CSRF tokens (for web)

### Performance Considerations
- Database query optimization
- Connection pooling
- Caching strategy
- Async job processing
- Pagination for lists
- Selective field loading

---

## 📊 Progress Tracking

### Completed
- ✅ Phase 0: Critical Fixes
- ✅ Phase 1.1: JWT Auth & Logout

### In Progress
- 🔄 Phase 1.2: Password Reset Flow

### Upcoming
- ⏳ Phase 1.3: Email Service
- ⏳ Phase 2: User Management
- ⏳ Phase 3: Workspace Management
- ⏳ Phase 4: Additional Features

---

## 🚀 Deployment Considerations

### Environment-specific Configs
- Development
- Staging  
- Production

### Database Migrations
- Automated migration on startup (development)
- Manual approval for production
- Rollback procedures

### Monitoring & Alerting
- Health checks
- Error rate monitoring
- Performance metrics
- Token blacklist size

### Scaling Considerations
- Redis for token blacklist
- Load balancing
- Database read replicas
- CDN for static assets

---

## 📝 Next Immediate Action

**Start Phase 1.2: Password Reset Flow**

1. Create password reset tokens migration
2. Implement repository methods
3. Add service logic
4. Create handlers
5. Update interfaces and routes
6. Test end-to-end

**Estimated Time:** 2-3 hours  
**Complexity:** Medium  
**Dependencies:** None (ready to start)

---

**Strategy Last Updated:** November 18, 2025  
**Current Phase:** 1.2 (Password Reset)  
**Overall Progress:** ~25% Complete
