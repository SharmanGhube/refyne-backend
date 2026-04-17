# Frontend Implementation Guide - Refyne
## Next.js Frontend for Production-Grade SaaS

**Target:** Vercel deployment with Railway backend  
**Framework:** Next.js 14+ (App Router)  
**Status:** Ready for development  
**Last Updated:** 2026-04-18

---

## 📋 Table of Contents

1. [Production Architecture Decisions](#production-architecture-decisions)
2. [Authentication & Token Management](#authentication--token-management)
3. [API Integration Setup](#api-integration-setup)
4. [User Onboarding Flow](#user-onboarding-flow)
5. [Freemium Model & Feature Gating](#freemium-model--feature-gating)
6. [Complete Endpoint Reference](#complete-endpoint-reference)
7. [Error Handling & Validation](#error-handling--validation)
8. [State Management](#state-management)
9. [Rate Limiting Strategies](#rate-limiting-strategies)
10. [Subscription & Payment Flow](#subscription--payment-flow)
11. [Environment Configuration](#environment-configuration)

---

## Production Architecture Decisions

### 1. Authentication & Token Storage Strategy

**Decision: Hybrid Secure Approach (Production Best Practice)**

```
├── Access Token (15 minutes)
│   └── Storage: Memory (or zustand store)
│   └── When to use: Authorization header for API calls
│   └── Security: Lost on page refresh (safer)
│
├── Refresh Token (7 days)
│   └── Storage: httpOnly cookie (javascript cannot access)
│   └── When to use: Auto-renew access token when expired
│   └── Security: Protected from XSS attacks
│
└── CSRF Token (optional but recommended)
    └── Storage: localStorage
    └── When to use: For state-changing operations
```

**Why This Approach:**
- ✅ Access token in memory: Prevents XSS from stealing long-lived tokens
- ✅ Refresh token in httpOnly cookie: Backend sets, only sent over HTTPS, inaccessible to JavaScript
- ✅ Auto-refresh: User experience remains seamless even if token expires
- ✅ Logout security: Blacklist token on backend, clear cookies on frontend

**Implementation Flow:**

```javascript
// 1. Login → Get both tokens
POST /api/auth/verify-otp
Response: {
  data: {
    user: { ... },
    token_pair: {
      access_token: "eyJhbGc...",      // 15min expiry
      refresh_token: "refresh_eyJ...", // 7d expiry
      token_type: "Bearer"
    }
  }
}

// Frontend:
// - Store access_token in memory/zustand
// - Backend sets refresh_token in httpOnly cookie automatically
// - (No need to handle refresh token in frontend code for storing)

// 2. API Call → Attach access token
GET /api/user/profile
Authorization: Bearer eyJhbGc...

// 3. Token expires → Auto-refresh
GET /api/auth/refresh + refresh_token (in cookie)
Response: { new access_token }

// 4. Logout → Clear everything
POST /api/auth/logout
// Backend: Blacklist token + invalidate session
// Frontend: Clear memory, cookies cleared by Set-Cookie header
```

---

### 2. State Management Architecture

**Decision: TanStack Query (React Query) + Zustand**

**Why This Combination:**
- TanStack Query: Server state (API data, caching, invalidation)
- Zustand: Client state (UI, user preferences, auth)

**Setup:**

```typescript
// lib/queryClient.ts
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10,   // 10 minutes (formerly cacheTime)
      retry: 1,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
    mutations: {
      retry: 1,
    },
  },
});

// stores/authStore.ts - Zustand
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
  accessToken: string | null;
  user: User | null;
  setAccessToken: (token: string) => void;
  setUser: (user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      user: null,
      setAccessToken: (token) => set({ accessToken: token }),
      setUser: (user) => set({ user }),
      logout: () => set({ accessToken: null, user: null }),
    }),
    {
      name: 'auth-storage', // localStorage key
      partialize: (state) => ({ user: state.user }), // Only persist user, not token
    }
  )
);

// stores/uiStore.ts - For UI state
import { create } from 'zustand';

interface UIState {
  theme: 'light' | 'dark';
  sidebarOpen: boolean;
  setTheme: (theme: 'light' | 'dark') => void;
  toggleSidebar: () => void;
}

export const useUIStore = create<UIState>((set) => ({
  theme: 'light',
  sidebarOpen: true,
  setTheme: (theme) => set({ theme }),
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
}));
```

**Server State Management (TanStack Query):**

```typescript
// hooks/useUser.ts
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export function useUser() {
  return useQuery({
    queryKey: ['user', 'profile'],
    queryFn: async () => {
      const response = await api.get('/user/profile');
      return response.data.data;
    },
    enabled: !!useAuthStore.getState().accessToken, // Only if logged in
  });
}

// hooks/useUpdateProfile.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';

export function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (profileData) => {
      const response = await api.put('/user/profile', profileData);
      return response.data.data;
    },
    onSuccess: (data) => {
      // Invalidate and refetch user profile
      queryClient.setQueryData(['user', 'profile'], data);
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
    onError: (error) => {
      console.error('Profile update failed:', error);
    },
  });
}
```

---

### 3. Error Handling Strategy (Production-Grade)

**Decision: Centralized Error Handler with Retry Logic**

```typescript
// lib/errorHandler.ts
import axios, { AxiosError } from 'axios';

enum ErrorSeverity {
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
  CRITICAL = 'critical',
}

interface AppError {
  code: string;
  message: string;
  severity: ErrorSeverity;
  fieldErrors?: Record<string, string>;
  retryable: boolean;
  statusCode: number;
}

export function parseApiError(error: unknown): AppError {
  if (axios.isAxiosError(error)) {
    const response = error.response?.data;

    // Handle our standardized error format
    if (response?.success === false) {
      return {
        code: response.error?.code || 'UNKNOWN_ERROR',
        message: response.error?.message || 'An unexpected error occurred',
        severity: getSeverity(error.response?.status),
        fieldErrors: response.details,
        retryable: isRetryable(error.response?.status),
        statusCode: error.response?.status || 500,
      };
    }
  }

  // Fallback for unexpected errors
  return {
    code: 'NETWORK_ERROR',
    message: 'Network error. Please check your connection.',
    severity: ErrorSeverity.ERROR,
    retryable: true,
    statusCode: 0,
  };
}

function getSeverity(status?: number): ErrorSeverity {
  if (!status) return ErrorSeverity.ERROR;
  if (status < 400) return ErrorSeverity.INFO;
  if (status < 500) return ErrorSeverity.WARNING;
  return ErrorSeverity.CRITICAL;
}

function isRetryable(status?: number): boolean {
  if (!status) return true;
  // Retry on: 408 (timeout), 429 (rate limit), 5xx (server error)
  return status === 408 || status === 429 || status >= 500;
}

// Show error to user (integrate with your toast/notification system)
export function handleError(error: AppError) {
  switch (error.severity) {
    case ErrorSeverity.CRITICAL:
      // Show banner, log to Sentry
      console.error('[CRITICAL]', error);
      break;
    case ErrorSeverity.ERROR:
      // Show toast error
      break;
    case ErrorSeverity.WARNING:
      // Show warning toast
      break;
    case ErrorSeverity.INFO:
      // Show info toast or silent
      break;
  }
}
```

---

### 4. Rate Limiting Handling

**Decision: Exponential Backoff + User-Friendly Messaging**

```typescript
// lib/api.ts - Axios instance with retry logic
import axios from 'axios';
import { useAuthStore } from '@/stores/authStore';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  timeout: 30000, // 30 seconds
  withCredentials: true, // Important: Allow cookies (refresh token)
});

// Request interceptor: Attach access token
api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: Handle token refresh + rate limiting
let isRefreshing = false;
let failedQueue: any[] = [];

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Handle 401 - Refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // Queue request while refreshing
        return new Promise((resolve, reject) => {
          failedQueue.push(() => {
            resolve(api(originalRequest));
          });
        });
      }

      isRefreshing = true;
      originalRequest._retry = true;

      try {
        // Refresh token (request includes httpOnly cookie automatically)
        const response = await axios.post(
          `${process.env.NEXT_PUBLIC_API_URL}/auth/refresh`
        );
        
        const { access_token } = response.data.data.token_pair;
        useAuthStore.getState().setAccessToken(access_token);

        // Retry original request with new token
        api(originalRequest).then(() => {
          failedQueue.forEach((callback) => callback());
          failedQueue = [];
        });

        isRefreshing = false;
        return api(originalRequest);
      } catch (refreshError) {
        // Refresh failed - logout user
        useAuthStore.getState().logout();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // Handle 429 - Rate Limit Exceeded
    if (error.response?.status === 429) {
      const retryAfter = parseInt(error.response.headers['retry-after'] || '60');
      
      // Show user-friendly message
      console.warn(`Rate limited. Retry after ${retryAfter}s`);
      
      // Auto-retry after suggested delay
      if (!originalRequest._retryCount) {
        originalRequest._retryCount = 0;
      }
      
      if (originalRequest._retryCount < 3) {
        originalRequest._retryCount++;
        await new Promise((resolve) => setTimeout(resolve, retryAfter * 1000));
        return api(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);

export default api;
```

---

## Authentication & Token Management

### Complete Login Flow

```
┌─────────────────────────────────────────────────────────┐
│ 1. Registration (POST /api/auth/register)               │
├─────────────────────────────────────────────────────────┤
│ Request:                                                 │
│ {                                                        │
│   "first_name": "John",                                 │
│   "last_name": "Doe",                                   │
│   "username": "john_doe",                               │
│   "email": "john@example.com",                          │
│   "password": "SecurePass123!"  ← Min 8 chars           │
│ }                                                        │
│                                                          │
│ Response (201 Created):                                 │
│ {                                                        │
│   "success": true,                                      │
│   "data": {                                             │
│     "user_id": "uuid",                                  │
│     "email": "...",                                     │
│     "is_verified": false,  ← Key: Not verified yet      │
│     "status": "inactive"   ← Awaiting verification      │
│   }                                                      │
│ }                                                        │
│                                                          │
│ Next Step: User receives verification email             │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 2. Request OTP (POST /api/auth/request-otp)             │
├─────────────────────────────────────────────────────────┤
│ Request:                                                 │
│ {                                                        │
│   "email": "john@example.com",                          │
│   "password": "SecurePass123!"                          │
│ }                                                        │
│                                                          │
│ Response (200 OK):                                      │
│ {                                                        │
│   "success": true,                                      │
│   "data": {                                             │
│     "expires_in": 300,  ← 5 minutes validity            │
│     "message": "OTP sent to your email"                 │
│   }                                                      │
│ }                                                        │
│                                                          │
│ ⚠️ Important: OTP is NEVER in response for security     │
│ User gets 6-digit code via email                        │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Verify OTP & Login (POST /api/auth/verify-otp)       │
├─────────────────────────────────────────────────────────┤
│ Request:                                                 │
│ {                                                        │
│   "email": "john@example.com",                          │
│   "otp": "123456"  ← 6-digit code from email           │
│ }                                                        │
│                                                          │
│ Response (200 OK):                                      │
│ {                                                        │
│   "success": true,                                      │
│   "data": {                                             │
│     "user": {                                           │
│       "user_id": "uuid",                                │
│       "email": "john@example.com",                      │
│       "is_verified": true,  ← Now verified              │
│       "status": "active"    ← Account active            │
│     },                                                   │
│     "token_pair": {                                     │
│       "access_token": "eyJhbGc...",  ← 15 min expiry   │
│       "refresh_token": "refresh_...", ← 7 day expiry   │
│       "token_type": "Bearer"                            │
│     }                                                    │
│   }                                                      │
│ }                                                        │
│                                                          │
│ Frontend Action:                                        │
│ 1. Store access_token in Zustand                        │
│ 2. Refresh token auto-stored in httpOnly cookie         │
│ 3. Store user data in Zustand                           │
│ 4. Redirect to /dashboard/onboarding                    │
└─────────────────────────────────────────────────────────┘
```

### Token Refresh Flow (Automatic)

```typescript
// When access token expires (15 minutes)
// Automatically happens on next API call

GET /api/user/profile (with expired token)
  ↓
Response: 401 Unauthorized
  ↓
Frontend detects 401 → Triggers refresh
  ↓
POST /api/auth/refresh (refresh token in cookie)
  ↓
Response: {
  "data": {
    "token_pair": {
      "access_token": "new_token",
      "refresh_token": "new_refresh"
    }
  }
}
  ↓
Frontend updates access token in Zustand
  ↓
Retries original request with new token
  ↓
Request succeeds!
```

---

## API Integration Setup

### Environment Configuration

```bash
# .env.local (Development)
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Refyne

# .env.production (Production - Vercel)
NEXT_PUBLIC_API_URL=https://api.refyne.app
NEXT_PUBLIC_APP_NAME=Refyne
```

**Important Notes:**
- `NEXT_PUBLIC_*` prefix makes variables accessible in browser
- Never put sensitive secrets in these (no API keys, no secrets)
- Refresh tokens are handled via httpOnly cookies (secure, server-side)

### API Client Setup

```typescript
// lib/api.ts - Already shown above with interceptors
// lib/apiTypes.ts - Type definitions

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  request_id: string;
  timestamp: string;
  error?: {
    code: string;
    message: string;
  };
  details?: Record<string, string>;
}

export interface User {
  user_id: string;
  email: string;
  first_name: string;
  last_name: string;
  username: string;
  is_verified: boolean;
  is_active: boolean;
  status: 'inactive' | 'active';
  onboarding_completed: boolean;
  subscription_status: string;
  subscription_tier: string | null;
  created_at: string;
  updated_at?: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  token_type: 'Bearer';
}
```

---

## User Onboarding Flow

### Complete Onboarding Journey

```
Step 1: Email Verification (Auto after OTP login)
│
├─ Status: is_verified = true ✓
├─ Card Shows: "Email verified!" (success state)
└─ Next: Profile Setup

Step 2: Profile Setup (POST /api/user/profile)
│
├─ Fields:
│  ├─ First Name ✓ (from registration, editable)
│  ├─ Last Name ✓ (from registration, editable)
│  ├─ Username ✓ (from registration, editable)
│  └─ Bio (optional)
│
├─ Validation:
│  ├─ First Name: Required, 2-50 chars
│  ├─ Last Name: Required, 2-50 chars
│  ├─ Username: Required, 3-30 chars, alphanumeric + underscore
│  └─ Bio: Optional, max 500 chars
│
└─ On Success: Move to Step 3

Step 3: Preferences Setup (POST /api/user/settings)
│
├─ Fields:
│  ├─ Language: Select (English, Spanish, French, etc.)
│  ├─ Timezone: Select (Auto-detect or manual)
│  ├─ Email Notifications: Toggle
│  │  ├─ Marketing emails
│  │  ├─ Account updates
│  │  └─ Weekly digest
│  └─ Theme: Light / Dark (optional for onboarding)
│
├─ Validation:
│  ├─ Language: Must be valid locale
│  ├─ Timezone: Must be valid IANA timezone
│  └─ Booleans: Accept true/false
│
└─ On Success: Move to Step 4

Step 4: Create First Workspace (POST /api/workspaces)
│
├─ Fields:
│  ├─ Workspace Name: Required (e.g., "My Team")
│  └─ Description: Optional
│
├─ Validation:
│  ├─ Name: Required, 2-100 chars
│  └─ Description: Optional, max 500 chars
│
├─ Backend Behavior:
│  └─ Auto-creates with user as Owner
│
└─ On Success: Move to Step 5

Step 5: Start with FREE Tier (Optional Upgrade)
│
├─ Show: Dismissable "Upgrade to Pro" card with benefits
├─ User has 2 options:
│  ├─ "Upgrade to Pro Now" → POST /api/subscriptions/checkout → Paddle checkout
│  └─ "Start with Free" → Proceed to dashboard as FREE user
│
├─ Starting as FREE means:
│  ├─ ✅ 1 Workspace
│  ├─ ✅ Instagram connection & basic sync
│  ├─ ✅ Manual posting
│  ├─ ✅ 5 Otto AI requests/month
│  ├─ ❌ Scheduled posting (Pro only)
│  ├─ ❌ Unlimited workspaces (Pro only)
│  └─ ❌ Advanced AI (Pro only)
│
└─ Upgrade available anytime in dashboard → Settings

Step 6: Mark Onboarding Complete (POST /api/user/onboarding)
│
├─ Request: {} (empty body)
├─ Response: { "status": "completed" }
│
└─ Frontend Action:
   ├─ Set onboarding_completed = true
   ├─ Redirect to /dashboard
   └─ Show welcome modal
```

### Implementation Example

```typescript
// app/onboarding/page.tsx
'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import { useMutation } from '@tanstack/react-query';
import api from '@/lib/api';

type OnboardingStep = 'profile' | 'preferences' | 'workspace' | 'subscription' | 'complete';

export default function OnboardingPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [currentStep, setCurrentStep] = useState<OnboardingStep>('profile');

  // Profile mutation
  const updateProfileMutation = useMutation({
    mutationFn: (data) => api.put('/user/profile', data),
    onSuccess: () => setCurrentStep('preferences'),
  });

  // Settings mutation
  const updateSettingsMutation = useMutation({
    mutationFn: (data) => api.put('/user/settings', data),
    onSuccess: () => setCurrentStep('workspace'),
  });

  // Workspace creation
  const createWorkspaceMutation = useMutation({
    mutationFn: (data) => api.post('/workspaces', data),
    onSuccess: () => setCurrentStep('subscription'),
  });

  // Complete onboarding
  const completeOnboardingMutation = useMutation({
    mutationFn: () => api.post('/user/onboarding', {}),
    onSuccess: () => {
      useAuthStore.getState().setUser({ ...user, onboarding_completed: true });
      router.push('/dashboard');
    },
  });

  const handleSkipToComplete = () => {
    completeOnboardingMutation.mutate();
  };

  return (
    <div className="onboarding-container">
      {currentStep === 'profile' && (
        <ProfileStep onNext={(data) => updateProfileMutation.mutate(data)} />
      )}
      {currentStep === 'preferences' && (
        <PreferencesStep onNext={(data) => updateSettingsMutation.mutate(data)} />
      )}
      {currentStep === 'workspace' && (
        <WorkspaceStep onNext={(data) => createWorkspaceMutation.mutate(data)} />
      )}
      {currentStep === 'subscription' && (
        <SubscriptionStep 
          onUpgrade={() => window.location.href = checkoutUrl}
          onSkip={handleSkipToComplete}  // User can skip and access dashboard as FREE user
        />
      )}
    </div>
  );
}
```

---

## Freemium Model & Feature Gating

### Feature Matrix: FREE vs PRO

```
FEATURE                          | FREE TIER  | PRO TIER
─────────────────────────────────┼────────────┼──────────
Workspaces                       | 1          | Unlimited
Instagram Connections           | ✅ Yes     | ✅ Yes
Instagram Media Sync            | ✅ Yes     | ✅ Yes (Advanced)
Manual Instagram Posting        | ✅ Yes     | ✅ Yes
Scheduled Instagram Posting     | ❌ No      | ✅ Yes
Otto AI Requests/Month          | 5          | Unlimited
Otto AI Advanced Features       | ❌ No      | ✅ Yes (Content generation, auto-posting)
Team Members per Workspace      | 1 (Owner)  | Up to 10
Basic Analytics                 | ✅ Yes     | ✅ Yes
Advanced Analytics & Reports    | ❌ No      | ✅ Yes
Priority Support                | ❌ No      | ✅ Yes
```

### KEY: When to Gate Features

**1. In Frontend (Better UX)**
```typescript
// Lock UI for Pro-only features
if (user.subscription_tier !== "pro") {
  return <UpgradePrompt feature="Scheduled Posting" />;
}
// Show the feature
return <FeatureComponent />;
```

**2. In Backend (For Security)**
```go
// Backend validates subscription before allowing action
if err := h.subscriptionRepo.IsProUser(ctx, userID); err != nil {
  return errors.New("feature_requires_pro_subscription")
}
// Process the feature
```

**Frontend MUST gate because:**
- Better user experience (show prompts instead of errors)
- Reduces API calls
- Improved performance

**Backend MUST validate because:**
- Security (user can't bypass frontend checks)
- Prevents unauthorized Pro feature access
- Enforces subscription rules

### Implementation: Feature Gate Component

```typescript
// components/ProFeatureGate.tsx
import { useSubscription } from '@/hooks/useSubscription';
import { useRouter } from 'next/navigation';

interface ProFeatureGateProps {
  feature: string;
  description?: string;
  children: React.ReactNode;
}

export function ProFeatureGate({ feature, description, children }: ProFeatureGateProps) {
  const { data: subscription } = useSubscription();
  const router = useRouter();

  const isProUser = subscription?.tier === 'pro' && subscription?.status === 'active';

  if (isProUser) {
    return <>{children}</>;
  }

  return (
    <div className="pro-feature-locked border-2 border-amber-200 bg-amber-50 p-6 rounded-lg">
      <div className="flex items-center gap-3">
        <span className="text-2xl">🔒</span>
        <div>
          <h3 className="font-semibold">Upgrade to Pro</h3>
          <p className="text-sm text-gray-600">
            {description || `${feature} is only available for Pro users`}
          </p>
        </div>
      </div>
      <button
        onClick={() => router.push('/settings/subscription')}
        className="mt-3 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
      >
        Upgrade to Pro
      </button>
    </div>
  );
}

// Usage:
export function ScheduledPostingPage() {
  return (
    <ProFeatureGate 
      feature="Scheduled Posting"
      description="Schedule posts to publish at optimal times - Pro only"
    >
      <ScheduledPostingForm />
    </ProFeatureGate>
  );
}
```

### Implementation: Per-Workspace Limit Gate

```typescript
// hooks/useWorkspaceCheck.ts
import { useQuery } from '@tanstack/react-query';
import { useSubscription } from './useSubscription';
import api from '@/lib/api';

export function useCanCreateWorkspace() {
  const { data: subscription } = useSubscription();
  const { data: workspaces } = useQuery({
    queryKey: ['workspaces'],
    queryFn: async () => {
      const response = await api.get('/workspaces');
      return response.data.data.workspaces;
    },
  });

  // Free users limited to 1 workspace
  if (subscription?.tier !== 'pro') {
    return {
      canCreate: !workspaces || workspaces.length < 1,
      reason: workspaces?.length > 0 
        ? 'Free users can only create 1 workspace. Upgrade to Pro for unlimited.' 
        : null,
    };
  }

  // Pro users can create unlimited
  return { canCreate: true, reason: null };
}

// Component usage:
export function CreateWorkspaceModal() {
  const { canCreate, reason } = useCanCreateWorkspace();

  if (!canCreate) {
    return <UpgradePrompt message={reason} />;
  }

  return <WorkspaceForm />;
}
```

### Implementation: Otto AI Request Counter

```typescript
// hooks/useOttoAIRequests.ts
import { useQuery } from '@tanstack/react-query';
import { useSubscription } from './useSubscription';

const FREE_TIER_MONTHLY_LIMIT = 5;
const PRO_TIER_MONTHLY_LIMIT = Infinity; // Unlimited

export function useOttoAIRequests() {
  const { data: subscription } = useSubscription();
  const { data: usage } = useQuery({
    queryKey: ['otto', 'usage', new Date().toISOString().slice(0, 7)],
    queryFn: async () => {
      const response = await api.get('/otto/usage/current-month');
      return response.data.data;
    },
  });

  const limit = subscription?.tier === 'pro' 
    ? PRO_TIER_MONTHLY_LIMIT 
    : FREE_TIER_MONTHLY_LIMIT;

  return {
    used: usage?.request_count || 0,
    limit,
    remaining: Math.max(0, limit - (usage?.request_count || 0)),
    isLimitReached: (usage?.request_count || 0) >= limit,
  };
}

// Component usage:
export function OttoAIChat() {
  const { remaining, isLimitReached, limit } = useOttoAIRequests();
  const { data: subscription } = useSubscription();

  if (isLimitReached && subscription?.tier !== 'pro') {
    return (
      <ProFeatureGate 
        feature="Otto AI"
        description={`You've reached your monthly limit of ${limit} requests. Upgrade to Pro for unlimited AI requests.`}
      >
        <ChatForm disabled />
      </ProFeatureGate>
    );
  }

  return (
    <>
      <ChatForm />
      <p className="text-sm text-gray-500 mt-2">
        {remaining} requests remaining this month
      </p>
    </>
  );
}
```

### Implementation: Team Member Limit Gate

```typescript
// hooks/useTeamMemberLimit.ts
import { useQuery } from '@tanstack/react-query';
import { useSubscription } from './useSubscription';
import api from '@/lib/api';

export function useTeamMemberLimit(workspaceId: string) {
  const { data: subscription } = useSubscription();
  const { data: members } = useQuery({
    queryKey: ['workspace', workspaceId, 'members'],
    queryFn: async () => {
      const response = await api.get(`/workspaces/${workspaceId}/members`);
      return response.data.data.members;
    },
  });

  const limit = subscription?.tier === 'pro' ? 10 : 1; // Free: owner only, Pro: up to 10
  const canInviteMore = (members?.length || 0) < limit;

  return {
    currentMembers: members?.length || 0,
    limit,
    canInviteMore,
    spotsRemaining: Math.max(0, limit - (members?.length || 0)),
  };
}

// Component usage:
export function InviteTeamModal({ workspaceId }: { workspaceId: string }) {
  const { canInviteMore, spotsRemaining, limit } = useTeamMemberLimit(workspaceId);
  const { data: subscription } = useSubscription();

  if (!canInviteMore && subscription?.tier !== 'pro') {
    return (
      <ProFeatureGate
        feature="Team Collaboration"
        description={`Free users can only have 1 team member (workspace owner). Upgrade to Pro to invite up to ${limit} members.`}
      >
        <InviteForm disabled />
      </ProFeatureGate>
    );
  }

  return (
    <>
      <InviteForm />
      {subscription?.tier === 'pro' && (
        <p className="text-sm text-gray-500">
          {spotsRemaining} spots remaining out of {limit}
        </p>
      )}
    </>
  );
}
```

### Dashboard: Upgrade Prompts Location

```typescript
// components/DashboardLayout.tsx
export function DashboardLayout() {
  const { data: subscription } = useSubscription();
  const isFreeTier = subscription?.tier !== 'pro';

  return (
    <div className="dashboard">
      {/* Banner at top for free users */}
      {isFreeTier && (
        <div className="sticky top-0 z-40 bg-gradient-to-r from-blue-600 to-blue-700 text-white p-4 flex items-center justify-between">
          <div>
            <p className="font-semibold">🚀 Unlock Pro Features</p>
            <p className="text-sm opacity-90">Get unlimited AI requests, scheduled posting, and more</p>
          </div>
          <button 
            onClick={() => router.push('/settings/subscription')}
            className="bg-white text-blue-600 px-6 py-2 rounded font-semibold hover:bg-gray-100"
          >
            Upgrade Now
          </button>
        </div>
      )}

      {/* Main content */}
      <div className="flex gap-4 p-6">
        <Sidebar />
        <MainContent />

        {/* Sidebar card for free users */}
        {isFreeTier && (
          <div className="w-80 border border-amber-200 bg-amber-50 rounded-lg p-4">
            <h3 className="font-semibold flex items-center gap-2">
              <span>⭐</span> Pro Benefits
            </h3>
            <ul className="mt-3 space-y-2 text-sm">
              <li>✅ Unlimited workspaces</li>
              <li>✅ Unlimited AI requests</li>
              <li>✅ Scheduled posting</li>
              <li>✅ Team collaboration (10 members)</li>
              <li>✅ Advanced analytics</li>
              <li>✅ Priority support</li>
            </ul>
            <button 
              onClick={() => router.push('/settings/subscription')}
              className="w-full mt-4 bg-blue-600 text-white py-2 rounded hover:bg-blue-700"
            >
              Upgrade to Pro
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
```

---

### Authentication Endpoints

#### 1. Register User
```
POST /api/auth/register
Content-Type: application/json

Request:
{
  "first_name": "John",
  "last_name": "Doe",
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}

Response 201 Created:
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user_id": "uuid",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "username": "john_doe",
    "is_verified": false,
    "is_active": false,
    "status": "inactive",
    "onboarding_completed": false,
    "subscription_status": "free",
    "created_at": "2026-04-18T..."
  }
}

Errors:
- 400: VALIDATION_ERROR (invalid email, weak password, etc.)
- 409: DUPLICATE_EMAIL or DUPLICATE_USERNAME
```

#### 2. Request OTP
```
POST /api/auth/request-otp
Content-Type: application/json

Request:
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}

Response 200 OK:
{
  "success": true,
  "message": "OTP sent successfully",
  "data": {
    "expires_in": 300,
    "message": "OTP sent successfully to your email"
  }
}

Errors:
- 400: INVALID_CREDENTIALS (wrong password)
- 404: NOT_FOUND (email not registered)
- 409: ALREADY_VERIFIED (account already verified)

⚠️ Important:
- OTP is sent to email, never returned in response
- User has 5 minutes to enter OTP
- After 3 failed attempts: account locked for 15 minutes
- Resend OTP: use /api/auth/resend-verification
```

#### 3. Verify OTP & Login
```
POST /api/auth/verify-otp
Content-Type: application/json

Request:
{
  "email": "john@example.com",
  "otp": "123456"
}

Response 200 OK:
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "user_id": "uuid",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "username": "john_doe",
      "status": "active",
      "is_verified": true
    },
    "token_pair": {
      "access_token": "eyJhbGc...",
      "refresh_token": "refresh_eyJ...",
      "token_type": "Bearer"
    }
  }
}

Errors:
- 400: INVALID_OTP (wrong 6-digit code)
- 401: OTP_EXPIRED (more than 5 min)
- 429: ACCOUNT_LOCKED (exceeded attempts, try in 15 min)

Frontend Actions:
1. Store access_token in Zustand
2. Set Authorization header for future requests
3. Refresh token stored in httpOnly cookie (automatic)
4. Redirect to onboarding if onboarding_completed=false
5. Redirect to /dashboard if onboarding_completed=true
```

#### 4. Refresh Token
```
POST /api/auth/refresh
Authorization: Bearer (old_token or leave empty)
Cookie: (refresh_token automatically sent)

Response 200 OK:
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "token_pair": {
      "access_token": "new_eyJhbGc...",
      "refresh_token": "new_refresh_eyJ...",
      "token_type": "Bearer"
    }
  }
}

Errors:
- 401: TOKEN_EXPIRED (refresh token expired, must login again)
- 400: INVALID_REQUEST (no refresh token found)

⚠️ Automatic:
- Backend will set new refresh_token in httpOnly cookie
- Frontend must update access_token in Zustand
```

#### 5. Logout
```
POST /api/auth/logout
Authorization: Bearer {access_token}
Cookie: (refresh_token in cookie)

Response 200 OK:
{
  "success": true,
  "message": "Logged out successfully",
  "data": {
    "status": "logged_out"
  }
}

Frontend Actions:
1. Clear Zustand auth state (accessToken, user)
2. Remove any cached queries
3. Redirect to /login
4. Cookies automatically cleared by Set-Cookie header
```

#### 6. Logout All Devices
```
POST /api/auth/logout-all-devices
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Logged out from all devices successfully",
  "data": {
    "status": "logged_out_all_devices"
  }
}

⚠️ Important:
- Blacklists ALL refresh tokens for this user
- User must sign in again on all devices
- Useful for security concerns
```

#### 7. Verify Account (Email Verification Link)
```
POST /api/auth/verify-account
Content-Type: application/json

Request:
{
  "token": "verification_token_from_email_link"
}

Response 200 OK:
{
  "success": true,
  "message": "Account verified successfully",
  "data": {
    "status": "verified"
  }
}

Errors:
- 400: INVALID_TOKEN (token not found or expired)
- 409: ALREADY_VERIFIED (account already verified)

⚠️ Flow:
1. User registers
2. Email with link: /verify-account?token=xyz
3. User clicks → Frontend calls this endpoint
4. Backend verifies → Sets is_verified=true
5. OTP login becomes available
```

#### 8. Resend Verification Email
```
POST /api/auth/resend-verification
Content-Type: application/json

Request:
{
  "email": "john@example.com"
}

Response 200 OK:
{
  "success": true,
  "message": "Verification email sent",
  "data": {
    "message": "If an account exists with that email, a verification email has been sent."
  }
}

⚠️ Security:
- Always returns success (doesn't reveal if email exists)
- Rate-limited: max 3 requests per hour per email
- Email sent with new verification link
```

#### 9. Forgot Password
```
POST /api/auth/forgot-password
Content-Type: application/json

Request:
{
  "email": "john@example.com"
}

Response 200 OK:
{
  "success": true,
  "message": "Password reset email sent",
  "data": {
    "message": "If the email exists, a password reset link has been sent"
  }
}

⚠️ Security:
- Always returns success (doesn't reveal if email exists)
- Email with reset link: /reset-password?token=xyz
- Link expires in 24 hours
- Can only be used once
```

#### 10. Reset Password
```
POST /api/auth/reset-password
Content-Type: application/json

Request:
{
  "token": "reset_token_from_email",
  "new_password": "NewSecurePass456!"
}

Response 200 OK:
{
  "success": true,
  "message": "Password reset successfully",
  "data": {
    "status": "password_reset"
  }
}

Errors:
- 400: INVALID_TOKEN (token expired or invalid)
- 400: WEAK_PASSWORD (password doesn't meet requirements)

Frontend Flow:
1. User clicks reset link from email
2. Form: Enter new password
3. Submit to /api/auth/reset-password
4. Success → Redirect to login
5. User logs in with new password
```

#### 11. Validate Reset Token
```
POST /api/auth/validate-reset-token
Content-Type: application/json

Request:
{
  "token": "reset_token_from_email"
}

Response 200 OK:
{
  "success": true,
  "message": "Token is valid",
  "data": {
    "user_id": "uuid"
  }
}

Errors:
- 400: INVALID_TOKEN (token expired or invalid)

Use Case:
- Before showing password reset form
- Verify token is still valid
- Pre-fill which user is resetting
```

---

### User Management Endpoints

#### Get Profile
```
GET /api/user/profile
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "user_id": "uuid",
    "first_name": "John",
    "last_name": "Doe",
    "username": "john_doe",
    "email": "john@example.com",
    "bio": "Software engineer",
    "avatar_url": "https://...",
    "is_verified": true,
    "is_active": true,
    "status": "active",
    "created_at": "2026-04-18T...",
    "updated_at": "2026-04-18T..."
  }
}

Use Case: Display user info, populate edit forms
Query Key for TanStack Query: ['user', 'profile']
Refetch on: User login, after profile update
```

#### Update Profile
```
PUT /api/user/profile
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "first_name": "John",
  "last_name": "Smith",
  "username": "john_smith",
  "bio": "Updated bio"
}

Response 200 OK:
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "user_id": "uuid",
    "first_name": "John",
    "last_name": "Smith",
    "username": "john_smith",
    ...
  }
}

Validation:
- first_name: 2-50 chars
- last_name: 2-50 chars
- username: 3-30 chars, alphanumeric + underscore, unique
- bio: max 500 chars

Possible Errors:
- 400: VALIDATION_ERROR (invalid field values)
- 409: DUPLICATE_USERNAME (username taken)

Frontend Actions:
- Invalidate ['user', 'profile'] query
- Update Zustand user state
- Show success toast
```

#### Get Settings
```
GET /api/user/settings
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Settings retrieved successfully",
  "data": {
    "user_id": "uuid",
    "language": "en",
    "timezone": "America/New_York",
    "email_notifications": {
      "marketing": true,
      "updates": true,
      "weekly_digest": false
    },
    "theme": "light"
  }
}

Query Key: ['user', 'settings']
Use Case: Display settings form, apply theme/language preferences
```

#### Update Settings
```
PUT /api/user/settings
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "language": "es",
  "timezone": "Europe/London",
  "email_notifications": {
    "marketing": false,
    "updates": true,
    "weekly_digest": true
  },
  "theme": "dark"
}

Response 200 OK:
{
  "success": true,
  "message": "Settings updated successfully",
  "data": {
    ...updated settings...
  }
}

Valid Languages: en, es, fr, de, pt, ja, zh, etc.
Valid Timezones: IANA timezone database (e.g., America/New_York, Europe/London, Asia/Tokyo)

Frontend Actions:
- Update UI theme immediately (don't wait for backend)
- Invalidate settings query
- Show success toast
- Refresh language strings if language changed
```

#### Complete Onboarding
```
POST /api/user/onboarding
Authorization: Bearer {access_token}
Content-Type: application/json

Request: {} (empty body)

Response 200 OK:
{
  "success": true,
  "message": "Onboarding completed successfully",
  "data": {
    "status": "completed"
  }
}

⚠️ Important:
- Can ONLY be called once (after all onboarding steps)
- Sets onboarding_completed = true in database
- User can't go back to onboarding after this
- Frontend should cache user state before redirecting

Frontend Logic:
1. After all steps complete & subscription handled
2. Call this endpoint
3. Update Zustand: user.onboarding_completed = true
4. Redirect to /dashboard
```

#### Delete Account
```
DELETE /api/user/account
Authorization: Bearer {access_token}

Request: {} (empty body, or no body)

Response 200 OK:
{
  "success": true,
  "message": "Account deleted successfully",
  "data": {
    "status": "deleted"
  }
}

⚠️ CRITICAL:
- SOFT DELETE: Account marked as deleted, data retained for 30 days
- Can't be undone - user must contact support to restore
- All sessions invalidated
- All tokens blacklisted
- Email can be re-registered after 30 days

Frontend Flow:
1. Show warning modal: "This cannot be undone"
2. Ask for password confirmation
3. Call DELETE /api/user/account
4. On success:
   - Clear all auth state
   - Clear all local data
   - Logout from all devices
   - Redirect to /goodbye or /
5. Show success message: "Account deleted"

Optional: Show feedback form first
```

---

### Workspace Endpoints

#### Create Workspace
```
POST /api/workspaces
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "name": "Marketing Team",
  "description": "Our marketing department workspace"
}

Response 201 Created:
{
  "success": true,
  "message": "Workspace created successfully",
  "data": {
    "id": "ws_uuid",
    "name": "Marketing Team",
    "description": "Our marketing department workspace",
    "owner_id": "user_uuid",
    "created_at": "2026-04-18T..."
  }
}

Validation:
- name: Required, 2-100 chars
- description: Optional, max 500 chars

Backend Auto-Behavior:
- Sets requesting user as Owner
- Creates default roles: Owner, Member
- Creates workspace settings with defaults

Frontend Query Key: ['workspaces']
Invalidate after creation
```

#### List Workspaces
```
GET /api/workspaces
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Workspaces retrieved successfully",
  "data": {
    "workspaces": [
      {
        "id": "ws_001",
        "name": "Marketing Team",
        "description": "...",
        "owner_id": "user_uuid",
        "role": "owner",  /* User's role in this workspace */
        "created_at": "2026-04-18T...",
        "member_count": 5
      },
      {
        "id": "ws_002",
        "name": "Analytics",
        "description": "...",
        "owner_id": "different_user_uuid",
        "role": "member",  /* User is member here */
        "created_at": "2026-04-18T...",
        "member_count": 3
      }
    ]
  }
}

Frontend:
- Query Key: ['workspaces']
- Refetch on: workspace create, delete, update
- Use for: Workspace switcher, dashboard sidebar
```

#### Get Workspace
```
GET /api/workspaces/:id
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Workspace retrieved successfully",
  "data": {
    "id": "ws_001",
    "name": "Marketing Team",
    "description": "...",
    "owner_id": "uuid",
    "role": "owner",  /* User's role */
    "created_at": "2026-04-18T...",
    "settings": {
      "archived": false,
      "public": false
    }
  }
}

Query Key: ['workspaces', workspaceId]
Use Case: Workspace details page, permissions check
```

#### Update Workspace
```
PUT /api/workspaces/:id
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "name": "New Team Name",
  "description": "Updated description"
}

Response 200 OK:
{
  "success": true,
  "message": "Workspace updated successfully",
  "data": {
    "id": "ws_001",
    "name": "New Team Name",
    ...
  }
}

⚠️ Permissions:
- Only Owner can update workspace
- Frontend should check role before showing edit button

Errors:
- 403: FORBIDDEN (not owner)
- 404: NOT_FOUND (workspace doesn't exist)
```

#### Delete Workspace
```
DELETE /api/workspaces/:id
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Workspace deleted successfully",
  "data": {
    "status": "deleted"
  }
}

⚠️ Important:
- Only Owner can delete
- HARD DELETE (cascades to members, invitations)
- Can't be undone
- All members lose access immediately

Frontend:
- Show warning modal
- Ask for confirmation
- On success: Invalidate ['workspaces'], redirect to other workspace or onboarding
```

#### List Workspace Members
```
GET /api/workspaces/:id/members
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Members retrieved successfully",
  "data": {
    "members": [
      {
        "id": "member_001",
        "user": {
          "user_id": "uuid",
          "email": "john@example.com",
          "first_name": "John",
          "last_name": "Doe"
        },
        "role": "owner",
        "joined_at": "2026-04-18T...",
        "status": "active"
      },
      {
        "id": "member_002",
        "user": {
          "user_id": "uuid2",
          "email": "jane@example.com",
          "first_name": "Jane",
          "last_name": "Smith"
        },
        "role": "member",
        "joined_at": "2026-04-18T...",
        "status": "active"
      }
    ]
  }
}

Query Key: ['workspaces', workspaceId, 'members']
Use Case: Members list page, member management
```

#### Invite Workspace Member
```
POST /api/workspaces/:id/members/invite
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "email": "newmember@example.com"
}

Response 200 OK:
{
  "success": true,
  "message": "Invitation sent successfully",
  "data": {
    "status": "invited"
  }
}

Validation:
- email: Must be valid email format
- email: Can't be workspace owner's email
- email: Can't invite same member twice

Errors:
- 400: VALIDATION_ERROR (invalid email)
- 409: ALREADY_MEMBER (user already in workspace)
- 403: FORBIDDEN (can't invite - insufficient permissions)

Backend Behavior:
- Sends invitation email to user
- Creates invite record (expires in 7 days)
- Invitation link: /join-workspace?invite_token=xyz

Frontend:
- Show success message
- Invalidate members list
- Optionally show "Pending invitations" section
```

#### Remove Workspace Member
```
DELETE /api/workspaces/:id/members/:memberId
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Member removed successfully",
  "data": {
    "status": "removed"
  }
}

Permissions:
- Only Owner can remove members
- Owner can't remove themselves (use delete workspace instead)

Errors:
- 403: FORBIDDEN (not owner)
- 400: INVALID_REQUEST (can't remove owner)
- 404: NOT_FOUND (member not found)

Frontend:
- Show confirmation modal
- Invalidate members list
- Remove member row from UI
```

---

### Subscription Endpoints

#### Create Checkout
```
POST /api/subscriptions/checkout
Authorization: Bearer {access_token}
Content-Type: application/json

Request: {} (empty body - always Pro tier)

Response 200 OK:
{
  "success": true,
  "message": "Checkout URL generated successfully",
  "data": {
    "checkout_url": "https://checkout.paddle.com/p/preview/...",
    "mode": "sandbox"  /* "sandbox" or "live" */
  }
}

⚠️ Important:
- Always generates Pro tier checkout (only tier available)
- Mode indicates if Paddle is in sandbox (test) or live (production)
- checkout_url is external - should open in new tab

Frontend Flow:
1. User clicks "Upgrade to Pro"
2. Call POST /api/subscriptions/checkout
3. Response includes checkout_url
4. Redirect: window.location.href = checkout_url
5. User completes payment on Paddle
6. Paddle webhook notifies backend
7. Backend creates subscription
8. User redirected back to app (/dashboard?subscription=success)
9. Frontend refetches subscription status

Testing (Sandbox Mode):
- Use test card: 4242 4242 4242 4242
- Any future expiry date
- Any 3-digit CVC
```

#### Get Subscription Status
```
GET /api/subscriptions/status
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Subscription status retrieved successfully",
  "data": {
    "tier": "pro",
    "status": "active",  /* "active", "expired", "cancelled" */
    "expires_at": "2027-04-18T23:59:59Z",
    "paddle_customer_id": "ctm_uuid",
    "management_portal_url": "https://app.paddle.com/billing/..."
  }
}

Query Key: ['subscription', 'status']
Refetch on:
- App load
- After checkout redirect
- Daily refresh (optional)

Frontend Use:
- Check tier for feature access (Pro-only features)
- Show expiry date if approaching renewal
- Link to management portal for billing
- Display subscription badge in dashboard
```

#### Get Customer Portal URL
```
GET /api/subscriptions/portal
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Customer portal URL generated successfully",
  "data": {
    "portal_url": "https://app.paddle.com/billing/..."
  }
}

⚠️ No subscription required:
- Returns portal URL even for free users
- Free users can view portal but have no subscriptions

Frontend:
- "Manage Subscription" link
- "Billing" link in settings
- Opens in new tab: window.open(portal_url)
```

#### Handle Webhook (Backend Only)
```
POST /api/subscriptions/webhook
(No Authorization header needed - verified via signature)
Content-Type: application/json
Paddle-Signature: signature_header

Request:
{
  "event_id": "evt_uuid",
  "event_type": "subscription.created",  /* or .updated, .cancelled */
  "data": {
    "subscription_id": "sub_uuid",
    "customer_id": "ctm_uuid",
    "status": "active"
  }
}

Response 200 OK:
{
  "success": true,
  "message": "Webhook processed successfully",
  "data": {
    "event_id": "evt_uuid"
  }
}

⚠️ Webhook Events:
- subscription.created → User purchased Pro
- subscription.updated → Billing cycle, changes
- subscription.cancelled → User cancelled subscription

Backend Actions:
1. Verify Paddle signature
2. Find user by paddle_customer_id
3. Update subscription in database
4. Invalidate user's subscription cache (for next API call)

Frontend:
- No frontend interaction needed
- Backend handles silently
- Next API call will reflect new status
```

---

### Instagram Integration Endpoints (18 endpoints)

#### 1. Connect Instagram Account (OAuth)
```
POST /api/instagram/auth/connect
Authorization: Bearer {access_token}
Content-Type: application/json

Request: {} (empty body)

Response 200 OK:
{
  "success": true,
  "message": "OAuth connection initiated",
  "data": {
    "auth_url": "https://instagram.com/oauth/authorize?client_id=...",
    "message": "Redirect to this URL to authorize with Instagram"
  },
  "request_id": "req_xyz",
  "timestamp": "2026-04-18T10:30:45Z"
}

State Flow:
1. Frontend clicks "Connect Instagram"
2. Get auth_url from this endpoint
3. Redirect user to auth_url (Instagram OAuth page)
4. User authorizes Refyne app
5. Instagram redirects to /api/instagram/auth/callback?code=...&state=...
6. Backend automatically creates connection
7. Frontend polls /api/instagram/accounts to verify connection

Frontend:
- Button: "Connect Instagram Account"
- On click: window.location.href = auth_url
```

#### 2. OAuth Callback (Frontend handles redirect)
```
GET /api/instagram/auth/callback?code=AUTH_CODE&state=STATE
(Automatic redirect from Instagram, no manual call)

Backend:
- Verifies code with Instagram API
- Creates InstagramAccount record for user
- Returns redirect to frontend dashboard

Frontend:
- After redirect completes, user sees success message
- Account appears in "Connected Accounts" list
```

#### 3. Disconnect Instagram Account
```
POST /api/instagram/auth/disconnect
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "account_id": "instagram_account_id_123"
}

Response 200 OK:
{
  "success": true,
  "message": "Instagram account disconnected",
  "data": {
    "status": "disconnected",
    "account_id": "instagram_account_id_123"
  },
  "request_id": "req_abc",
  "timestamp": "2026-04-18T10:30:45Z"
}

Error 404:
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Account not found"
  }
}

Frontend:
- Confirmation dialog: "Are you sure? Content won't sync anymore"
- On confirm: POST disconnect
- Remove account from list
```

#### 4. List Connected Instagram Accounts
```
GET /api/instagram/accounts
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Accounts retrieved successfully",
  "data": {
    "accounts": [
      {
        "id": "acc_123",
        "instagram_user_id": "987654321",
        "username": "john_doe",
        "profile_picture_url": "https://...",
        "biography": "Social media influencer",
        "followers_count": 5000,
        "connected_at": "2026-04-18T10:00:00Z",
        "last_sync_at": "2026-04-18T10:15:00Z",
        "sync_status": "idle",
        "sync_error_message": null
      }
    ]
  },
  "request_id": "req_def",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Display in Settings → Connected Accounts
- Show sync status (idle, syncing, error)
- Show last sync timestamp
- Disconnect button per account
```

#### 5. Get Single Account Details
```
GET /api/instagram/accounts/:id
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Account retrieved successfully",
  "data": {
    "id": "acc_123",
    "instagram_user_id": "987654321",
    "username": "john_doe",
    "profile_picture_url": "https://...",
    "biography": "Social media influencer",
    "followers_count": 5000,
    "connected_at": "2026-04-18T10:00:00Z",
    "last_sync_at": "2026-04-18T10:15:00Z",
    "sync_status": "idle",
    "sync_error_message": null
  },
  "request_id": "req_ghi",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Show on account detail page
- Display profile info, sync status
```

#### 6. Get Instagram Media (paginated)
```
GET /api/instagram/media?account_id=ACC_ID&limit=20&offset=0
Authorization: Bearer {access_token}

Query Parameters:
- account_id: Account ID (required)
- limit: 1-100, default 20
- offset: 0-based offset for pagination

Response 200 OK:
{
  "success": true,
  "message": "Media retrieved successfully",
  "data": {
    "media": [
      {
        "id": "media_456",
        "instagram_id": "18456789012345678",
        "account_id": "acc_123",
        "caption": "Beautiful sunset 🌅",
        "media_type": "IMAGE",
        "media_url": "https://instagram.com/p/ABC123/",
        "thumbnail_url": "https://instagram.com/media/...",
        "likes_count": 150,
        "comments_count": 25,
        "engagement_rate": 3.5,
        "posted_at": "2026-04-18T08:00:00Z",
        "synced_at": "2026-04-18T08:15:00Z"
      }
    ],
    "pagination": {
      "total": 250,
      "limit": 20,
      "offset": 0,
      "has_more": true
    }
  },
  "request_id": "req_jkl",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Display in Content Library/Gallery
- Use offset/limit for infinite scroll
- Show engagement metrics (likes, comments)
```

#### 7. Get Single Media Details
```
GET /api/instagram/media/:id
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Media retrieved successfully",
  "data": {
    "id": "media_456",
    "instagram_id": "18456789012345678",
    "account_id": "acc_123",
    "caption": "Beautiful sunset 🌅",
    "media_type": "IMAGE",
    "media_url": "https://instagram.com/p/ABC123/",
    "likes_count": 150,
    "comments_count": 25,
    "engagement_rate": 3.5,
    "posted_at": "2026-04-18T08:00:00Z"
  },
  "request_id": "req_mno",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Click on media in gallery → Show details page
```

#### 8. Get Media AI Recommendations (Pro Only)
```
GET /api/instagram/media/:id/ai
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "AI recommendations generated",
  "data": {
    "media_id": "media_456",
    "suggestions": {
      "caption_optimization": "Try adding a call-to-action at the end",
      "hashtag_recommendations": ["#sunset", "#nature", "#instagram", "#travel"],
      "best_posting_time": "Tuesday 6-7 PM",
      "engagement_forecast": "2-3x better engagement with optimizations"
    }
  },
  "request_id": "req_pqr",
  "timestamp": "2026-04-18T10:30:45Z"
}

Feature Gate: Pro Only
Frontend:
- Show in media detail page
- Display within <ProFeatureGate> component
```

#### 9. Get Account Analytics
```
GET /api/instagram/analytics?account_id=ACC_ID&period=7d
Authorization: Bearer {access_token}

Query Parameters:
- account_id: Account ID (required)
- period: 7d, 30d, 90d, 365d (default 30d)

Response 200 OK:
{
  "success": true,
  "message": "Analytics retrieved successfully",
  "data": {
    "account_id": "acc_123",
    "period": "30d",
    "followers_gained": 250,
    "followers_lost": 10,
    "engagement_rate": 4.2,
    "reach": 15000,
    "impressions": 25000,
    "profile_visits": 500,
    "website_clicks": 45,
    "growth_trend": 2.5
  },
  "request_id": "req_stu",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Display in Analytics Dashboard
- Show graphs/charts with metrics
```

#### 10. Get Media Analytics
```
GET /api/instagram/analytics/media?account_id=ACC_ID&limit=10
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Media analytics retrieved",
  "data": {
    "top_media": [
      {
        "id": "media_456",
        "caption": "Beautiful sunset...",
        "likes": 500,
        "comments": 75,
        "engagement_rate": 6.5,
        "reach": 8000,
        "impressions": 10000
      }
    ]
  },
  "request_id": "req_vwx",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Show top performing posts
- Compare metrics across posts
```

#### 11. Get Analytics Trends
```
GET /api/instagram/analytics/trends?account_id=ACC_ID
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Trends retrieved successfully",
  "data": {
    "posting_frequency": "3x per week",
    "best_posting_times": ["Wed 2-3 PM", "Sat 8-9 AM"],
    "top_hashtags": ["#instagram", "#photography", "#nature"],
    "top_content_types": ["Carousel", "Reel"],
    "audience_demographics": {
      "top_locations": ["USA", "UK", "Canada"],
      "age_groups": {"18-24": 35, "25-34": 40, "35+": 25},
      "gender": {"Male": 45, "Female": 55}
    }
  },
  "request_id": "req_yza",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Display in Analytics Trends page
- Show best times to post
- Show audience breakdown
```

#### 12. Generate Caption Suggestions (Pro Only)
```
POST /api/instagram/ai/caption-suggest
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "media_id": "media_456",
  "tone": "professional",  /* professional, casual, fun, inspirational */
  "include_call_to_action": true,
  "max_length": 2200
}

Response 200 OK:
{
  "success": true,
  "message": "Caption suggestions generated",
  "data": {
    "suggestions": [
      {
        "caption": "Experiencing the magic of nature 🌅✨ Every moment counts when you're surrounded by beauty. What's your favorite nature moment?\n\n#sunset #nature #photography #blessed #explore",
        "engagement_score": 8.5
      },
      {
        "caption": "Golden hour magic ✨ Nature never ceases to amaze us. Share your favorite sunset photo in the comments! 👇\n\n#nature #sunset #beautiful #photography #moments",
        "engagement_score": 8.2
      }
    ]
  },
  "request_id": "req_bcd",
  "timestamp": "2026-04-18T10:30:45Z"
}

Feature Gate: Pro Only
Frontend:
- Show in content creation/editing flow
- Display within <ProFeatureGate> component
- Allow user to copy/edit suggestions
```

#### 13. Generate Hashtag Suggestions (Pro Only)
```
POST /api/instagram/ai/hashtag-suggest
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "media_id": "media_456",
  "max_hashtags": 20,
  "focus_area": "general"  /* general, trending, niche */
}

Response 200 OK:
{
  "success": true,
  "message": "Hashtag suggestions generated",
  "data": {
    "hashtags": [
      {"tag": "sunset", "popularity": "high", "volume": 250000},
      {"tag": "nature", "popularity": "high", "volume": 500000},
      {"tag": "photography", "popularity": "medium", "volume": 180000},
      {"tag": "beautifulmoment", "popularity": "low", "volume": 2500}
    ]
  },
  "request_id": "req_efg",
  "timestamp": "2026-04-18T10:30:45Z"
}

Feature Gate: Pro Only
Frontend:
- Suggest hashtags during post creation
- Show popularity/volume metrics
```

#### 14. Get Optimal Posting Strategy (Pro Only)
```
GET /api/instagram/ai/posting-time?account_id=ACC_ID
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Posting strategy generated",
  "data": {
    "best_times": [
      {"day": "Monday", "hour": 18, "value": 8.5},
      {"day": "Wednesday", "hour": 14, "value": 8.8},
      {"day": "Friday", "hour": 19, "value": 9.2}
    ],
    "frequency_recommendation": "3-4 posts per week",
    "content_mix_suggestion": {
      "reels": "40%",
      "carousel": "30%",
      "single_image": "30%"
    },
    "engagement_prediction": "Following this schedule could increase engagement by 35%"
  },
  "request_id": "req_hij",
  "timestamp": "2026-04-18T10:30:45Z"
}

Feature Gate: Pro Only
Frontend:
- Display optimal posting times
- Suggest when to schedule next post
```

#### 15. Manual Media Sync
```
POST /api/instagram/media/sync
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "account_id": "acc_123"
}

Response 202 Accepted:
{
  "success": true,
  "message": "Sync started in background",
  "data": {
    "status": "syncing",
    "account_id": "acc_123"
  },
  "request_id": "req_klm",
  "timestamp": "2026-04-18T10:30:45Z"
}

Sync Flow:
- POST triggers background job
- Frontend shows "Syncing..." spinner
- Backend pulls latest media from Instagram
- Updates InstagramMedia table
- User can see new posts in gallery after sync completes (~30s)

Frontend:
- "Refresh" button in Content Library
- On click: POST sync → Show spinner
- Poll /api/instagram/accounts to check sync_status
- Update gallery when sync_status changes to "idle"
```

#### 16. Manual Media Analysis
```
POST /api/instagram/media/analyze
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "account_id": "acc_123",
  "media_ids": ["media_456", "media_789"]  /* optional, if not provided analyzes all */
}

Response 202 Accepted:
{
  "success": true,
  "message": "Analysis started in background",
  "data": {
    "status": "analyzing",
    "account_id": "acc_123",
    "media_count": 2
  },
  "request_id": "req_nop",
  "timestamp": "2026-04-18T10:30:45Z"
}

Analysis includes:
- AI caption suggestions
- Hashtag recommendations
- Performance predictions
- Optimization tips

Frontend:
- "Analyze All Posts" button
- Shows progress spinner
- After analysis, show suggestions per post
```

#### 17. Webhook Receiver (Backend Only)
```
GET /api/instagram/webhooks (Instagram token validation)
POST /api/instagram/webhooks (Real-time event)

Backend handles:
- Media posted
- Media liked/commented
- Follower changes
- DM received

Frontend:
- No direct interaction
- Backend updates automatically
- Frontend can poll to get latest updates
```

#### 18. Manual Operations Summary
```
For FREE users:
- Connect/disconnect accounts
- View media and analytics
- 5 AI requests per month

For PRO users:
- All FREE features
- Unlimited AI requests
- Optimal posting times
- Caption/hashtag suggestions
- Performance predictions

Feature Gates:
- /media/:id/ai → ProFeatureGate
- /ai/caption-suggest → ProFeatureGate
- /ai/hashtag-suggest → ProFeatureGate
- /ai/posting-time → ProFeatureGate
```

---

### Otto AI Assistant Endpoints (11 endpoints)

#### 1. Create Conversation
```
POST /api/otto/conversations
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "title": "Q2 Marketing Strategy",
  "description": "Discussing campaign strategy for Q2 2026",
  "context": {
    "workspace_id": "ws_123",
    "account_ids": ["acc_123", "acc_456"],
    "date_range": "2026-04-01 to 2026-06-30",
    "focus": "engagement_optimization"
  }
}

Context: Provides AI with background data (accounts, dates, goals)

Response 201 Created:
{
  "success": true,
  "message": "Conversation created successfully",
  "data": {
    "id": "conv_789",
    "workspace_id": "ws_123",
    "title": "Q2 Marketing Strategy",
    "description": "Discussing campaign strategy for Q2 2026",
    "status": "active",
    "is_bookmarked": false,
    "message_count": 0,
    "last_message_at": null,
    "created_at": "2026-04-18T10:30:00Z",
    "updated_at": "2026-04-18T10:30:00Z"
  },
  "request_id": "req_qrs",
  "timestamp": "2026-04-18T10:30:45Z"
}

Feature: Available to all tiers, limited to 5/month for FREE users

Frontend:
- "New Conversation" button
- Show title, description form
- Pre-fill with account context
```

#### 2. List Conversations
```
GET /api/otto/conversations?limit=20&offset=0
Authorization: Bearer {access_token}

Query Parameters:
- limit: 1-50, default 20
- offset: 0-based pagination
- status: active, archived (optional filter)

Response 200 OK:
{
  "success": true,
  "message": "Conversations retrieved successfully",
  "data": {
    "conversations": [
      {
        "id": "conv_789",
        "workspace_id": "ws_123",
        "title": "Q2 Marketing Strategy",
        "description": "Discussing campaign strategy for Q2 2026",
        "status": "active",
        "is_bookmarked": true,
        "message_count": 12,
        "last_message_at": "2026-04-18T09:45:00Z",
        "created_at": "2026-04-18T08:00:00Z"
      }
    ],
    "pagination": {
      "total": 45,
      "limit": 20,
      "offset": 0,
      "has_more": true
    }
  },
  "request_id": "req_tuv",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Display conversation list in sidebar
- Show bookmarked conversations at top
- Use pagination for infinite scroll
```

#### 3. Get Conversation Details
```
GET /api/otto/conversations/:id
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Conversation retrieved successfully",
  "data": {
    "id": "conv_789",
    "workspace_id": "ws_123",
    "title": "Q2 Marketing Strategy",
    "description": "Discussing campaign strategy for Q2 2026",
    "status": "active",
    "is_bookmarked": true,
    "message_count": 12,
    "last_message_at": "2026-04-18T09:45:00Z",
    "created_at": "2026-04-18T08:00:00Z",
    "updated_at": "2026-04-18T09:45:00Z"
  },
  "request_id": "req_wxy",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Show when opening conversation details
```

#### 4. Update Conversation
```
PUT /api/otto/conversations/:id
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "title": "Q2 & Q3 Marketing Strategy",
  "description": "Updated strategy for Q2 and Q3",
  "is_bookmarked": true
}

Response 200 OK:
{
  "success": true,
  "message": "Conversation updated successfully",
  "data": {
    "id": "conv_789",
    "title": "Q2 & Q3 Marketing Strategy",
    "description": "Updated strategy for Q2 and Q3",
    "is_bookmarked": true,
    "updated_at": "2026-04-18T10:31:00Z"
  },
  "request_id": "req_zab",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Edit conversation title/description
- Click bookmark star to toggle is_bookmarked
```

#### 5. Archive Conversation
```
POST /api/otto/conversations/:id/archive
Authorization: Bearer {access_token}
Content-Type: application/json

Request: {} (empty body)

Response 200 OK:
{
  "success": true,
  "message": "Conversation archived successfully",
  "data": {
    "id": "conv_789",
    "status": "archived"
  },
  "request_id": "req_cde",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- "Archive" button in conversation menu
- Removes from active list
- Can restore with unarchive endpoint (optional)
```

#### 6. Bookmark Conversation
```
POST /api/otto/conversations/:id/bookmark
Authorization: Bearer {access_token}
Content-Type: application/json

Request: {} (empty body)

Response 200 OK:
{
  "success": true,
  "message": "Conversation bookmarked successfully",
  "data": {
    "id": "conv_789",
    "is_bookmarked": true
  },
  "request_id": "req_fgh",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Star icon to toggle bookmark
- Show bookmarked conversations first
```

#### 7. Delete Conversation
```
DELETE /api/otto/conversations/:id
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Conversation deleted successfully",
  "data": {
    "id": "conv_789",
    "status": "deleted"
  },
  "request_id": "req_ijk",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- "Delete" button with confirmation dialog
- "Are you sure? This cannot be undone."
```

#### 8. Send Message (User to Otto)
```
POST /api/otto/conversations/:id/messages
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "content": "What are the top 3 hashtags I should focus on for Q2?"
}

Max length: 5000 characters

Response 200 OK:
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "messages": [
      {
        "id": "msg_123",
        "conversation_id": "conv_789",
        "role": "user",
        "content": "What are the top 3 hashtags I should focus on for Q2?",
        "tokens_used": 15,
        "model_used": "user_input",
        "created_at": "2026-04-18T10:31:00Z"
      },
      {
        "id": "msg_124",
        "conversation_id": "conv_789",
        "role": "assistant",
        "content": "Based on your Q2 performance data, I recommend focusing on: 1) #marketing (high volume, good engagement), 2) #springcampaign (trending), 3) #engagement2026 (growing trend). Here's why...",
        "tokens_used": 156,
        "model_used": "claude-opus-4-6",
        "is_liked": null,
        "created_at": "2026-04-18T10:31:02Z"
      }
    ]
  },
  "request_id": "req_lmn",
  "timestamp": "2026-04-18T10:30:45Z"
}

Flow:
1. User types message → POST
2. Backend sends to Claude API
3. Claude responds
4. Response saved as OttoMessage with role="assistant"
5. Both messages returned in response
6. Frontend renders both immediately

Frontend:
- Chat input at bottom
- Messages stream in conversation thread
- Show "Otto is thinking..." while awaiting response
- AI request counter (FREE: 5/month consumed)
```

#### 9. Get Messages in Conversation
```
GET /api/otto/conversations/:id/messages?limit=50&offset=0
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Messages retrieved successfully",
  "data": {
    "messages": [
      {
        "id": "msg_123",
        "conversation_id": "conv_789",
        "role": "user",
        "content": "What insights can you give about my audience?",
        "created_at": "2026-04-18T10:30:00Z"
      },
      {
        "id": "msg_124",
        "conversation_id": "conv_789",
        "role": "assistant",
        "content": "Based on your analytics...",
        "created_at": "2026-04-18T10:30:02Z"
      }
    ],
    "pagination": {
      "total": 24,
      "limit": 50
    }
  },
  "request_id": "req_opq",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Load message history when opening conversation
- Show oldest messages first, newest at bottom
```

#### 10. Add Message Feedback
```
POST /api/otto/messages/:id/feedback
Authorization: Bearer {access_token}
Content-Type: application/json

Request:
{
  "is_liked": true,
  "feedback_notes": "Very helpful suggestion, exactly what I needed!"
}

Response 200 OK:
{
  "success": true,
  "message": "Feedback recorded successfully",
  "data": {
    "message_id": "msg_124",
    "is_liked": true,
    "feedback_notes": "Very helpful suggestion..."
  },
  "request_id": "req_rst",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Thumbs up/down buttons below each assistant message
- Optional note box: "What could be better?"
- Send feedback to improve AI responses
```

#### 11. Get Conversation Context
```
GET /api/otto/conversations/:id/context
Authorization: Bearer {access_token}

Response 200 OK:
{
  "success": true,
  "message": "Context retrieved successfully",
  "data": {
    "conversation_id": "conv_789",
    "context": {
      "workspace_id": "ws_123",
      "account_ids": ["acc_123", "acc_456"],
      "date_range": "2026-04-01 to 2026-06-30",
      "focus": "engagement_optimization",
      "recent_metrics": {
        "total_posts": 45,
        "avg_engagement": 4.2,
        "follower_growth": 2.5
      }
    }
  },
  "request_id": "req_uvw",
  "timestamp": "2026-04-18T10:30:45Z"
}

Frontend:
- Show context used by Otto for responses
- Display in conversation sidebar
- Helps user understand AI's perspective
```

---

### Otto AI Request Counter

```typescript
// Track FREE tier limit
const AI_REQUEST_LIMITS = {
  FREE: 5,     // per month
  PRO: Infinity // unlimited
};

// Implementation
export function useOttoAIRequests() {
  const { data: subscription } = useSubscription();
  const { data: usage } = useQuery({
    queryKey: ['otto', 'usage', currentMonth()],
    queryFn: async () => {
      const response = await api.get('/otto/messages/usage');
      return response.data.data;
    },
  });

  return {
    used: usage?.count || 0,
    limit: subscription?.tier === 'pro' ? Infinity : 5,
    remaining: Math.max(0, (subscription?.tier === 'pro' ? Infinity : 5) - (usage?.count || 0)),
  };
}

Frontend Chat Component:
- Show "X requests remaining" for FREE users
- Disable input when limit reached
- Show "Upgrade to Pro" button
```

---

### Standard Error Codes & HTTP Status

```typescript
// Complete error reference
const ERROR_MAPPING = {
  // 400 - Bad Request (Client Error)
  '400:VALIDATION_ERROR': {
    message: 'Please check the highlighted fields',
    actionable: true,
    showFields: true,
  },
  '400:INVALID_EMAIL': {
    message: 'Please enter a valid email address',
    actionable: true,
  },
  '400:WEAK_PASSWORD': {
    message: 'Password must be at least 8 characters with uppercase, lowercase, and number',
    actionable: true,
  },
  '400:INVALID_REQUEST': {
    message: 'Invalid request format. Please try again.',
    actionable: false,
  },

  // 401 - Unauthorized (Auth Failed)
  '401:UNAUTHORIZED': {
    message: 'Please log in to continue',
    action: 'redirect_to_login',
    actionable: true,
  },
  '401:INVALID_CREDENTIALS': {
    message: 'Email or password is incorrect',
    actionable: true,
  },
  '401:TOKEN_EXPIRED': {
    message: 'Your session expired. Please log in again.',
    action: 'redirect_to_login',
    actionable: true,
  },
  '401:OTP_EXPIRED': {
    message: 'OTP expired. Please request a new one.',
    actionable: true,
  },

  // 403 - Forbidden (Permission Denied)
  '403:FORBIDDEN': {
    message: "You don't have permission to access this resource",
    actionable: false,
  },

  // 404 - Not Found
  '404:NOT_FOUND': {
    message: 'The resource you requested does not exist',
    actionable: false,
  },

  // 409 - Conflict (Resource Conflict)
  '409:DUPLICATE_EMAIL': {
    message: 'This email is already registered',
    suggestion: 'Try logging in or use a different email',
    actionable: true,
  },
  '409:DUPLICATE_USERNAME': {
    message: 'This username is already taken',
    suggestion: 'Please choose a different username',
    actionable: true,
  },
  '409:ALREADY_VERIFIED': {
    message: 'This account is already verified',
    actionable: false,
  },

  // 429 - Too Many Requests (Rate Limited)
  '429:RATE_LIMIT_EXCEEDED': {
    message: 'Too many requests. Please try again in {retry_after} seconds',
    retryable: true,
    actionable: false,
  },
  '429:ACCOUNT_LOCKED': {
    message: 'Too many failed login attempts. Please try again in 15 minutes',
    retryable: true,
    actionable: false,
  },

  // 500 - Server Error
  '500:INTERNAL_ERROR': {
    message: 'Something went wrong on our end. Please try again later.',
    actionable: false,
    reportable: true,
  },
};

// Frontend Toast Logic
function showError(error: AppError) {
  const mapping = ERROR_MAPPING[`${error.statusCode}:${error.code}`];
  
  if (mapping?.showFields && error.fieldErrors) {
    // Show inline field errors
    Object.entries(error.fieldErrors).forEach(([field, message]) => {
      showFieldError(field, message);
    });
  } else if (mapping?.retryable) {
    // Show toast with retry button
    showToast({
      type: 'warning',
      message: mapping.message,
      action: { label: 'Retry', onClick: retryFunction },
    });
  } else if (mapping?.action === 'redirect_to_login') {
    logout();
  } else {
    // Default error toast
    showToast({
      type: 'error',
      message: mapping?.message || error.message,
    });
  }
  
  if (mapping?.reportable) {
    reportToSentry(error);
  }
}
```

### Validation Rules by Endpoint

```typescript
// Frontend validation (before sending to backend)
const VALIDATION_RULES = {
  email: {
    required: true,
    pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
    maxLength: 255,
    message: 'Please enter a valid email address',
  },
  
  password: {
    required: true,
    minLength: 8,
    pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
    message: 'Password must contain uppercase, lowercase, number, and special character',
  },
  
  username: {
    required: true,
    minLength: 3,
    maxLength: 30,
    pattern: /^[a-zA-Z0-9_]+$/,
    message: 'Username can only contain letters, numbers, and underscores',
  },
  
  firstName: {
    required: true,
    minLength: 2,
    maxLength: 50,
    message: 'First name must be 2-50 characters',
  },
  
  lastName: {
    required: true,
    minLength: 2,
    maxLength: 50,
    message: 'Last name must be 2-50 characters',
  },
  
  timezone: {
    required: true,
    type: 'select',
    options: moment.tz.names(),
    message: 'Please select a valid timezone',
  },
  
  language: {
    required: true,
    type: 'select',
    options: ['en', 'es', 'fr', 'de', 'pt', 'ja', 'zh'],
    message: 'Please select a language',
  },
};

// Use librarylike Zod or Yup for this
import { z } from 'zod';

const LoginSchema = z.object({
  email: z.string().email('Invalid email format'),
  password: z.string().min(1, 'Password is required'),
});

const RegisterSchema = z.object({
  first_name: z.string().min(2).max(50),
  last_name: z.string().min(2).max(50),
  username: z.string().min(3).max(30).regex(/^[a-zA-Z0-9_]+$/),
  email: z.string().email(),
  password: z
    .string()
    .min(8)
    .regex(/[A-Z]/, 'Must contain uppercase')
    .regex(/[a-z]/, 'Must contain lowercase')
    .regex(/[\d]/, 'Must contain number'),
});
```

---

## Rate Limiting Strategies

### Understanding Rate Limits

```
Global Rate Limit: 100 requests per minute (per IP)
Endpoint Limits:
├─ Auth endpoints: 10 req/min (prevent brute force)
├─ User endpoints: 30 req/min
├─ Workspace endpoints: 30 req/min
└─ Subscription endpoints: 10 req/min

Account Lockout: 5 failed login attempts → 15 min lock
```

### Frontend Handling

```typescript
// hooks/useRateLimit.ts
import { useCallback, useState } from 'react';

export function useRateLimit() {
  const [isRateLimited, setIsRateLimited] = useState(false);
  const [retryAfter, setRetryAfter] = useState<number | null>(null);

  const handleRateLimitError = useCallback((response: AxiosError) => {
    if (response.status === 429) {
      const retryAfterHeader = response.headers['retry-after'];
      const retrySeconds = parseInt(retryAfterHeader) || 60;

      setIsRateLimited(true);
      setRetryAfter(retrySeconds);

      // Auto-enable after delay
      setTimeout(() => {
        setIsRateLimited(false);
        setRetryAfter(null);
      }, retrySeconds * 1000);

      return {
        shouldRetry: true,
        delayMs: retrySeconds * 1000,
      };
    }

    return { shouldRetry: false };
  }, []);

  return { isRateLimited, retryAfter, handleRateLimitError };
}

// Usage in component
export function LoginForm() {
  const { isRateLimited, retryAfter } = useRateLimit();
  const [isSubmitting, setIsSubmitting] = useState(false);

  return (
    <form>
      <button 
        type="submit"
        disabled={isRateLimited || isSubmitting}
      >
        {isRateLimited ? `Try again in ${retryAfter}s` : 'Sign In'}
      </button>
      {isRateLimited && (
        <p className="text-warning">
          Too many attempts. Please wait before trying again.
        </p>
      )}
    </form>
  );
}
```

### Exponential Backoff for Retries

```typescript
// lib/retryConfig.ts
export const retryConfig = {
  maxRetries: 3,
  getDelayMs: (attempt: number) => {
    // 1s, 2s, 4s, 8s...
    return Math.min(1000 * Math.pow(2, attempt - 1), 10000);
  },
  shouldRetry: (status: number) => {
    // Retry on: timeout, rate limit, server error
    return status === 408 || status === 429 || status >= 500;
  },
};

// Used in API interceptor (shown earlier)
```

---

## Subscription & Payment Flow

### Complete Subscription Journey (Freemium Model)

```
State 1: FREE User (Default on Signup)
├─ subscription_tier: null
├─ subscription_status: "free"
├─ Features: Limited (1 workspace, 5 AI requests/month, manual posting)
├─ UI: Shows "Upgrade to Pro" prompts throughout dashboard
└─ CTA: "Upgrade to Pro" button in settings & dismissable banner

    ↓ User clicks "Upgrade to Pro" (from dashboard/settings) ↓

State 2: During Checkout
├─ POST /api/subscriptions/checkout
├─ Redirect to Paddle external checkout
├─ User enters payment details
└─ Confirms purchase ($XX/year)

    ↓ Payment processed by Paddle ↓

State 3: Paddle Webhook → Backend
├─ Backend receives subscription.created webhook
├─ Creates subscription record
├─ Sets subscription_status: "active"
├─ Sets subscription_tier: "pro"
└─ Expires 1 year from now

    ↓ Webhook triggers frontend refresh ↓

State 4: PRO User (Active Subscription)
├─ subscription_tier: "pro"
├─ subscription_status: "active"
├─ expires_at: "2027-04-18T..."
├─ Features: All unlocked (unlimited workspaces, AI, scheduled posting, team invite)
├─ UI: Pro badge, "Manage Subscription" button
└─ CTA: "Manage Subscription" → Customer Portal

Timeline:
- Immediate: Pro features unlock on dashboard
- 30 days before expiry: Email reminder
- On expiry: Auto-retry payment
- If failed: Mark as "expired", show renewal CTA
- If cancelled: subscription_status: "cancelled", features downgrade to free
```

### Key Difference: Freemium vs Paywall

```
FREEMIUM MODEL (✅ Your App)
├─ Users can explore FREE tier fully
├─ Pro features are locked behind gates
├─ Upgrade is optional, available anytime
├─ Better conversion: Users experience value first
├─ Dashboard: "Upgrade" CTAs + banners
└─ Result: Most users upgrade after trying

PAYWALL MODEL (❌ Not this app)
├─ Users must subscribe before accessing
├─ All features require payment
├─ Upgrade is required to proceed
└─ Result: Higher churn, users leave before trying
```

### Frontend Subscription Implementation

```typescript
// hooks/useSubscription.ts
import { useQuery, useMutation } from '@tanstack/react-query';
import api from '@/lib/api';

export function useSubscription() {
  return useQuery({
    queryKey: ['subscription', 'status'],
    queryFn: async () => {
      const response = await api.get('/subscriptions/status');
      return response.data.data;
    },
    refetchInterval: 1000 * 60 * 60, // Refetch every hour
  });
}

export function useCreateCheckout() {
  return useMutation({
    mutationFn: async () => {
      const response = await api.post('/subscriptions/checkout', {});
      return response.data.data.checkout_url;
    },
    onSuccess: (checkoutUrl) => {
      // Open checkout in new tab/window
      window.location.href = checkoutUrl;
    },
  });
}

export function useGetPortal() {
  return useMutation({
    mutationFn: async () => {
      const response = await api.get('/subscriptions/portal');
      return response.data.data.portal_url;
    },
    onSuccess: (portalUrl) => {
      window.open(portalUrl, '_blank');
    },
  });
}

// Component usage
export function UpgradeButton() {
  const { data: subscription } = useSubscription();
  const createCheckout = useCreateCheckout();

  if (subscription?.status === 'active') {
    return (
      <button onClick={() => useGetPortal().mutate()}>
        Manage Subscription
      </button>
    );
  }

  return (
    <button 
      onClick={() => createCheckout.mutate()}
      disabled={createCheckout.isPending}
    >
      {createCheckout.isPending ? 'Loading...' : 'Upgrade to Pro'}
    </button>
  );
}
```

### Handling Subscription State in UI

```typescript
// utils/subscriptionUtils.ts
export function canAccessProFeature(subscription: Subscription): boolean {
  return (
    subscription.tier === 'pro' && 
    subscription.status === 'active'
  );
}

export function getSubscriptionBadge(subscription: Subscription): string {
  if (subscription.status === 'active') {
    return 'Pro';
  }
  if (subscription.status === 'expired') {
    return 'Expired - Renew';
  }
  if (subscription.status === 'cancelled') {
    return 'Cancelled';
  }
  return 'Free';
}

export function getDaysUntilExpiry(expiresAt: string): number {
  return Math.ceil(
    (new Date(expiresAt).getTime() - new Date().getTime()) / 
    (1000 * 60 * 60 * 24)
  );
}

// Component: Pro-Only Feature
export function ProFeatureCard({ children }) {
  const { data: subscription } = useSubscription();
  const canAccess = canAccessProFeature(subscription);

  if (!canAccess) {
    return (
      <div className="pro-feature-locked">
        <p>This feature is Pro only</p>
        <button className="upgrade-btn">
          Upgrade to Pro
        </button>
      </div>
    );
  }

  return <div>{children}</div>;
}

// Component: Expiry Warning
export function SubscriptionStatus() {
  const { data: subscription } = useSubscription();

  if (!subscription?.expires_at) return null;

  const daysLeft = getDaysUntilExpiry(subscription.expires_at);

  if (daysLeft < 7) {
    return (
      <Alert type="warning">
        Your Pro subscription expires in {daysLeft} days.{' '}
        <a href="/settings/billing">Manage subscription</a>
      </Alert>
    );
  }

  return null;
}
```

---

## Environment Configuration

### Next.js Environment Setup

```bash
# .env.local (Development)
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Refyne
NEXT_PUBLIC_SENTRY_DSN=https://your-sentry-key
NEXT_PUBLIC_POSTHOG_KEY=your-posthog-key

# .env.staging (Staging)
NEXT_PUBLIC_API_URL=https://api-staging.refyne.app
NEXT_PUBLIC_APP_NAME=Refyne (Staging)
NEXT_PUBLIC_ANALYTICS_ID=staging_key

# .env.production (Production - Vercel)
NEXT_PUBLIC_API_URL=https://api.refyne.app
NEXT_PUBLIC_APP_NAME=Refyne
NEXT_PUBLIC_SENTRY_DSN=https://production-sentry-key
NEXT_PUBLIC_POSTHOG_KEY=production-posthog-key
```

**Important Notes:**
- Only variables prefixed with `NEXT_PUBLIC_` are exposed to browser
- Never put sensitive keys, JWT secrets, or API keys in `.env.local`
- Refresh tokens are handled via httpOnly cookies (server-side, not in env)
- Database credentials never go to frontend

### Vercel Deployment Configuration

```
1. Connect GitHub repository to Vercel
2. Set environment variables in Vercel dashboard:
   - NEXT_PUBLIC_API_URL=https://api.refyne.app
   - Other NEXT_PUBLIC_* variables
3. Set Node.js version: 18.x or higher
4. Build command: next build
5. Output directory: .next
6. Install command: npm ci (or yarn install)

Environment Behavior:
- development: Uses .env.local
- preview: Uses .env.production (for branch previews)
- production: Uses .env.production (for main branch)
```

### CORS Configuration

**Backend (Railway):**
```go
// Already configured in Refyne backend
cors.AllowedOrigins: [
  "http://localhost:3000",        // Dev
  "https://*.vercel.app",          // Vercel preview deployments
  "https://refyne.app",            // Production
  "https://www.refyne.app",
]
cors.AllowedMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
cors.AllowCredentials: true        // For httpOnly cookies
cors.AllowedHeaders: ["Authorization", "Content-Type"]
```

**Frontend (Next.js):**
```typescript
// Automatic with next/image, fetch API handles CORS
// Credentials automatically sent for same-origin requests
```

---

## Complete Project Checklist

### Pre-Development Checklist

- [ ] Repository created and connected to Vercel
- [ ] Env variables configured in Vercel dashboard
- [ ] TypeScript configured in `tsconfig.json`
- [ ] TailwindCSS installed (or your CSS framework)
- [ ] Zustand store structure created
- [ ] React Query configured with QueryClientProvider
- [ ] API client (axios instance) created with interceptors
- [ ] Error boundary component created
- [ ] Loading states component created
- [ ] Toast/notification system set up

### Development Phases

**Phase 1: Auth (Week 1)**
- [ ] Login page
- [ ] Registration page
- [ ] OTP verification flow
- [ ] Password reset flow
- [ ] Token management & auto-refresh
- [ ] Protected routes middleware
- [ ] Logout functionality

**Phase 2: Onboarding (Week 2)**
- [ ] Profile setup step
- [ ] Preferences setup step
- [ ] Workspace creation step
- [ ] Subscription upgrade prompt
- [ ] Onboarding completion
- [ ] Progress indicator
- [ ] Persist progress (local storage backup)

**Phase 3: Dashboard (Week 3)**
- [ ] Main dashboard layout
- [ ] Workspace switcher
- [ ] Sidebar navigation
- [ ] User profile dropdown
- [ ] Settings page skeleton
- [ ] Theme provider (light/dark)

**Phase 4: User Settings (Week 3-4)**
- [ ] Profile settings (edit name, username)
- [ ] Preferences settings (language, timezone)
- [ ] Email notification preferences
- [ ] Account deletion with warning
- [ ] Password management link

**Phase 5: Workspace Management (Week 4)**
- [ ] Workspace list page
- [ ] Create workspace flow
- [ ] Workspace settings
- [ ] Members page
- [ ] Invite member flow
- [ ] Member removal with confirmation

**Phase 6: Subscription & Billing (Week 5)**
- [ ] Subscription status display
- [ ] Upgrade to Pro flow
- [ ] Paddle checkout redirect
- [ ] Billing dashboard link
- [ ] Subscription status badge

### Testing Checklist

- [ ] All auth flows tested (register, login, reset password)
- [ ] Token refresh works automatically
- [ ] Rate limiting shows user-friendly messages
- [ ] Network errors handled gracefully
- [ ] Form validation matches backend rules
- [ ] Onboarding progress persists
- [ ] Subscription status updates after purchase
- [ ] Logout clears all data properly
- [ ] Dark mode toggle works
- [ ] Mobile responsive on all pages

### Deploy Checklist

- [ ] All environment variables set in Vercel
- [ ] API URL points to production backend
- [ ] Analytics configured (Sentry, PostHog, etc.)
- [ ] CDN/caching configured
- [ ] SSL certificate valid
- [ ] Security headers set (CSP, X-Frame-Options, etc.)
- [ ] Performance monitoring enabled
- [ ] Error reporting to Sentry enabled
- [ ] Backup & disaster recovery plan

---

## Critical Backend Integration Points

### 1. Token Lifecycle

```
Access Token (15 min):
├─ Generated: After OTP verification
├─ Stored: Zustand (memory)
├─ Sent: Authorization: Bearer {token}
└─ Expires: Automatic refresh on 401

Refresh Token (7 days):
├─ Generated: After OTP verification
├─ Stored: httpOnly cookie (automatic)
├─ Sent: Automatic with requests (CORS credentialed)
└─ Expires: Auto-redirect to login
```

### 2. Subscription Flow Integration

```
Frontend                           Backend
     │                                │
     └─ POST /subscriptions/checkout ─┤
                                      ├─ Generate checkout URL
     ┌─ Response with URL ←──────────┤
     │
     ├─ Redirect to Paddle ─┐
     │                      │
     │ User pays ←── Paddle checkout
     │                      │
     └─ Return to /dashboard
                            ├─ Paddle webhook
                            ├─ subscription.created
                            ├─ Create subscription record
                            ├─ Update user tier
                            └─ Backend ready for next API call
     
     GET /subscriptions/status ──┤
                                 ├─ Return new subscription
     ← Receive updated status ←──┤
     
     Update UI with Pro features
```

### 3. Error Recovery Examples

```typescript
// Scenario 1: Rate Limited
Attempt: POST /api/auth/request-otp
Response: 429 Too Many Requests
Retry-After: 60

Client: Wait 60 seconds, show countdown to user
Then: Auto-retry or user clicks "Try Again"

---

// Scenario 2: Token Expired
Attempt: GET /api/user/profile
Response: 401 Unauthorized

Client: 
1. Detect 401
2. Call POST /api/auth/refresh (has refresh token in cookie)
3. Get new access token
4. Retry original request
5. User doesn't notice anything

---

// Scenario 3: Network Error
Attempt: POST /api/workspaces (create workspace)
Response: Network timeout

Client:
1. Show: "Connection error. Retrying..."
2. Wait 1 second, retry
3. If fails again: Show "Please check connection"
4. Wait 2 seconds, retry
5. If fails again: Show "Try again manually"
6. Stop auto-retry

---

// Scenario 4: Validation Error
Attempt: POST /api/auth/register
Response: 400 VALIDATION_ERROR
Details: { email: "Invalid email format", password: "... weak ..." }

Client:
1. Show inline error on email field
2. Show inline error on password field
3. Prevent form submission until fixed
4. User fixes and resubmits
```

---

## Performance Optimization Tips

### Query Caching Strategy

```typescript
// Queries that should be cached long-term
const LONG_CACHE = 1000 * 60 * 60 * 24; // 24 hours
// - User profile (rarely changes)
// - Workspaces list (queries)

// Queries that should be fresh
const SHORT_CACHE = 1000 * 60; // 1 minute
// - Subscription status (changes externally)
// - Members list (can change by someone else inviting)

// Queries that should always be fresh
const NO_CACHE = 0;
// - Current auth status
// - Real-time features

useQuery({
  queryKey: ['user', 'profile'],
  queryFn: fetchUserProfile,
  staleTime: LONG_CACHE,
  gcTime: LONG_CACHE * 2,
});
```

### Image Optimization

```typescript
// Use Next.js Image component for optimization
import Image from 'next/image';

<Image
  src={user.avatar_url}
  alt={user.first_name}
  width={40}
  height={40}
  className="rounded-full"
  priority={true} // For above-the-fold
/>

// Or use external image service
<img 
  src={`https://images.refyne.app/resize/40x40/${user.id}`}
  alt="User avatar"
  loading="lazy"
/>
```

### Bundle Size Optimization

```typescript
// Use dynamic imports for heavy components
import dynamic from 'next/dynamic';

const AdvancedAnalytics = dynamic(
  () => import('@/components/Analytics'),
  { loading: () => <Skeleton /> }
);

// Tree-shake unused code
// Only import what you need from libraries
import { QueryClient } from '@tanstack/react-query'; // Good
// import * from '@tanstack/react-query'; // Bad - includes everything
```

---

## Security Best Practices

### CORS & CSRF

```typescript
// Frontend: CORS is automatic with fetch + withCredentials: true
// Backend: Already configured to allow your frontend domain

// CSRF Protection: Not needed for modern APIs with SameSite cookies
// Backend sets: Set-Cookie: refresh_token=...; SameSite=Strict; HttpOnly
```

### XSS Prevention

```typescript
// Safe: React auto-escapes by default
<p>{user.bio}</p> // Safe - escapes HTML

// UNSAFE: DOMPurify required if rendering HTML
<div dangerouslySetInnerHTML={{ __html: user.bio }} /> // Dangerous!
import DOMPurify from 'dompurify';
<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(user.bio) }} /> // Safe
```

### SQL Injection / NoSQLInjection

```typescript
// ✅ Safe: Backend uses parameterized queries
const user = await db.query('SELECT * FROM users WHERE email = $1', [email]);

// ❌ Unsafe (backend side)
const user = await db.query(`SELECT * FROM users WHERE email = '${email}'`); // Don't do this!

// Frontend doesn't make raw queries, so this is not your concern
// But be aware when reviewing backend code
```

### Secrets Management

```typescript
// ✅ Do: Use environment variables for public URLs only
NEXT_PUBLIC_API_URL=https://api.refyne.app

// ✅ Do: httpOnly cookies for sensitive tokens
Set-Cookie: refresh_token=...; HttpOnly; SameSite=Strict

// ❌ Don't: Store secrets in localStorage
localStorage.setItem('api_key', 'secret_key'); // Never do this

// ❌ Don't: Commit .env.local
# .gitignore
.env.local
.env.*.local
```

---

## Monitoring & Analytics

### Recommended Tools

```
Error Tracking: Sentry
- Automatic error capture
- 401, 429, 5xx endpoints tracked
- Source maps for production debugging

Analytics: PostHog
- User behavior tracking
- Funnel analysis (signup → onboarding → upgrade)
- Feature usage

Performance: Vercel Analytics
- Web Vitals (LCP, FID, CLS)
- Real User Monitoring
- Deployment performance

Health Checks: Uptime Robot
- Monitor /health endpoint
- 5-minute intervals
- Alert if down
```

### Key Metrics to Track

```typescript
// Authentication funnel
- Registration attempts
- Registration completions
- OTP request rate
- Login success rate
- Login failure rate (by reason)

// Onboarding funnel
- Onboarding started
- Profile completed
- Preferences completed
- Workspace created
- Subscription upgrade attempts
- Onboarding completed

// Subscription metrics
- Upgrade click rate
- Checkout success rate
- Failed checkout reasons
- Churn rate
- Feature usage by tier
```

---

End of Frontend Implementation Guide

For questions about specific endpoints, refer back to Section 5: "Complete Endpoint Reference"

Last Updated: 2026-04-18  
Ready for Development: ✅
