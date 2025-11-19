# 🚀 Quick Start - Postman in VS Code

## Step 1: Import Collection & Environment
```
Ctrl+Shift+P → "Postman: Import"
Select: Refyne_API.postman_collection.json
Select: Refyne_Local.postman_environment.json
```

## Step 2: Activate Environment
- Click Postman icon in Activity Bar (left)
- Click "Environments" → "Refyne Local Development"
- Verify ✓ checkmark appears

## Step 3: Start Server
```powershell
cd D:\Refyne\refyne-backend
.\bin\app.exe
```

## Step 4: Test Flow (Click Send on each)
1. ✅ Health Check
2. ✅ Register User
3. 🔧 **Run SQL:** `UPDATE users SET is_verified=true, is_active=true, status='active' WHERE email='test@example.com';`
4. ✅ Request OTP (OTP auto-saves)
5. ✅ Login with OTP (tokens auto-save)
6. ✅ Test Protected Route (uses saved token)
7. ✅ Logout
8. ❌ Test Protected Route (should fail - 401)

---

## 🎯 All Set!

- Collection: **Refyne Backend API**
- Environment: **Refyne Local Development**  
- Base URL: **http://localhost:8080/api**

**Automatic Features:**
- 💾 Tokens saved on login
- 💾 OTP saved on request
- 🔄 Tokens refreshed automatically
- ✅ Tests run automatically

**Keyboard Shortcuts:**
- `Ctrl+Enter` - Send request
- `Ctrl+Shift+P` - Command palette

---

## ⚡ Troubleshooting

**"Connection refused"** → Server not running
**"User not verified"** → Run the SQL UPDATE command
**"Invalid OTP"** → Check the `otp` variable in environment
**"Unauthorized"** → Check `access_token` variable is set

---

See **VSCODE_POSTMAN_GUIDE.md** for detailed instructions!
