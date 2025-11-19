# Testing with Postman VS Code Extension

## 🚀 Quick Setup

### 1. Open Postman in VS Code
- Press `Ctrl+Shift+P`
- Type "Postman: Import"
- Select `Refyne_API.postman_collection.json`

### 2. Import Environment
- In Postman sidebar, click "Environments"
- Click "Import Environment"
- Select `Refyne_Local.postman_environment.json`
- Select "Refyne Local Development" as active environment

### 3. Start the Server
```powershell
cd D:\Refyne\refyne-backend
.\bin\app.exe
```

---

## 📋 Testing Flow (Use VS Code Postman Extension)

### **Method 1: Use Postman Sidebar**

1. **Open Postman View:**
   - Click Postman icon in Activity Bar (left side)
   - Or press `Ctrl+Shift+P` → "Postman: Focus on Postman View"

2. **Expand Collections:**
   - Find "Refyne Backend API"
   - You'll see folders:
     - Health Check
     - Authentication
     - Protected - Auth Required
     - Error Cases

3. **Run Requests in Order:**
   - Click "Health Check" → Click "Send"
   - Click "Register User" → Click "Send"
   
   **⚠️ ACTIVATE USER (Run this SQL):**
   ```sql
   UPDATE users 
   SET is_verified = true, is_active = true, status = 'active' 
   WHERE email = 'test@example.com';
   ```
   
   - Click "Request OTP" → Click "Send" (OTP auto-saves)
   - Click "Login with OTP" → Click "Send" (tokens auto-save)
   - Click "Test Protected Route" → Click "Send"
   - Click "Logout" → Click "Send"

---

## 🎯 **VS Code Specific Features**

### **View Request History**
- Postman sidebar → "History" tab
- See all your previous requests

### **Environment Variables**
- Postman sidebar → "Environments" → "Refyne Local Development"
- Click to see current variables:
  - `access_token` (auto-updated on login)
  - `refresh_token` (auto-updated on login)
  - `otp` (auto-updated on request-otp)
  - `user_email` (manually set to test@example.com)

### **View Response**
After sending a request, you'll see tabs:
- **Body:** Response JSON
- **Headers:** Response headers (check `X-Request-ID`)
- **Test Results:** Auto-test results (✅/❌)
- **Cookies:** Any cookies set

### **Keyboard Shortcuts**
- `Ctrl+Enter` - Send request
- `Ctrl+N` - New request
- `Ctrl+S` - Save request

---

## 🔥 **Quick Test Commands**

### Test Full Flow (Run in Order):
```
1. Health Check          → Should return 200 OK
2. Register User         → Should return 201 Created
3. [MANUAL] Activate user in database
4. Request OTP           → Should return 200 with OTP
5. Login with OTP        → Should return 200 with tokens
6. Test Protected Route  → Should return 200 with user info
7. Logout               → Should return 200
8. Test Protected Route  → Should return 401 (token revoked)
```

---

## 📱 **Collection Structure in VS Code**

```
Refyne Backend API
├── 📁 Health Check
│   └── GET Health Check
├── 📁 Authentication (Public)
│   ├── POST 1. Register User
│   ├── POST 2. Request OTP
│   ├── POST 3. Login with OTP
│   ├── POST 4. Refresh Token
│   └── POST 5. Verify Account
├── 📁 Protected - Auth Required
│   ├── GET Test Protected Route
│   ├── POST Logout (Current Device)
│   ├── POST Logout All Devices
│   └── GET Test After Logout (Should Fail)
└── 📁 Error Cases
    ├── POST Register - Invalid Email
    ├── POST Register - Weak Password
    ├── POST Login - Invalid OTP
    ├── GET Protected Route - No Token
    └── GET Protected Route - Invalid Token
```

---

## ✅ **Automatic Tests**

Each request has automatic tests that show in the "Test Results" tab:

### Registration Tests:
- ✅ Status code is 201
- ✅ Response has success message

### Login Tests:
- ✅ Status code is 200
- ✅ Response contains tokens
- ✅ Response contains user data
- 💾 Tokens saved to environment (automatic)

### Protected Route Tests:
- ✅ Status code is 200
- ✅ Response contains user info

### Logout Tests:
- ✅ Status code is 200
- ✅ Success message present

### Error Tests:
- ✅ Correct error status codes
- ✅ Error messages present

---

## 🐛 **Troubleshooting in VS Code**

### "Cannot import collection"
**Solution:** 
- Make sure files are in the workspace root
- Try: `Ctrl+Shift+P` → "Postman: Import" → Select file again

### "Request failed" or "ECONNREFUSED"
**Solution:**
```powershell
# Check if server is running
cd D:\Refyne\refyne-backend
.\bin\app.exe

# Should see: Server starting on :8080
```

### "No environment selected"
**Solution:**
- Click environment dropdown in Postman panel (top)
- Select "Refyne Local Development"

### "Tests not showing results"
**Solution:**
- Click "Test Results" tab after sending request
- If blank, check Console (View → Output → Postman)

### "Token not being used"
**Solution:**
- Check Environment variables
- Make sure "Refyne Local Development" is active (checkmark)
- Variables should auto-populate after login

---

## 💡 **Pro Tips for VS Code Postman**

### 1. **Split Editor View**
```
Ctrl+\  → Split editor
```
- Keep API collection on left
- View responses on right

### 2. **Quick Switch Between Requests**
```
Ctrl+P → Type request name
```

### 3. **View Server Logs Side-by-Side**
- Open Terminal: `Ctrl+` `
- Run server in terminal
- See logs as you test

### 4. **Format JSON Responses**
- Response appears in editor
- Right-click → "Format Document"
- Or use `Shift+Alt+F`

### 5. **Copy Request as cURL**
- Right-click request → "Copy as cURL"
- Useful for sharing/debugging

### 6. **Use Pre-request Scripts**
Requests already have automatic token management!
- OTP is saved automatically from request-otp
- Tokens are saved automatically on login
- No manual copying needed!

---

## 🔍 **Verify Everything Works**

### 1. Check Collection Imported
- Postman sidebar should show "Refyne Backend API"
- Expand to see all folders and requests

### 2. Check Environment Active
- Look for "Refyne Local Development" at top of Postman panel
- Should have a ✓ checkmark

### 3. Test Health Check
```
GET {{base_url}}/health
```
- Should return: {"status": "ok"}
- Response time should be < 10ms

### 4. View Environment Variables
- Click "Refyne Local Development"
- Should see:
  - base_url: http://localhost:8080/api
  - user_email: test@example.com
  - access_token: (empty until login)
  - refresh_token: (empty until login)
  - otp: (empty until requested)

---

## 🎨 **Visual Guide**

### When Request Succeeds:
```
✅ Status: 200 OK
✅ Time: 15ms
✅ Size: 256 bytes

Body Tab:
{
  "message": "Success",
  "data": {...}
}

Test Results Tab:
✅ Status code is 200
✅ Response has expected fields
```

### When Authentication Works:
```
Protected Route Request:
Authorization: Bearer {{access_token}}  ← Auto-populated!

Response:
{
  "user_id": "abc123...",
  "email": "test@example.com",
  "username": "johndoe"
}
```

### When Token is Blacklisted:
```
After Logout:
❌ Status: 401 Unauthorized

{
  "error": "Unauthorized",
  "message": "Token has been revoked"
}
```

---

## 📊 **Expected Results Summary**

| Request | Status | Auto-Saves | Notes |
|---------|--------|------------|-------|
| Health Check | 200 | - | Server alive |
| Register | 201 | - | Creates user |
| Request OTP | 200 | ✅ otp | Need active user |
| Login | 200 | ✅ tokens, user_id | Returns JWT |
| Protected Route | 200 | - | Uses token |
| Refresh Token | 200 | ✅ new tokens | Updates tokens |
| Logout | 200 | - | Blacklists token |
| After Logout | 401 | - | Token revoked |

---

## 🚀 **Ready to Test!**

1. ✅ Collection imported in VS Code Postman
2. ✅ Environment imported and active
3. ✅ Server running on localhost:8080
4. ✅ PostgreSQL database running

**Start testing by clicking "Health Check" in the Postman sidebar!**

---

## 📞 **Need Help?**

- **Server logs:** Check the terminal running `app.exe`
- **Request details:** Look at Headers, Body tabs
- **Test failures:** Check Test Results tab for specific assertion failures
- **Variables:** Check environment to see what's saved

**All requests include automatic test scripts and variable management - just click Send! 🎉**
