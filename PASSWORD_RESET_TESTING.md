# Password Reset Testing Guide

## Overview
Phase 1.2 - Password Reset Flow is now fully implemented and ready to test!

## New API Endpoints

### 1. Forgot Password
**POST** `/api/auth/forgot-password`
```json
{
  "email": "test@example.com"
}
```
- Requests a password reset for the given email
- Always returns success (prevents email enumeration)
- Generates a 64-character hex token (valid for 1 hour)
- In production, token would be sent via email
- For development, retrieve token from database

### 2. Validate Reset Token
**POST** `/api/auth/validate-reset-token`
```json
{
  "token": "your-reset-token-here"
}
```
- Validates the reset token
- Checks: validity, expiration, usage status
- Returns user_id if valid

### 3. Reset Password
**POST** `/api/auth/reset-password`
```json
{
  "token": "your-reset-token-here",
  "new_password": "NewSecurePass123!"
}
```
- Resets the password using the token
- Invalidates all reset tokens for the user
- Marks the token as used

---

## Testing Steps with Postman

### Step 1: Reload Postman Collection
The collection has been updated with 3 new requests. If you already have it open:
1. In VS Code Postman extension, right-click on **"Refyne Backend API"** collection
2. Select **"Reload"** or re-import the collection

You should now see in the **Authentication** folder:
- 6. Forgot Password
- 7. Validate Reset Token
- 8. Reset Password

### Step 2: Request Password Reset
1. Open **"6. Forgot Password"** request
2. Body uses `{{user_email}}` (test@example.com)
3. Click **Send**
4. ✅ Response: `"If the email exists, a password reset link has been sent"`

### Step 3: Get Reset Token from Database
Since email service isn't implemented yet, retrieve the token from database:

```powershell
docker exec -it refyne-postgres psql -U refyne_user -d refyne_db -c "SELECT token, expires_at, is_valid FROM password_reset_tokens WHERE user_id = (SELECT id FROM users WHERE email = 'test@example.com') ORDER BY created_at DESC LIMIT 1;"
```

**Copy the token value** (64-character hex string)

### Step 4: Set Reset Token in Postman
1. In Postman, click **"Variables"** tab (at the top)
2. Find the `reset_token` variable
3. Paste the token in the **"Current value"** column
4. Press Enter to save

### Step 5: Validate Reset Token (Optional)
1. Open **"7. Validate Reset Token"** request
2. Body uses `{{reset_token}}`
3. Click **Send**
4. ✅ Response: Token is valid + user_id

### Step 6: Reset Password
1. Open **"8. Reset Password"** request
2. Body uses `{{reset_token}}` and sets new password: `NewSecurePass123!`
3. Click **Send**
4. ✅ Response: `"Password has been reset successfully"`

### Step 7: Verify Password Change
1. Go back to **"2. Request OTP"** request
2. Update the body to use the NEW password:
   ```json
   {
     "email": "test@example.com",
     "password": "NewSecurePass123!"
   }
   ```
3. Click **Send**
4. ✅ Should receive OTP successfully

---

## Quick PowerShell Commands

### Get Latest Reset Token
```powershell
docker exec -it refyne-postgres psql -U refyne_user -d refyne_db -c "SELECT token FROM password_reset_tokens WHERE user_id = (SELECT id FROM users WHERE email = 'test@example.com') ORDER BY created_at DESC LIMIT 1;" -t
```

### Check All Reset Tokens for User
```powershell
docker exec -it refyne-postgres psql -U refyne_user -d refyne_db -c "SELECT token, created_at, expires_at, used_at, is_valid FROM password_reset_tokens WHERE user_id = (SELECT id FROM users WHERE email = 'test@example.com') ORDER BY created_at DESC;"
```

### Manually Expire a Token (for testing)
```powershell
docker exec -it refyne-postgres psql -U refyne_user -d refyne_db -c "UPDATE password_reset_tokens SET expires_at = NOW() - INTERVAL '1 hour' WHERE token = 'your-token-here';"
```

### Clean Up Old Tokens
```powershell
docker exec -it refyne-postgres psql -U refyne_user -d refyne_db -c "DELETE FROM password_reset_tokens WHERE expires_at < NOW();"
```

---

## Testing Scenarios

### ✅ Happy Path
1. Request password reset → Success
2. Get token from database
3. Validate token → Valid
4. Reset password → Success
5. Login with new password → Success

### ❌ Error Cases to Test

**Invalid Token:**
- Use a fake token in "8. Reset Password"
- Expected: `401 Unauthorized` - "Password reset token not found"

**Expired Token:**
- Request reset, wait 1 hour, then try to use it
- Or manually expire it with the SQL command above
- Expected: `400 Bad Request` - "Reset token has expired"

**Used Token (Double Use):**
- Reset password successfully
- Try to use the same token again
- Expected: `400 Bad Request` - "Reset token has already been used"

**Non-existent Email:**
- Request reset for fake email
- Expected: `200 OK` (security - doesn't reveal if user exists)

**Inactive User:**
- Request reset for inactive user
- Expected: `400 Bad Request` - "Account is not active"

---

## Security Features Implemented

✅ **Email Enumeration Protection**: Always returns success, regardless of email existence
✅ **Token Expiration**: Tokens expire after 1 hour
✅ **One-Time Use**: Tokens are marked as used after password reset
✅ **Token Invalidation**: All user tokens invalidated after successful reset
✅ **Secure Token Generation**: Cryptographically secure 64-char random hex tokens
✅ **Password Hashing**: bcrypt with default cost (12 rounds)

---

## Next Steps

After testing Phase 1.2, the next phase would be:

**Phase 1.3 - Email Service Integration**
- Integrate SMTP email service
- Send password reset links via email instead of returning tokens
- Send OTP via email instead of in API response
- Welcome email on registration
- Password change confirmation email

Would you like to proceed with Phase 1.3 or test the current implementation first?
