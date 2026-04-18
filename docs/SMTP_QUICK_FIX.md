# 🚀 QUICK FIX - 3 Steps Only

## The Issue
OTP endpoint times out because `SMTP_USE_TLS` is not enabled.

## The Fix (Right Now)

### 1. Go to Railway Dashboard
**Service → Variables → Add/Update:**

```
SMTP_USE_TLS=true
```

### 2. Verify These Variables Exist
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_USE_SSL=false
```

⚠️ **Important:** If `SMTP_PASSWORD` is your regular Gmail password, change it to an **App Password** from: https://myaccount.google.com/apppasswords

### 3. Redeploy
```bash
git push origin main
# Wait for Railway auto-deploy (2-3 min)
```

---

## ✅ Test It Works

```bash
curl -X POST https://refyne-backend-production.up.railway.app/api/auth/otp/send \
  -H "Content-Type: application/json" \
  -d '{
    "email": "sharmanghube@gmail.com",
    "password": "TestPassword123!"
  }'
```

Should return **200 OK** within 1 second.

---

## If Still Broken

Check Railway logs for:
```
Failed to send email via SMTP: dial tcp ...
```

If still timeout = use **SendGrid instead** (more reliable)

---

**Time Needed:** 5 minutes  
**Difficulty:** Easy  
**Risk:** None
