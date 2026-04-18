# ⚙️ SMTP Configuration Fix - Step by Step

**Status:** SMTP timeout on Railway  
**Root Cause:** Missing SMTP_USE_TLS configuration + possible network issue  
**Fix Time:** 5-10 minutes

---

## 🔴 Problem

OTP requests timeout after 2+ minutes:
```
Error: dial tcp 142.250.101.108:587: connect: connection timed out
```

**Root Causes:**
1. ❌ `SMTP_USE_TLS` not set to `true` (defaults to false)
2. ❌ Railway container may not have outbound SMTP access
3. ❌ Gmail may be blocking the connection

---

## ✅ Solution

### Step 1: Update Railway Environment Variables

Go to **Railway Dashboard → Your Service → Variables** and add/update:

```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-gmail@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_USE_TLS=true
SMTP_USE_SSL=false
```

**Critical:** Make sure `SMTP_USE_TLS=true` is set!

### Step 2: Test Gmail App Password

1. Ensure Gmail account has 2FA enabled
2. Go to: https://myaccount.google.com/apppasswords
3. Generate an **App Password** for "Mail" on "Windows PC" (or your device)
4. Use that password in `SMTP_PASSWORD` (NOT your regular Gmail password)

### Step 3: Redeploy Backend

After updating variables, redeploy:

```bash
git push origin main
# Wait for Railway auto-deploy
# Or manually trigger deploy from Railway dashboard
```

### Step 4: Test SMTP Connectivity

Run from Railway container (via Railway shell or logs):

```bash
# Test DNS resolution
nslookup smtp.gmail.com

# Test TCP connection to port 587
timeout 5 bash -c 'cat < /dev/null > /dev/tcp/smtp.gmail.com/587' && echo "✅ Connected" || echo "❌ Failed"
```

---

## 🆘 If Still Not Working

### Option A: Check Network Access

Railway might be blocking outbound SMTP. Test with:

```bash
curl -v smtp.gmail.com:587
telnet smtp.gmail.com 587
```

If blocked, contact Railway support or switch to **Option B**.

### Option B: Switch to SendGrid (Recommended)

Gmail SMTP is unreliable in cloud environments. Use SendGrid instead:

**1. Create SendGrid Account**
- Sign up: https://sendgrid.com
- Verify sender domain
- Get API key from Settings

**2. Update Code** (modify `internal/domains/email/service/email.go`):

```go
// Replace SMTP with SendGrid
import "github.com/sendgrid/sendgrid-go"

// Use SendGrid SDK instead of SMTP
```

**3. Update Environment:**
```
SENDGRID_API_KEY=SG.xxxxx
SMTP_HOST=  # Leave empty if using SendGrid
```

---

## 📋 Configuration Checklist

- [ ] `SMTP_HOST=smtp.gmail.com`
- [ ] `SMTP_PORT=587`
- [ ] `SMTP_USE_TLS=true`
- [ ] `SMTP_USE_SSL=false`
- [ ] `SMTP_USERNAME=your-email@gmail.com`
- [ ] `SMTP_PASSWORD=app-password (not regular password!)`
- [ ] Gmail 2FA enabled
- [ ] Gmail App Password generated
- [ ] Backend redeployed after variable changes

---

## 🧪 Testing

After fix deployed, test OTP:

```bash
curl -X POST https://refyne-backend-production.up.railway.app/api/auth/otp/send \
  -H "Content-Type: application/json" \
  -d '{
    "email": "sharmanghube@gmail.com",
    "password": "your-password"
  }'
```

**Expected Response (within 1 second):**
```json
{
  "success": true,
  "code": 200,
  "message": "OTP sent successfully",
  "data": {
    "expires_in": 300
  }
}
```

---

## 📊 Before & After

### Before (Broken)
```
Request → Timeout (2+ minutes) → Error
```

### After (Fixed)
```
Request → STARTTLS connection → Auth → Email sent → Response (< 1s)
```

---

## 🎯 Next Steps

1. Update Railway environment variables with SMTP_USE_TLS=true
2. Redeploy backend
3. Test OTP endpoint
4. If still fails after 5 min, switch to SendGrid

---

**Fix Status:** ⏳ Awaiting Backend Implementation  
**Estimated Time:** 5-10 minutes  
**Risk Level:** Low (no code changes needed)
