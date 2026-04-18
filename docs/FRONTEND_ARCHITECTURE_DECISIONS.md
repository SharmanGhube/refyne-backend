# Frontend Architecture Decisions

This document answers critical architectural questions for frontend development with clear recommendations and reasoning.

**Date:** 2026-04-18  
**Status:** Approved for Implementation

---

## Question 1: Email Verification Flow - Auto-Redirect or Show Verification Page?

### Question
Should OTP verification auto-redirect to dashboard or show verification page?

### Answer: **SHOW VERIFICATION PAGE** (with auto-redirect option)

### Reasoning

**Recommended Flow:**
1. User registers → Email sent
2. User clicks verification link with token
3. Frontend shows **verification confirmation page** (not auto-redirect)
4. Display success message: "Email verified! Redirecting to login in 3 seconds..."
5. After 3-second countdown → Auto-redirect to `/login`

### Implementation Details

**Why NOT direct auto-redirect:**
- ❌ User doesn't see confirmation they completed verification
- ❌ No visual feedback on success
- ❌ Technical issues could go unnoticed
- ❌ Poor UX for slower networks (seems broken)
- ❌ User might try to re-verify, causing errors

**Why this approach is better:**
- ✅ Clear visual confirmation
- ✅ User knows exactly what happened
- ✅ Time for page to load properly
- ✅ Professional, polished feeling
- ✅ Mobile-friendly
- ✅ Accessibility: Screen readers get confirmation message

### Code Example

```typescript
// src/pages/VerifyEmailPage.tsx

export function VerifyEmailPage() {
  const [status, setStatus] = useState<'verifying' | 'success' | 'error'>('verifying');
  const [countdown, setCountdown] = useState(3);
  const navigate = useNavigate();
  const searchParams = new URLSearchParams(window.location.search);
  const token = searchParams.get('token');

  useEffect(() => {
    async function verifyEmail() {
      if (!token) {
        setStatus('error');
        return;
      }

      try {
        // Call backend verification endpoint
        await api.post('/api/auth/verify/email', { token });
        setStatus('success');

        // Start countdown
        const interval = setInterval(() => {
          setCountdown(prev => prev - 1);
        }, 1000);

        // Redirect after 3 seconds
        setTimeout(() => {
          clearInterval(interval);
          navigate('/login');
        }, 3000);
      } catch (error) {
        setStatus('error');
      }
    }

    verifyEmail();
  }, [token, navigate]);

  if (status === 'verifying') {
    return (
      <div className="verification-container">
        <Spinner />
        <p>Verifying your email...</p>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="verification-container success">
        <CheckCircleIcon />
        <h1>Email Verified! 🎉</h1>
        <p>Your email has been verified successfully.</p>
        <p>Redirecting to login in {countdown} seconds...</p>
        <button onClick={() => navigate('/login')}>
          Go to Login Now
        </button>
      </div>
    );
  }

  return (
    <div className="verification-container error">
      <ErrorIcon />
      <h1>Verification Failed</h1>
      <p>The verification link is invalid or has expired.</p>
      <button onClick={() => navigate('/register')}>
        Try Again
      </button>
      <a href="/forgot-password">Request New Link</a>
    </div>
  );
}
```

### Alternative: Skip Verification Page (Not Recommended)

If you absolutely must skip the page:
```typescript
// Only if UX specifically requires it
if (verified) {
  navigate('/login', { state: { verified: true } });
  // Show toast: "Email verified! Please log in"
}
```

### Database/Backend Notes

The backend `/api/auth/verify/email` endpoint:
- Takes token as parameter
- Returns success response
- Updates `email_verified` flag in database
- Does NOT auto-login user (requires explicit login)

---

## Question 2: Instagram OAuth Callback - Frontend or Backend?

### Question
Should `/instagram/callback` exist on frontend to handle OAuth redirect, or is it managed by backend?

### Answer: **FRONTEND MANAGES IT** (with backend handling the OAuth token exchange)

### Recommended Flow

```
1. User clicks "Connect Instagram"
   ↓
2. Frontend: GET /api/instagram/auth/url (from backend)
   ↓
3. Backend returns: { auth_url: "https://instagram.com/oauth/authorize?..." }
   ↓
4. Frontend: window.location.href = auth_url (redirect user to Instagram)
   ↓
5. Instagram: User authenticates, Instagram redirects back to callback URL
   ↓
6. Frontend: /instagram/callback route receives OAuth code
   ↓
7. Frontend: POST /api/instagram/auth/callback with { code }
   ↓
8. Backend: Exchanges code for access token, stores it
   ↓
9. Frontend: Shows success message, redirects to Instagram dashboard
```

### Implementation Details

**Frontend Route - `/instagram/callback`:**

```typescript
// src/pages/InstagramCallbackPage.tsx

export function InstagramCallbackPage() {
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const searchParams = new URLSearchParams(window.location.search);
  const code = searchParams.get('code');
  const state = searchParams.get('state');

  useEffect(() => {
    async function handleCallback() {
      if (!code) {
        setStatus('error');
        setError('No authorization code received');
        return;
      }

      try {
        // Verify state for CSRF protection
        const storedState = sessionStorage.getItem('instagram_oauth_state');
        if (state !== storedState) {
          throw new Error('State mismatch - CSRF protection failed');
        }

        // Exchange code for token
        const response = await api.post('/api/instagram/auth/callback', {
          code,
          state
        });

        setStatus('success');

        // Clear stored state
        sessionStorage.removeItem('instagram_oauth_state');

        // Redirect to Instagram feed after 2 seconds
        setTimeout(() => {
          navigate('/instagram/feed');
        }, 2000);

      } catch (err) {
        setStatus('error');
        setError(err.message || 'Failed to connect Instagram account');
      }
    }

    handleCallback();
  }, [code, state, navigate]);

  if (status === 'loading') {
    return (
      <div className="callback-container">
        <Spinner />
        <p>Connecting your Instagram account...</p>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="callback-container success">
        <CheckCircleIcon />
        <h1>Instagram Connected! 🎉</h1>
        <p>Your Instagram account has been successfully connected.</p>
        <p>Redirecting to your feed...</p>
        <button onClick={() => navigate('/instagram/feed')}>
          Go to Feed Now
        </button>
      </div>
    );
  }

  return (
    <div className="callback-container error">
      <ErrorIcon />
      <h1>Connection Failed</h1>
      <p>{error}</p>
      <button onClick={() => navigate('/instagram')}>
        Try Again
      </button>
    </div>
  );
}
```

**Frontend Connection Initiator:**

```typescript
// src/pages/InstagramPage.tsx or src/components/InstagramConnect.tsx

async function handleConnectInstagram() {
  try {
    // Get OAuth URL from backend
    const response = await api.get('/api/instagram/auth/url');
    const { auth_url, state } = response.data;

    // Store state for CSRF protection
    sessionStorage.setItem('instagram_oauth_state', state);

    // Redirect to Instagram OAuth
    window.location.href = auth_url;

  } catch (error) {
    showError('Failed to initialize Instagram connection');
  }
}
```

**Backend Flow - What Backend Does:**

```go
// Backend GET /api/instagram/auth/url
func GetInstagramAuthURL(c *gin.Context) {
  user := getAuthenticatedUser(c)
  
  // Generate state for CSRF protection
  state := generateRandomState()
  
  // Store state in Redis (expires in 10 minutes)
  redis.Set(fmt.Sprintf("instagram:oauth:state:%s", state), user.ID, 10*time.Minute)
  
  // Generate OAuth URL
  authURL := fmt.Sprintf(
    "https://api.instagram.com/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user_profile,user_media&response_type=code&state=%s",
    os.Getenv("INSTAGRAM_CLIENT_ID"),
    os.Getenv("INSTAGRAM_REDIRECT_URI"),
    state,
  )
  
  c.JSON(200, gin.H{
    "auth_url": authURL,
    "state": state,
  })
}

// Backend POST /api/instagram/auth/callback
func HandleInstagramCallback(c *gin.Context) {
  user := getAuthenticatedUser(c)
  code := c.PostForm("code")
  state := c.PostForm("state")
  
  // Verify state
  storedUserID := redis.Get(fmt.Sprintf("instagram:oauth:state:%s", state))
  if storedUserID != user.ID {
    c.JSON(400, errorResponse("Invalid state"))
    return
  }
  
  // Exchange code for access token
  token := exchangeCodeForToken(code)
  
  // Store token in database
  storeInstagramToken(user.ID, token)
  
  c.JSON(200, gin.H{
    "message": "Instagram account connected",
    "account_id": token.AccountID,
  })
}
```

### Callback URL Configuration

**In Backend `.env`:**
```env
INSTAGRAM_REDIRECT_URI=https://refyne-backend-production.up.railway.app/instagram/callback
# OR for development
INSTAGRAM_REDIRECT_URI=http://localhost:3000/instagram/callback
```

**In Frontend `.env`:**
```env
# Frontend doesn't need this - backend handles redirect
# But store for reference:
VITE_INSTAGRAM_REDIRECT_URI=https://app.refyne.io/instagram/callback
```

**In Instagram App Dashboard:**
```
Valid OAuth Redirect URIs:
- http://localhost:3000/instagram/callback (dev)
- https://app.refyne.io/instagram/callback (production)
```

### Security Considerations

✅ **CSRF Protection:**
- Use state parameter to prevent CSRF attacks
- Verify state matches before processing callback

✅ **Token Storage:**
- Never send access token to frontend (keep on backend only)
- Backend stores in secure database
- Frontend never handles raw token

✅ **Error Handling:**
- Instagram may return error: `?error=access_denied`
- Frontend should check for error parameter:
```typescript
const error = searchParams.get('error');
if (error === 'access_denied') {
  setError('You declined the Instagram connection. Please try again.');
  return;
}
```

---

## Question 3: Paddle Checkout - URL Redirect or Iframe Embed?

### Question
Should frontend open Paddle checkout URL or embed iframe?

### Answer: **USE PADDLE CHECKOUT (URL REDIRECT)** - Not Iframe

### Reasoning

| Aspect | URL Redirect | Iframe Embed |
|--------|--------------|--------------|
| **Security** | ✅ Highest (Paddle-hosted) | ⚠️ Risk of XSS |
| **PCI Compliance** | ✅ Automatic (Paddle handles) | ⚠️ Your responsibility |
| **Mobile UX** | ✅ Native fullscreen | ❌ Constrained in iframe |
| **Payment Methods** | ✅ All methods supported | ⚠️ May have issues |
| **Maintenance** | ✅ Auto-updates by Paddle | ❌ Manual updates |
| **Conversion Rate** | ✅ Higher (better UX) | ❌ Lower |
| **Implementation** | ✅ Simple | ❌ Complex |

### Recommended Implementation

**Option 1: Paddle Checkout (RECOMMENDED)**

```typescript
// src/pages/CheckoutPage.tsx

export function CheckoutPage() {
  const [loading, setLoading] = useState(false);

  async function handleSubscribe() {
    try {
      setLoading(true);

      // Create checkout session on backend
      const response = await api.post('/api/subscription/checkout', {
        plan_id: 'pro',
        billing_cycle: 'monthly' // or 'yearly'
      });

      // Redirect to Paddle checkout
      window.location.href = response.data.checkout_url;

    } catch (error) {
      showError('Failed to start checkout');
      setLoading(false);
    }
  }

  return (
    <div className="checkout-container">
      <div className="plan-card">
        <h2>Pro Plan</h2>
        <p className="price">$29<span>/month</span></p>
        
        <ul className="features">
          <li>✓ Connected Instagram accounts</li>
          <li>✓ AI Assistant (Otto)</li>
          <li>✓ Team members</li>
          <li>✓ Advanced analytics</li>
          <li>✓ Priority support</li>
        </ul>

        <button 
          onClick={handleSubscribe}
          disabled={loading}
        >
          {loading ? 'Starting Checkout...' : 'Subscribe Now'}
        </button>
      </div>

      <div className="info">
        <p>🔒 Secure payment powered by Paddle</p>
        <p>Cancel anytime - no lock-in</p>
      </div>
    </div>
  );
}
```

**Backend Implementation:**

```go
// Backend POST /api/subscription/checkout
func CreateCheckout(c *gin.Context) {
  user := getAuthenticatedUser(c)
  
  var req struct {
    PlanID       string `json:"plan_id"`
    BillingCycle string `json:"billing_cycle"`
  }
  
  if err := c.BindJSON(&req); err != nil {
    c.JSON(400, errorResponse("Invalid request"))
    return
  }
  
  // Create checkout via Paddle API
  checkoutURL := paddle.CreateCheckout(paddle.CheckoutRequest{
    ProductID:   getPaddleProductID(req.PlanID),
    CustomerID:  user.ID,
    Email:       user.Email,
    PriceID:     getBillingCyclePriceID(req.BillingCycle),
    ReturnURL:   "https://app.refyne.io/subscription/success",
    CancelURL:   "https://app.refyne.io/subscription/canceled",
  })
  
  c.JSON(200, gin.H{
    "checkout_url": checkoutURL,
  })
}
```

**After Checkout Success:**

```typescript
// src/pages/CheckoutSuccessPage.tsx

export function CheckoutSuccessPage() {
  const [status, setStatus] = useState<'checking' | 'success' | 'error'>('checking');

  useEffect(() => {
    async function verifySubscription() {
      try {
        // Check if subscription was created
        const response = await api.get('/api/subscription/status');
        
        if (response.data.status === 'active') {
          setStatus('success');
          
          // Redirect to dashboard after 3 seconds
          setTimeout(() => {
            navigate('/dashboard');
          }, 3000);
        } else {
          setStatus('error');
        }
      } catch (error) {
        setStatus('error');
      }
    }

    verifySubscription();
  }, []);

  if (status === 'checking') {
    return (
      <div className="success-container">
        <Spinner />
        <p>Verifying your subscription...</p>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="success-container">
        <CheckCircleIcon />
        <h1>Welcome to Pro! 🎉</h1>
        <p>Your subscription is active and ready to use.</p>
        <p>Redirecting to dashboard...</p>
        <button onClick={() => navigate('/dashboard')}>
          Go to Dashboard Now
        </button>
      </div>
    );
  }

  return (
    <div className="success-container error">
      <ErrorIcon />
      <h1>Subscription Verification Failed</h1>
      <p>We couldn't verify your subscription. Please contact support.</p>
      <button onClick={() => navigate('/subscription')}>
        Try Again
      </button>
    </div>
  );
}
```

### Why NOT Paddle Iframe

❌ **Don't use `<PaddleCheckout>` React component:**
- Adds unnecessary complexity
- Worse mobile experience
- Not optimized for conversions
- Requires iframe embedding
- Potential security issues

### Environment Configuration

```env
# Frontend .env
VITE_PADDLE_CLIENT_TOKEN=your_paddle_token  # Used for other Paddle features, not checkout

# Backend .env
PADDLE_LIVE_API_KEY=your_paddle_api_key
PADDLE_SANDBOX_API_KEY=your_sandbox_api_key
PADDLE_PRODUCT_ID_PRO=pro_123456
PADDLE_WEBHOOK_SECRET=your_webhook_secret
```

### Testing Checkout Locally

```bash
# Use Paddle Sandbox
PAYMENT_MODE=sandbox make run

# Test card numbers:
# Success: 4242 4242 4242 4242
# Decline: 4111 1111 1111 1111
```

---

## Question 4: Notification System - Toast vs Alert?

### Question
Should I create a toast notification component alongside the Alert system?

### Answer: **YES - Create Both** (Different purposes)

### When to Use What

| Situation | Toast | Alert |
|-----------|-------|-------|
| Success message (temporary) | ✅ Yes | ❌ No |
| Error message (requires action) | ❌ No | ✅ Yes |
| Confirmation needed | ❌ No | ✅ Yes |
| Non-blocking info | ✅ Yes | ❌ No |
| Blocking warning | ❌ No | ✅ Yes |
| Auto-dismiss | ✅ Yes | ❌ No |
| User needs to decide | ❌ No | ✅ Yes |
| "Saved successfully" | ✅ Yes | ❌ No |
| "Delete account?" | ❌ No | ✅ Yes |

### Architecture

**Toast Notification System:**

```typescript
// src/stores/notificationStore.ts
import { create } from 'zustand';

export interface Toast {
  id: string;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
  duration?: number; // ms, 0 = persistent
  action?: {
    label: string;
    onClick: () => void;
  };
}

interface NotificationStore {
  toasts: Toast[];
  addToast: (toast: Omit<Toast, 'id'>) => string;
  removeToast: (id: string) => void;
  clearAll: () => void;
}

export const useNotificationStore = create<NotificationStore>((set, get) => ({
  toasts: [],

  addToast: (toast) => {
    const id = Math.random().toString(36).substr(2, 9);
    const duration = toast.duration ?? 4000; // Default 4 seconds

    set(state => ({
      toasts: [...state.toasts, { ...toast, id, duration }]
    }));

    // Auto-remove after duration
    if (duration > 0) {
      setTimeout(() => {
        get().removeToast(id);
      }, duration);
    }

    return id;
  },

  removeToast: (id) => {
    set(state => ({
      toasts: state.toasts.filter(t => t.id !== id)
    }));
  },

  clearAll: () => {
    set({ toasts: [] });
  }
}));

// Helper functions
export function showSuccess(message: string, duration = 4000) {
  useNotificationStore.getState().addToast({
    message,
    type: 'success',
    duration
  });
}

export function showError(message: string, duration = 6000) {
  useNotificationStore.getState().addToast({
    message,
    type: 'error',
    duration
  });
}

export function showWarning(message: string, duration = 5000) {
  useNotificationStore.getState().addToast({
    message,
    type: 'warning',
    duration
  });
}

export function showInfo(message: string, duration = 4000) {
  useNotificationStore.getState().addToast({
    message,
    type: 'info',
    duration
  });
}
```

**Toast Component:**

```typescript
// src/components/Toast.tsx
import { useNotificationStore } from '@/stores/notificationStore';
import { useEffect } from 'react';

export function ToastContainer() {
  const { toasts, removeToast } = useNotificationStore();

  return (
    <div className="toast-container">
      {toasts.map(toast => (
        <ToastItem
          key={toast.id}
          toast={toast}
          onClose={() => removeToast(toast.id)}
        />
      ))}
    </div>
  );
}

interface ToastItemProps {
  toast: Toast;
  onClose: () => void;
}

function ToastItem({ toast, onClose }: ToastItemProps) {
  useEffect(() => {
    if (toast.duration > 0) {
      const timer = setTimeout(onClose, toast.duration);
      return () => clearTimeout(timer);
    }
  }, [toast.duration, onClose]);

  const iconMap = {
    success: '✓',
    error: '✕',
    warning: '⚠',
    info: 'ⓘ'
  };

  return (
    <div className={`toast toast-${toast.type}`}>
      <span className="toast-icon">{iconMap[toast.type]}</span>
      <span className="toast-message">{toast.message}</span>
      
      {toast.action && (
        <button 
          className="toast-action"
          onClick={() => {
            toast.action?.onClick();
            onClose();
          }}
        >
          {toast.action.label}
        </button>
      )}
      
      <button 
        className="toast-close"
        onClick={onClose}
        aria-label="Close"
      >
        ✕
      </button>
    </div>
  );
}
```

**CSS for Toast:**

```css
.toast-container {
  position: fixed;
  bottom: 20px;
  right: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 9999;
  pointer-events: none;
  max-width: 400px;
}

.toast {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  pointer-events: auto;
  animation: slideIn 0.3s ease-out;
}

.toast-success {
  border-left: 4px solid #10b981;
}

.toast-error {
  border-left: 4px solid #ef4444;
}

.toast-warning {
  border-left: 4px solid #f59e0b;
}

.toast-info {
  border-left: 4px solid #3b82f6;
}

.toast-icon {
  font-weight: bold;
  font-size: 18px;
}

.toast-message {
  flex: 1;
  font-size: 14px;
}

.toast-action {
  background: none;
  border: none;
  color: #3b82f6;
  cursor: pointer;
  font-weight: 500;
  padding: 0;
}

.toast-close {
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  font-size: 16px;
  padding: 0;
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@media (max-width: 640px) {
  .toast-container {
    left: 20px;
    right: 20px;
    bottom: 20px;
    max-width: none;
  }
}
```

**Alert/Modal System:**

```typescript
// src/stores/alertStore.ts
import { create } from 'zustand';

export interface AlertDialog {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
  buttons: AlertButton[];
}

export interface AlertButton {
  label: string;
  onClick: () => void | Promise<void>;
  variant: 'primary' | 'secondary' | 'danger';
}

interface AlertStore {
  alerts: AlertDialog[];
  showAlert: (alert: Omit<AlertDialog, 'id'>) => Promise<void>;
  removeAlert: (id: string) => void;
}

export const useAlertStore = create<AlertStore>((set, get) => ({
  alerts: [],

  showAlert: async (alert) => {
    const id = Math.random().toString(36).substr(2, 9);
    
    return new Promise((resolve) => {
      set(state => ({
        alerts: [...state.alerts, { ...alert, id }]
      }));

      // Store resolve in window for button click handlers
      (window as any)[`alert_${id}_resolve`] = resolve;
    });
  },

  removeAlert: (id) => {
    set(state => ({
      alerts: state.alerts.filter(a => a.id !== id)
    }));
    
    // Cleanup
    delete (window as any)[`alert_${id}_resolve`];
  }
}));

// Helper function for confirmation dialogs
export async function confirmDelete(itemName: string): Promise<boolean> {
  return new Promise((resolve) => {
    useAlertStore.getState().showAlert({
      title: 'Delete ' + itemName,
      message: `Are you sure you want to delete this ${itemName}? This action cannot be undone.`,
      type: 'warning',
      buttons: [
        {
          label: 'Cancel',
          onClick: () => resolve(false),
          variant: 'secondary'
        },
        {
          label: 'Delete',
          onClick: () => resolve(true),
          variant: 'danger'
        }
      ]
    });
  });
}
```

**Alert Component:**

```typescript
// src/components/AlertDialog.tsx
import { useAlertStore } from '@/stores/alertStore';

export function AlertContainer() {
  const { alerts, removeAlert } = useAlertStore();

  return (
    <>
      {alerts.map(alert => (
        <AlertDialog
          key={alert.id}
          alert={alert}
          onClose={() => removeAlert(alert.id)}
        />
      ))}
    </>
  );
}

interface AlertDialogProps {
  alert: AlertDialog;
  onClose: () => void;
}

function AlertDialog({ alert, onClose }: AlertDialogProps) {
  const handleButtonClick = async (button: AlertButton) => {
    await button.onClick();
    onClose();
  };

  const iconMap = {
    info: 'ⓘ',
    warning: '⚠',
    error: '✕',
    success: '✓'
  };

  return (
    <div className="alert-overlay">
      <div className={`alert-dialog alert-${alert.type}`}>
        <div className="alert-header">
          <span className="alert-icon">{iconMap[alert.type]}</span>
          <h2 className="alert-title">{alert.title}</h2>
        </div>

        <div className="alert-content">
          <p>{alert.message}</p>
        </div>

        <div className="alert-footer">
          {alert.buttons.map((button, idx) => (
            <button
              key={idx}
              className={`btn btn-${button.variant}`}
              onClick={() => handleButtonClick(button)}
            >
              {button.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
```

### Usage Examples

**Toast (temporary, non-blocking):**

```typescript
// In a form submission
async function handleSaveProfile(data) {
  try {
    await api.put('/api/user/profile', data);
    showSuccess('Profile updated successfully');
    // No modal, user continues working
  } catch (error) {
    showError('Failed to update profile');
  }
}
```

**Alert (blocking, requires action):**

```typescript
// For destructive actions
async function handleDeleteAccount() {
  const confirmed = await confirmDelete('account');
  
  if (confirmed) {
    try {
      await api.delete('/api/user/account');
      showSuccess('Account deleted');
      navigate('/login');
    } catch (error) {
      showError('Failed to delete account');
    }
  }
}
```

**Toast with Action:**

```typescript
function handleUndoableAction() {
  showSuccess('Item archived', {
    label: 'Undo',
    onClick: () => unarchiveItem()
  });
}
```

### Setup in App

```typescript
// src/App.tsx
import { ToastContainer } from '@/components/Toast';
import { AlertContainer } from '@/components/AlertDialog';

export function App() {
  return (
    <>
      {/* Main app content */}
      <Router>
        {/* ... */}
      </Router>

      {/* Global notification systems */}
      <ToastContainer />
      <AlertContainer />
    </>
  );
}
```

### Summary

✅ **Use Toasts for:**
- Success confirmations
- Informational messages
- Non-blocking feedback
- Auto-dismiss scenarios

✅ **Use Alerts for:**
- Confirmation dialogs
- Destructive action warnings
- User decisions
- Required acknowledgments

Both systems coexist and serve different UX purposes!

---

## Summary of Decisions

| Decision | Recommendation | Key Point |
|----------|---|---|
| **Email Verification** | Show verification page with 3s redirect | Better UX + visual confirmation |
| **Instagram OAuth** | Frontend `/instagram/callback` route | Frontend handles callback, backend exchanges token |
| **Paddle Checkout** | URL redirect (not iframe) | Better security, UX, and compliance |
| **Notifications** | Both toasts + alerts | Different purposes: toasts are temporary, alerts are blocking |

---

**Document Version:** 1.0  
**Status:** Approved for Implementation  
**Frontend Development can proceed with these decisions**
