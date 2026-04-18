# 🔴 CRITICAL FINDING: OTP Blocked Due to Email Verification Requirement

**To:** Frontend Team  
**From:** Backend Team  
**Date:** 2026-04-19  
**Priority:** HIGH  
**Status:** ROOT CAUSE IDENTIFIED ✅

---

## TL;DR

**Your OTP request is failing because the test user hasn't verified their email yet.**

New users created via registration are in an **unverified, inactive state** and cannot request OTP until they verify their email first.

---

## The Issue

### User Registration Creates Unverified Account
```
POST /api/auth/register
↓
User Created with:
  - is_verified: false
  - is_active: false
  - status: "inactive"
```

### OTP Request Requires Verified Account
```
POST /api/auth/otp/send
↓
Backend Checks:
  1. Is user verified? ❌ NO → BLOCKED
  2. Is user active? ❌ NO → BLOCKED  
  3. Is status "active"? ❌ NO → BLOCKED
↓
Returns: 400 "User Not Verified" or "User Not Active"
```

---

## Why This Happens

The backend treats OTP as a **2-factor authentication** feature for already-verified users, not as an alternative signup method.

**Required State for OTP Login:**
- ✅ `is_verified = true` (must verify email first)
- ✅ `is_active = true` (must complete verification)
- ✅ `status = "active"` (account fully activated)

---

## How to Test OTP Now

### Step 1: Register User
```bash
POST /api/auth/register
{
  "first_name": "Test",
  "last_name": "User",
  "username": "testuser123",
  "email": "test@example.com",
  "password": "SecurePass123!"
}
```

**Response:** User created (unverified)

### Step 2: Get Verification Token
Either:
- **Option A:** Check email for verification link (copy token from URL)
- **Option B:** Check backend logs/database directly
- **Option C:** Have backend team send you a test token

### Step 3: Verify Email First
```bash
POST /api/auth/verify/email
{
  "token": "[verification-token-from-email]"
}
```

**Response:** Account now verified & active

### Step 4: Now OTP Works
```bash
POST /api/auth/otp/send
{
  "email": "test@example.com",
  "password": "SecurePass123!"
}
```

**Response:** ✅ OTP sent successfully!

---

## The Complete Flow

```
┌─────────────────────────────────────────────┐
│ 1. User Registration                        │
├─────────────────────────────────────────────┤
│ POST /api/auth/register                     │
│ ↓ Creates user (unverified)                 │
│ ↓ Sends verification email                  │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ 2. Email Verification (Required First!)     │
├─────────────────────────────────────────────┤
│ User clicks link in email                   │
│ Frontend: POST /api/auth/verify/email       │
│ ↓ Marks user as verified                    │
│ ↓ Activates account                         │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ 3. OTP Login (Now Works!)                   │
├─────────────────────────────────────────────┤
│ POST /api/auth/otp/send                     │
│ ↓ OTP generated & sent via email            │
│ ↓ User receives OTP code                    │
│ ↓ POST /api/auth/otp/verify with OTP       │
│ ↓ User logged in with tokens                │
└─────────────────────────────────────────────┘
```

---

## Backend Decision Point ⚠️

The current design requires email verification before OTP login.

**Backend Team:** Please confirm:

1. **Is this the intended design?**
   - If YES: Document this requirement clearly
   - If NO: Remove verification check from OTP endpoint

2. **What's the OTP use case?**
   - 2FA for verified users? (current implementation)
   - Alternative auth for unverified users? (needs change)
   - Both? (needs configuration)

---

## For Frontend Implementation

### Update Your Signup/Login Flow

**For New Users:**
```javascript
// After registration
1. Show "Check your email for verification"
2. User clicks link in email → verify email
3. Redirect to login page
4. User can now use password OR OTP login

// NOT:
// Allow OTP immediately after registration ❌
```

**For Existing Verified Users:**
```javascript
// OTP works anytime
1. Enter email & password
2. Click "Use OTP instead"
3. OTP sent via email
4. Enter OTP → logged in ✅
```

---

## Quick Test Script

```bash
#!/bin/bash

# 1. Register
REGISTER=$(curl -X POST https://refyne-backend-production.up.railway.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name":"Test",
    "last_name":"User",
    "username":"testuser'$(date +%s)'",
    "email":"test'$(date +%s)'@example.com",
    "password":"SecurePass123!"
  }')

echo "Registration Response:"
echo $REGISTER | jq .
EMAIL=$(echo $REGISTER | jq -r '.data.email')

# 2. Get token from backend/logs (you'll need this)
# TOKEN="[get from email or logs]"

# 3. Verify email (requires token)
# VERIFY=$(curl -X POST https://refyne-backend-production.up.railway.app/api/auth/verify/email \
#   -H "Content-Type: application/json" \
#   -d '{"token":"'$TOKEN'"}')

# 4. Now try OTP
# OTP=$(curl -X POST https://refyne-backend-production.up.railway.app/api/auth/otp/send \
#   -H "Content-Type: application/json" \
#   -d '{"email":"'$EMAIL'","password":"SecurePass123!"}')
```

---

## What Changed

**Nothing code-related.** The OTP endpoint is working correctly.

**What we discovered:** OTP requires verified users. This is by design or needs backend decision.

---

## Next Actions

- [ ] **Frontend:** Follow Step 1-4 above to test OTP flow
- [ ] **Backend:** Confirm if OTP should allow unverified users
- [ ] **Both:** Update documentation to reflect actual flow
- [ ] **Test:** Once verified, OTP should work perfectly

---

## Backend Team Support

**If you need to enable OTP for unverified users:**
Contact backend team - requires modifying RequestOTP validation logic (2-line change).

**If current design is correct:**
Update frontend documentation to match this requirement.

---

**Status:** ROOT CAUSE FOUND ✅  
**Blocker:** Design Decision on OTP for Unverified Users  
**Frontend Action:** Follow email verification → then test OTP  

**Let's move forward! 🚀**
